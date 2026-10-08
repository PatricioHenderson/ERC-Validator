# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

ERC-Validator: a set of independent Go microservices (each its own Go module) that fetch and validate on-chain contract data. The root `go.mod` (`erc-validator`) exists only to pin the `golangci-lint` tool version — it is not a shared module for the services.

## Commands

All commands run from the repo root via the `Makefile`.

```
make build              # compile all 5 binaries into ./bin
make build-api          # compile a single service (build-helpers/build-api/build-validator/build-web3/build-admin)
make fmtvet             # go fmt + go vet across every service
make tidy               # go mod tidy across every service
make pre-commit         # tidy + fmtvet + build (run this before committing)
make test               # go test ./... -v across every service
make docker-build-up    # pre-commit, then docker-compose build --no-cache && up
```

To run/test a single service directly (each has its own go.mod):
```
cd services/<name> && go test ./... -v                 # all tests
cd services/<name> && go test ./internal/routes/handlers -run TestCreateUserHandler_Success -v   # one test
cd services/<name> && go build -o ../../bin/<name> ./cmd
```

CI (`.github/workflows/ci.yml`) discovers every `go.mod` in the repo and runs `go fmt`, `go vet`, and `go test` per module — mirror that when validating changes across services.

`.golangci.yml` at the repo root is currently empty (no active lint config beyond `go vet`).

## Architecture

### Multi-module layout

Each service under `services/<name>/` is a **separate Go module** with its own `go.mod`/`go.sum`, own `cmd/main.go` entrypoint, and its own `internal/` package tree. Cross-service imports go through `replace` directives in `go.mod` (e.g. `services/api/go.mod` has `replace erc-validator/helpers => ../helpers`), not through a workspace. When adding a dependency between services, add both the `require` and the matching `replace` line.

Services:
- **`helpers`** — shared library, not a runnable service. `db/connection` opens the shared Postgres/GORM connection (reads `DATABASE_URL`, enforces `POSTGRES_CA_PATH`+TLS when `NODE_ENV=production`); `http_errors` is the standard JSON error envelope (`{"error": "...", ...fields}`) used by handlers that want structured error responses.
- **`api`** — the public-facing gateway. `internal/routes/proxy.go` builds one `httputil.ReverseProxy` per backend service (`admin`, `validator`, `web3`, read from `ADMIN_SERVICE_URL`/`VALIDATOR_SERVICE_URL`/`WEB3_VALIDATOR_URL`) keyed by the first path segment, and `internal/routes/index.routes.go` splits routes into a public group (`/admin/users/login`, `/admin/users/create`) and a private group behind `middleware.Auth` that proxies everything else. `helpers/auth` issues/verifies the JWTs (HS256, `JWT_SECRET` required when `ENVIRONMENT=production`, falls back to a fixed dev secret otherwise).
- **`admin`** — owns user accounts. `internal/models` has the GORM models (`User`, `Token`, many2many `user_tokens`); `internal/routes/handlers/user_handler.go` has create/login/logout/me/change-password. Note: it imports `erc-validator/api/helpers/auth` directly for token creation/verification rather than going through `helpers` — that's the existing pattern, follow it rather than duplicating auth logic.
- **`web3`** — talks to chain RPC providers. `internal/models/rpc.go` defines `RpcEndpoint` (per-`ChainID`, typed `query`/`execution`, with an `Active` flag) looked up from Postgres to choose which RPC URL to dial; `internal/rpc/client.go` caches `ethclient.Client` instances per URL in a `sync.Map`; `internal/routes/handlers/provider_handler.go` (`GET /web3/contract?chainId=&address=`) resolves the RPC endpoint for a chain, dials it, and returns bytecode/block-number info via `types.FetchContractResponse`.
- **`validator`** — intended to validate/persist contract data (`internal/models/contract.go` defines `Contract`, keyed on `(ChainID, Address)`); `cmd/main.go` is currently a stub (routes/handlers not wired up yet — don't assume its HTTP surface exists).

### Conventions to preserve

- DB access goes through a package-level `db.Conn *gorm.DB` var per service (set once in `cmd/main.go` after `connection.ConnectToDb()`), not injected — handlers reference it directly.
- Each service loads its own `.env` via `godotenv.Load()` in `main()`; `.env.sample` per service lists required vars (`PORT`, `DATABASE_URL`, plus service-specific ones like `ADMIN_SERVICE_URL`/`JWT_SECRET`/`ENVIRONMENT` for `api`).
- Routing uses `gorilla/mux` in every service that has HTTP routes.
- docker-compose maps: `admin`→3000 in-container (host mapping duplicated with `api`, check before running both via compose), `validator`→3001, `web3`→3002; a single shared Postgres `db` container backs all of them.
