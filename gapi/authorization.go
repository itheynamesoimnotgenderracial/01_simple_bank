package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/projects/go/01_simple_bank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) authorizedUser(ctx context.Context) (*token.Payload, error) {
	mtdt, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, fmt.Errorf("missing metadata: %v", ok)
	}

	values := mtdt.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header: %v", values)
	}

	authHeader := values[0]

	// Bearer ....
	fields := strings.Fields(authHeader)

	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format: %v", fields)
	}

	authType := strings.ToLower(fields[0])

	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type: %v", values)
	}

	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)

	if err != nil {
		return nil, fmt.Errorf("invalid access token: %v", err)
	}

	return payload, nil
}
