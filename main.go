package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projects/go/01_simple_bank/api"
	db "github.com/projects/go/01_simple_bank/db/sqlc"
	_ "github.com/projects/go/01_simple_bank/docs/statik"
	"github.com/projects/go/01_simple_bank/email"
	"github.com/projects/go/01_simple_bank/gapi"
	"github.com/projects/go/01_simple_bank/pb"
	"github.com/projects/go/01_simple_bank/util"
	"github.com/projects/go/01_simple_bank/worker"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interrupSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Error().Msgf("cannot load configuration: %v", err)
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interrupSignals...)
	defer stop()

	conn, err := pgxpool.New(ctx, config.DBSource)

	if err != nil {
		log.Error().Msgf("cannot connect to db: %v", err)
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(config, ctx, waitGroup, redisOpt, store)
	runGatewayServer(config, ctx, waitGroup, store, taskDistributor)
	runGrpcServer(config, ctx, waitGroup, store, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Error().Msgf("cannot create new migrate instance: %v", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error().Msgf("failed to migrate up: %v", err)
	}

	log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(
	config util.Config,
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt,
	store db.Store) {
	mailer := email.NewGmailSend(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.
		Info().
		Msg("start task processor")

	err := taskProcessor.Start()

	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("failed to task task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		taskProcessor.Shutdown()
		log.Info().Msg("task processor is stopped")

		return nil
	})
}

func runGrpcServer(
	config util.Config,
	ctx context.Context,
	waitGroup *errgroup.Group,
	store db.Store,
	taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)

	if err != nil {
		log.Error().Msgf("token create server error: %v", err)
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Error().Msgf("cannot create gRPC listener: %v", err)
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
		err = grpcServer.Serve(listener)

		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Error().Msgf("gRPC server failed to serve: %v", err)
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown gRPC server")

		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server is stopped")

		return nil
	})
}

func runGatewayServer(
	config util.Config,
	ctx context.Context,
	waitGroup *errgroup.Group,
	store db.Store,
	taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)

	if err != nil {
		log.Error().Msgf("token create server error: %v", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Error().Msgf("cannot register handler server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()

	if err != nil {
		log.Error().Msgf("cannot statik file system: %v", err)
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	httpServer := &http.Server{
		Handler: gapi.HttpLogger(mux),
		Addr:    config.HTTPServerAddress,
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe()

		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Msgf("HTTP gateway server failed to serve: %v", err)
			return err
		}

		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown HTTP gateway server")

		err := httpServer.Shutdown(context.Background())

		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP gateway server")

		}

		log.Info().Msg("HTTP gateway server is stoppped")
		return nil
	})
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)

	if err != nil {
		log.Error().Msgf("token create server error: %v", err)
	}

	err = server.Start(config.HTTPServerAddress)

	if err != nil {
		log.Error().Msgf("cannot start server: %v", err)
	}

}
