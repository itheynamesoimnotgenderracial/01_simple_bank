package gapi

import (
	"context"

	db "github.com/projects/go/01_simple_bank/db/sqlc"
	"github.com/projects/go/01_simple_bank/pb"
	"github.com/projects/go/01_simple_bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	hashedPassword, err := util.HashPassword(req.GetPassword())

	// intentional error
	// hashedPassword, err := util.HashPassword("xyz")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullname(),
		Email:          req.GetEmail(),
	}

	// arg = db.CreateUserParams{}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		errorCode := db.ErrorCode(err)
		if errorCode == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(user),
	}

	return rsp, nil
}
