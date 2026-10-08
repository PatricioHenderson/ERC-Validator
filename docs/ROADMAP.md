# ERC-Validator Roadmap: open source transition

## Background

This project was shelved after an initial build-out, originally intended to be monetized. It is now being finished and made open source: public repo, hosted on a free tier (Railway), usable either through the hosted public API (no login, rate-limited) or by cloning and running it locally.

## What "validating" a contract means

A validation result is **not** a boolean and **not** a formal-verification certificate. Three levels of rigor were considered and rejected down to the one that is actually buildable by a solo maintainer on a free hosting tier:

1. Bytecode/selector grepping — rejected: produces false positives, doesn't execute anything.
2. A fixed battery of read-only `eth_call`s — rejected as the sole method: it never exercises real state transitions (transfer/approve/transferFrom), so "all queried responses are compatible" never implies "complies with the standard."
3. Formal verification of properties (Certora/K framework) — rejected for v1: this is a research project in its own right, not a feature. Under that bar, even fuzzing/invariant tests that find no violation still wouldn't count as a certifying proof.

**Decision: fork-based scenario + fuzz testing.** A local fork (Anvil) is spun up at the current block, a real balance is prepared for a test account (see "Balance preparation" below), and a fixed scenario matrix is run against real state transitions, plus native Go-driven fuzzing/invariant checks from v1.

