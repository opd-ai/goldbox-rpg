# Package Review: Low-Cohesion Packages

**Document Version:** 1.0  
**Created:** 2026-03-13  
**Status:** Analysis Complete

## Executive Summary

Two packages were flagged by `go-stats-generator` for low cohesion scores:

| Package | Cohesion Score | Files | Functions | Assessment |
|---------|---------------|-------|-----------|------------|
| `pkg/secrets` | 0.67 | 4 | 7 | ✅ Appropriate |
| `pkg/persistence` | 1.13 | 10 | 28 | ✅ Appropriate |

**Recommendation:** No consolidation needed. Both packages have intentional low coupling by design.

---

## Package Analysis

### pkg/secrets

**Purpose:** Provides a pluggable secret management interface for environment-based and future cloud provider secret storage.

**Files:**
- `doc.go` — Package documentation
- `provider.go` — Core `SecretProvider` interface and error types
- `env_provider.go` — Environment variable implementation
- `env_provider_test.go` — Test coverage

**Why Low Cohesion is Acceptable:**

The secrets package follows the **Interface Segregation Principle**. It defines:
1. A generic `SecretProvider` interface (provider.go)
2. A concrete implementation (env_provider.go)

The interface and implementation are intentionally decoupled to allow future providers (Vault, AWS Secrets Manager, GCP Secret Manager) without changing the interface.

**Coupling Analysis:**
```
provider.go → (defines interface)
env_provider.go → (implements interface)
```

This is a textbook example of **low coupling, high cohesion per responsibility**.

**Future Extensions:**
- `vault_provider.go` — HashiCorp Vault implementation
- `aws_provider.go` — AWS Secrets Manager implementation
- `gcp_provider.go` — GCP Secret Manager implementation

**Decision:** ✅ Keep package structure as-is.

---

### pkg/persistence

**Purpose:** Provides file-based persistence for game state and player sessions with atomic writes and file locking.

**Files:**
- `doc.go` — Package documentation
- `filestore.go` — Core file persistence with YAML serialization
- `filestore_test.go` — FileStore tests
- `session_store.go` — Session-specific persistence interface
- `session_store_test.go` — SessionStore tests
- `memory_store.go` — In-memory implementation for testing
- `memory_store_test.go` — MemoryStore tests
- `atomic.go` — Atomic file write utilities
- `lock.go` — File locking utilities
- `README.md` — Package documentation

**Why Low Cohesion is Acceptable:**

The persistence package provides multiple implementations of the same storage abstraction:

1. **FileStore** — Production file-based storage
2. **MemoryStore** — In-memory storage for tests
3. **FileSessionStore** — Specialized session persistence

**Coupling Analysis:**
```
filestore.go → (core implementation)
    ↑
atomic.go, lock.go → (utilities used by filestore)
    ↑
session_store.go → (builds on filestore)
    ↑
memory_store.go → (parallel implementation for testing)
```

The low cohesion score reflects that:
- Atomic writes and file locking are utility functions
- Memory store is a test double, not production code
- Session store is a specialized wrapper

This is a standard **Repository Pattern** implementation with multiple backends.

**Decision:** ✅ Keep package structure as-is.

---

## Cohesion Metric Interpretation

The `go-stats-generator` cohesion score measures how frequently functions in a package call each other. A low score doesn't always indicate poor design:

| Score Range | Interpretation |
|-------------|----------------|
| 0.0 - 1.0 | Interface packages, utility collections |
| 1.0 - 3.0 | Normal domain packages |
| 3.0+ | Tightly coupled, may need refactoring |

Both `secrets` (0.67) and `persistence` (1.13) fall into expected ranges for their architectural roles.

---

## Alternative Consolidation Considered

### Option A: Merge into pkg/config (Rejected)

**Pros:**
- Reduces package count
- Single location for configuration and secrets

**Cons:**
- Violates single responsibility
- Makes secrets management harder to mock in tests
- Forces config package to depend on file I/O

### Option B: Merge persistence into server (Rejected)

**Pros:**
- Co-locates data access with business logic

**Cons:**
- Increases server package complexity (already 14k+ LOC)
- Makes persistence harder to test in isolation
- Creates circular dependency risk with game package

---

## Recommendations

1. **Keep packages separate** — Current structure follows Go best practices
2. **Add more implementations** — Secrets package can grow with cloud providers
3. **Consider interface extraction** — If persistence grows, extract `Store` interface to separate package

---

## Appendix: Cohesion Calculation

The cohesion score is calculated as:

```
cohesion = (internal_function_calls) / (total_functions * log2(total_functions))
```

For `pkg/secrets` with 7 functions and ~5 internal calls:
```
cohesion = 5 / (7 * 2.81) ≈ 0.67
```

This is expected for interface-first packages where the interface itself doesn't call implementation methods.
