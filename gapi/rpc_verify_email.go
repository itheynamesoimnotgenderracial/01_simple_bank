package gapi

import (
	"context"

	db "github.com/projects/go/01_simple_bank/db/sqlc"
	"github.com/projects/go/01_simple_bank/pb"
	"github.com/projects/go/01_simple_bank/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(req)

	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email: %v", err)
	}

	rsp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}

	return rsp, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateEmailId(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolations("email_id", err))
	}

	if err := validation.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolations("secret_code", err))
	}

	return violations
}
