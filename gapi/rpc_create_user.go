package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	db "github.com/projects/go/01_simple_bank/db/sqlc"
	"github.com/projects/go/01_simple_bank/pb"
	"github.com/projects/go/01_simple_bank/util"
	"github.com/projects/go/01_simple_bank/validation"
	"github.com/projects/go/01_simple_bank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)

	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())

	// intentional error
	// hashedPassword, err := util.HashPassword("xyz")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashedPassword,
			FullName:       req.GetFullname(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			// Todo: user db transaction to check if transaction is successful or not to avoid duplicate DB request

			// Send verification email
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	// arg = db.CreateUserParams{}

	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		errorCode := db.ErrorCode(err)
		if errorCode == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(txResult.User),
	}

	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolations("username", err))
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolations("password", err))
	}

	if err := validation.ValidateFullname(req.GetFullname()); err != nil {
		violations = append(violations, fieldViolations("full_name", err))
	}

	if err := validation.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolations("email", err))
	}

	return violations
}