Scenario matrix (ERC-20): transfers of partial amount, full balance, zero value, and self-transfer; the receiving account re-transferring what it received to prove the balance is actually usable; approve that sets/replaces/revokes an allowance; `transferFrom` against a finite allowance (partial spend, then exhausting it) without assuming `uint256.max` gets decremented (OpenZeppelin's infinite-allowance convention); and rejected operations (over-balance or over-allowance spends), accepting either a revert or a `false` return — the standard permits both, and a test runner must not treat that ambiguity as failure.

Every scenario captures the **actual return value** (via an `eth_call` identical to the state-changing call, executed immediately before the real `eth_sendTransaction` — never just `receipt.status == 1`, which only reports that execution didn't revert, not what the function returned), balances/allowance before and after, and emitted events (`Transfer`/`Approval`) via `eth_getLogs`.

## Result schema: two axes, never a boolean

```
scenarioValidation: PASSED | FAILED | INCONCLUSIVE | ERROR   // did the suite run, and what did it observe?
conformance:        VALIDATED | NON_COMPLIANT | UNPROVEN     // what does that imply about standard compliance?
```

`VALIDATED` is effectively unreachable without a formal-proof layer — a token that passes every scenario still only reaches `conformance: UNPROVEN`. The product is **a verified-behavior report with an explicit scope**, never a binary "is ERC-20" certification. This was an explicit, accepted tradeoff, not an oversight.

`conformance: NON_COMPLIANT` is reserved for an unambiguous, reproduced violation (e.g. a transfer that returns success but doesn't move the balance, or doesn't emit the required event) — the counterexample is stored alongside the verdict.

## Balance preparation: two strategies, never one mandatory path

The goal is a test account with a valid token balance — not necessarily the "real" current holder, since what's being tested is the correctness of transfer/approve/transferFrom, not the provenance of a balance.

1. **Storage injection (primary)**. A Go reimplementation of the slot brute-force algorithm behind `forge-std`'s `stdStorage.find()`/`deal()`: candidate slots for `balanceOf[testAccount]` (the standard Solidity `mapping(address => uint256)` storage layout) are probed via `anvil_setStorageAt`, then verified by reading back `balanceOf(testAccount)`; `totalSupply` is adjusted the same way to stay consistent. This avoids depending on `eth_getLogs`/holder discovery against rate-limited public RPCs as the default path.
2. **Holder impersonation (fallback)**. If slot injection can't find a valid slot within a bounded number of attempts (non-standard accounting — rebasing tokens, share-based balances) — fall back to finding a real holder via `eth_getLogs` on `Transfer` events, confirmed with a `balanceOf` read, and impersonate it with `anvil_impersonateAccount`.

The report always records which strategy (`balancePreparationStrategy`) was used; if neither works, the result is `scenarioValidation: INCONCLUSIVE`, never a silent failure and never interpreted as non-compliance.

## Proxies

There is no single reliable "implementation address." EIP-1967 direct and beacon slots are resolved; diamonds (EIP-2535) are detected via the loupe interface and marked `unresolved` — an empty slot is never interpreted as "not a proxy." A validation report is an immutable historical record (base block + hash + suite version); it is not reused indefinitely as a claim about current state if the implementation has since changed.

## Service boundaries

`validator` owns the Anvil fork lifecycle and the scenario/fuzz runner directly — dozens of stateful calls against a *local* fork don't cross a service boundary, so routing them through `web3` over HTTP doesn't make sense. `web3` stays read-only: chain bytecode, `RpcEndpoint` lookups, and holder discovery (the fallback balance-preparation path).

## API: asynchronous by necessity

Spinning up a fork, preparing a balance, and running ~20+ stateful transactions takes seconds to tens of seconds — not viable as a synchronous endpoint on a public, no-login API on a free hosting tier (DoS surface + timeouts). The API is job-based: `POST /validate {chainId, address}` → `{jobId, status}`, `GET /validations/{jobId}` → job + report once done. Reports are immutable; "current" validation for an address is a query over existing reports, which may enqueue a fresh run.

## Known gaps (not solved yet, not to be silently assumed away)

- Holder discovery via `eth_getLogs` over a block range is heavy on free-tier RPCs and may need archive state; this is likely the dominant cause of `INCONCLUSIVE` results until mitigated.
- Anvil's memory/startup footprint on Railway's free tier is unvalidated — a spike (Phase 1 below) must confirm this before the rest of the engine is built out, since it may force a change in the hosting model.
- Generalizing the fuzz/invariant harness for arbitrary, unknown ERC-20 bytecode (actors, initial balances/allowances, random operation sequences) is the single highest scope-risk piece of this project. Included in v1 by explicit choice, with that risk acknowledged.
- Diamond (EIP-2535) selector→facet resolution, and behavior changes driven by contract state rather than implementation changes (e.g. a pause flag), are explicitly out of scope — flagged as `unresolved`/limitations, not silently ignored.

## Public access model

The hosted public instance is open, no login, protected by per-IP rate limiting. The existing `admin` user-account/JWT system is not a requirement for using the hosted instance — it remains available only for anyone running their own instance.

---

## Phased implementation plan

### Phase 0 — OSS hygiene (low risk, independent of the rest)

Gets the repo into publishable shape before the new validation engine is touched.

1. **Remove tracked `.env` files from git.** `services/{admin,api,validator,web3}/.env` are tracked. `git rm --cached` all four, add `.env` (not `.env.sample`) to the root `.gitignore`.
2. **No insecure JWT fallback.** `services/api/helpers/auth/auth.go` (`init()`) falls back to a hardcoded `"secret-key"` whenever `ENVIRONMENT != "production"`. It should always require `JWT_SECRET` from the environment (fail-fast if missing), regardless of `ENVIRONMENT`. Document in every `.env.sample`.
3. **Move `auth` out of `api` into `helpers`.** `admin` currently imports `erc-validator/api/helpers/auth` directly, coupling two services that shouldn't know about each other. Move the package to `services/helpers/auth/`, update imports in `api` (middleware, login handler) and `admin` (user handler), and the `go.mod`/`replace` directives of both.
4. **Fix the broken test.** `services/admin/internal/routes/handlers/user_handler_test.go::TestLoginUserHandler_Success` expects a 303 redirect to `/home/user`, but `LogInUserHandler` returns 200 + JSON with the token — the current behavior is correct for a JSON API. Fix the test's expectation, not the handler.
5. **`docker-compose.yml` port mapping.** Doesn't match what each service actually listens on (`admin` runs on `PORT=3004` per its `.env` but is mapped `3000:3000`, colliding with `api`'s own `3000:3000`). Fix all four: `api`→3000, `admin`→3004, `validator`→3001, `web3`→3002.
6. **Broken Dockerfiles.** The three existing ones (`api`, `validator`, `web3`) have `COPY go.mod go.sum /` (wrong path) and a fixed `EXPOSE 3000` regardless of the service. Fix `COPY`/`WORKDIR` and `EXPOSE` per service; add a Dockerfile for `admin` (missing despite being referenced in `docker-compose.yml`).
7. **Remove repo cruft**: personal roadmap notes and scratch drawing files that don't belong in a public repo (already captured elsewhere) — remove from tracking, keeping git history.
8. **LICENSE**: add MIT (default recommendation for a permissive Go OSS tool/API) — confirm before committing if a different license is preferred.
9. **Real README**: setup (Docker vs. local), per-service environment variables, available endpoints, and an explicit section on what a validation result means (`scenarioValidation`/`conformance`, and why it isn't a certification) so nobody reads `PASSED` as "guaranteed to be ERC-20."

### Phase 1 — Feasibility spike: Anvil on Railway's free tier

**Before building the full engine**: a minimal Dockerfile that installs Foundry (`foundryup`) and runs `anvil --fork-url <mainnet RPC> --fork-block-number <N>`, deployed by hand to Railway's free tier, measuring real startup time and memory usage. If it doesn't fit comfortably, the hosting model needs revisiting (e.g. a paid tier for `validator`, or moving the fork to an external service) before investing in the rest of the phases. This is the highest-risk point in the whole plan — do it first and report the result before continuing.

### Phase 2 — Data model (`services/validator`)

New GORM models in `services/validator/internal/models/`, replacing the current `Contract` (which has no result fields today):

- `ValidationJob`: `ID`, `ChainID`, `Address`, `BlockNumber` (fixed when the job starts), `SuiteVersion`, `Status` (`queued`/`running`/`done`/`error`), `CreatedAt`, `CompletedAt`.
- `ValidationReport` (1:1 with the job): `ScenarioValidation`, `Conformance`, `BalancePreparationStrategy`, `ProxyResolution` (JSON: detected type — none/1967-direct/1967-beacon/diamond-unresolved, implementation address if resolved), `ScenarioResults` (JSON per-scenario detail: name, outcome, before/after balances/allowance, observed events), `FuzzSummary` (JSON: iterations run, invariants checked, counterexample if any), `ErrorMessage` (if `ERROR`).
- Reports are **immutable** once `done`/`error` — a new validation of the same `(chainId, address)` creates a new job, never overwrites an old one. "Latest validation" is a query (`ORDER BY CreatedAt DESC LIMIT 1`), not an upsert.

### Phase 3 — Balance preparation

1. **Storage injection (primary, lives in `validator`, no `web3` dependency)**: new package `services/validator/internal/erc20check/balanceprep/` implementing the slot brute-force + `anvil_setStorageAt` approach described above.
2. **Holder impersonation (fallback)**: add to `services/web3/internal/rpc/client.go` (or a new `holder.go`) a function that finds a real holder via `eth_getLogs`, confirmed via `balanceOf`. Exposed as `GET /web3/holder?chainId=&address=` in `services/web3/internal/routes/handlers/`, following the same pattern as `provider_handler.go`.

### Phase 4 — `validator`: Anvil fork lifecycle

New package `services/validator/internal/fork/`:

- `fork.Start(ctx, rpcURL string, blockNumber uint64) (*Fork, error)`: spawns `anvil` as a subprocess with a fixed `--fork-url`/`--fork-block-number` (never "latest", so reports are reproducible), polls `eth_blockNumber` until ready.
- `Fork.Client()`/`Fork.RPCClient()`: connections to the local node for both standard JSON-RPC and Anvil-specific methods (`anvil_impersonateAccount`, `anvil_setBalance`, `anvil_setStorageAt`).
- `Fork.Close()`: kills the process, called via `defer` and backed by a job-level timeout.

### Phase 5 — ERC-20 scenario runner

New package `services/validator/internal/erc20check/`. Given a `Fork` + token address + a prepared test account (from Phase 3): fund it and two fresh test accounts with ETH for gas, run the fixed scenario matrix described above, classify each scenario by what was observed (never by the absence of an exception), and aggregate into `scenarioValidation`/`conformance` per the rules above. The mapping "`scenarioValidation == PASSED` → `conformance = UNPROVEN`, never `VALIDATED`" is documented inline in code, since it is not obvious why a passing suite doesn't imply certification.

### Phase 6 — Native Go fuzzing/invariants

Same `Fork`, new file in `erc20check/`. A bounded loop (configurable N iterations) generating random `{actor, operation, amount}` sequences over the already-funded accounts, checking invariants after each step (balances + totalSupply consistency, allowance never increasing without an explicit `approve`). A broken invariant is strong evidence for `NON_COMPLIANT`, stored in `FuzzSummary` with the exact reproducing sequence — along with how many operations were actually accepted vs. rejected, so a suite that reverted everything can't be reported as "passed."

### Phase 7 — Asynchronous API + rate limiting

- `services/validator/internal/routes/`: `POST /validate` (returns a cached report if one exists and isn't forced, otherwise creates a `ValidationJob`), `GET /validations/{jobId}`.
- In-process goroutine worker pool (no external queue infra — out of scope for the free tier); known limitation: a restart loses queued jobs.
- Per-IP rate limiting in `services/api` (the public gateway) using `golang.org/x/time/rate`.

### Phase 8 — Railway deployment

- One Railway service per microservice + the shared Postgres add-on, using the Phase 0 Dockerfiles.
- `validator`'s image needs Foundry installed (per the Phase 1 spike).
- Environment variables set per service via Railway, never committed.
- CORS enabled on `api` so the public API is browser-consumable.

## Verification

- Phase 0: `make pre-commit` passes across all 5 modules; `make test` green including the corrected login test; `docker-compose up` starts all four services without a port collision.
- Phase 1: a concrete RAM/startup-time report for Anvil forking real mainnet on Railway's free tier — gates whether the rest proceeds as designed.
- Phases 2-6: integration tests in `services/validator` against a known real ERC-20 (e.g. USDC on a mainnet fork) expecting `PASSED`/`UNPROVEN`, and against a non-ERC-20 contract (e.g. an NFT) expecting early interface detection to exclude it from the ERC-20 flow.
- Phase 7: `POST /validate` followed by polling `/validations/{id}` until `done`; a second call for the same `chainId`+`address` returns the cached report without re-running the fork.
- Phase 8: a real Railway deploy, smoke-tested against all four public endpoints.
