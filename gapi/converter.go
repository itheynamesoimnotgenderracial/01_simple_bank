package gapi

import (
	db "github.com/projects/go/01_simple_bank/db/sqlc"
	"github.com/projects/go/01_simple_bank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:         user.Username,
		Fullname:         user.FullName,
		Email:            user.Email,
		PasswordChangeAt: timestamppb.New(user.PasswordChangedAt.Time),
		CreatedAt:        timestamppb.New(user.CreatedAt.Time),
	}
}
