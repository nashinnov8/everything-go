# Auth Module Implementation Guide

This project now has a user module and an auth module boundary:

- `internal/user/...` owns user data, password hashes, HTTP user routes, and the gRPC user service.
- `internal/auth/...` owns login orchestration, token issuing, and auth-facing handlers.
- `api/proto/user/v1/user.proto` is the service-to-service contract.
- `gen/go/user/v1` contains generated gRPC code used by both modules.

The auth module should not import user repositories, user services, or user database models directly. It should talk to the user module through `internal/auth/client/usergrpc`.

## Current Auth Module Skeleton

The current auth skeleton is intentionally small:

- `internal/auth/domain/auth.go`: request, response, token, and identity DTOs.
- `internal/auth/service/service.go`: auth use case interfaces and login orchestration.
- `internal/auth/client/usergrpc/client.go`: gRPC client wrapper for the user module.

The user module exposes:

- `GetUserByID`: read user identity by UUID.
- `ValidateCredentials`: verify email/password inside the user module and return user identity.

This keeps password hashes inside the user module. The auth module receives only the authenticated user identity and then issues tokens.

## Recommended Auth Module Layout

Add these packages as the auth module grows:

```text
internal/auth/
  domain/
    auth.go
  service/
    service.go
  token/
    jwt.go
  transport/
    http/
      handler.go
      routes.go
  client/
    usergrpc/
      client.go
```

If auth runs as a separate process, add:

```text
cmd/auth/
  main.go
```

Keep `cmd/auth/main.go` as a composition root only. It should load config, create logger, dial gRPC clients, construct services, register routes, and start the HTTP server.

## Step 1: Add Auth Configuration

Create config values for the auth process:

```env
AUTH_HTTP_HOST=localhost
AUTH_HTTP_PORT=8081
USER_GRPC_ADDR=localhost:9090
JWT_ISSUER=gin-crud-api-auth
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
JWT_SECRET=change-me
```

For local development, `grpc.WithTransportCredentials(insecure.NewCredentials())` is acceptable. For production, configure TLS or run gRPC only on a trusted internal network.

## Step 2: Implement Token Issuing

Create `internal/auth/token/jwt.go`. The token package should implement `service.TokenIssuer`.

Recommended rules:

- Put only stable claims in the access token: user ID, email, username, issuer, issued-at, expiry.
- Keep access tokens short lived.
- Store refresh tokens server-side if you need revocation.
- Do not put password hashes, permissions from stale sources, or sensitive profile data in JWT claims.

Example shape:

```go
type JWTIssuer struct {
	issuer     string
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func (i *JWTIssuer) IssueTokenPair(ctx context.Context, user domain.UserIdentity) (*domain.TokenPair, error) {
	// Build signed JWT access token here.
	// Build refresh token here.
	// Return domain.TokenPair.
}
```

Use `github.com/golang-jwt/jwt/v5` when you are ready to implement signing.

## Step 3: Add HTTP Transport

Create `internal/auth/transport/http/handler.go`.

The handler should:

- Bind `domain.LoginRequest`.
- Call `AuthService.Login`.
- Return `domain.LoginResponse`.
- Avoid user lookup, password checking, or token signing inside the handler.

Expected route:

```text
POST /api/v1/auth/login
```

Keep HTTP validation and response formatting in the transport layer. Keep business logic in `internal/auth/service`.

## Step 4: Wire gRPC User Client

In `cmd/auth/main.go`, dial the user module and wrap the generated client:

```go
conn, err := grpc.NewClient(
	cfg.UserGRPCAddr,
	grpc.WithTransportCredentials(insecure.NewCredentials()),
)
if err != nil {
	log.Fatal("failed to dial user service", "error", err)
}
defer conn.Close()

userClient := usergrpc.NewClient(conn)
tokenIssuer := token.NewJWTIssuer(...)
authService := service.NewAuthService(userClient, tokenIssuer)
```

Only `internal/auth/client/usergrpc` should know about the generated gRPC client. The auth service should depend on its own `UserClient` interface.

## Step 5: Error Handling

Map gRPC errors at the client boundary:

- `codes.Unauthenticated` to `ErrInvalidCredentials`.
- `codes.NotFound` to `ErrUserNotFound`.
- `codes.InvalidArgument` to `ErrInvalidRequest`.
- unknown errors to `ErrInternalServer`.

The current `internal/auth/client/usergrpc/client.go` already follows this pattern.

## Step 6: Testing Strategy

Test auth in layers:

- Unit test `internal/auth/service` with mocked `UserClient` and `TokenIssuer`.
- Unit test `internal/auth/token` with known signing inputs.
- Handler test login success and validation failure with a fake `AuthService`.
- Integration test gRPC only when both auth and user services are running.

The auth service unit test should not start Gin, gRPC, or PostgreSQL.

## Updating the gRPC Contract

When `api/proto/user/v1/user.proto` changes, regenerate checked-in Go code:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

protoc \
  --go_out=. \
  --go_opt=module=github.com/yourusername/gin-crud-api \
  --go-grpc_out=. \
  --go-grpc_opt=module=github.com/yourusername/gin-crud-api \
  api/proto/user/v1/user.proto
```

If you change the module path in `go.mod`, update the `go_package` option inside the proto file before regenerating.

## Best Practice Checklist

- Keep auth, user, and shared infrastructure in separate packages.
- Keep database access inside repositories.
- Keep password hash verification inside the module that owns the hash.
- Keep token creation behind `TokenIssuer`.
- Keep generated protobuf code under `gen/go`.
- Keep protobuf definitions under `api/proto`.
- Keep process wiring in `cmd/...`, not inside services.
- Do not import `internal/user/repository` from `internal/auth`.
