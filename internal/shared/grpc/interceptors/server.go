package interceptors

import (
	"context"
	"crypto"
	"fmt"
	"strings"
	"time"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ServerInterceptor(validBaoToken string, publicKey crypto.PublicKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata missing")
		}

		// 1. ПРОВЕРКА SERVICE TOKEN (OPENBAO) — ВСЕГДА
		tokens := md.Get("x-service-token")
		if len(tokens) == 0 || tokens[0] != validBaoToken {
			return nil, status.Error(codes.PermissionDenied, "invalid service token")
		}
		fmt.Println(info.FullMethod)
		// 2. ПРОВЕРКА JWT — ТОЛЬКО ДЛЯ USER SERVICE
		// info.FullMethod выглядит как "/auth.UserService/UpdateUser"
		if strings.Contains(info.FullMethod, "UserService") {
			authHeader := md.Get("authentication")
			if len(authHeader) == 0 {
				return nil, status.Error(codes.Unauthenticated, "jwt token missing")
			}
			tokenStr := strings.TrimSpace(authHeader[0])
			claims, err := jwttoken.ParseToken(tokenStr, publicKey)
			if err != nil {
				return nil, fmt.Errorf("%w: error parsing token: %w", errs.Unauthorized, err)
			}
			iss, err := claims.GetIssuer()
			if err != nil || iss != "Payments" {
				return nil, fmt.Errorf("%w: invalid issuer %s", errs.Unauthorized, iss)
			}
			exp, err := claims.GetExpirationTime()
			if err != nil {
				return nil, fmt.Errorf("%w: error getting expiration date: %w", errs.Unauthorized, err)
			}
			if time.Now().After(exp.Time) {
				return nil, fmt.Errorf("%w: token expired at %s", errs.Unauthorized, exp.Time.String())
			}
			sub, err := claims.GetSubject()
			ctx = context.WithValue(ctx, "user_id", sub)
			return handler(ctx, req)
		}
		return handler(ctx, req)
	}
}
