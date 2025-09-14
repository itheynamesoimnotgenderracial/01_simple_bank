package gapi

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/projects/go/01_simple_bank/db/sqlc"
	"github.com/projects/go/01_simple_bank/pb"
	"github.com/projects/go/01_simple_bank/token"
	"github.com/projects/go/01_simple_bank/util"
	"github.com/projects/go/01_simple_bank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	router          *gin.Engine
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

// NewServer creates a new HTTP gRPC
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %v", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
