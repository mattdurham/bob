# Go Ownership, Concurrency, and Input Hazards

Apply these checks to relevant changed behavior using the repository's Go
version, documented contracts, workload, and platform. These are investigation
criteria, not a rule that every occurrence requires a finding or a new abstraction.

## Pools and Borrowed Data

Do not return an object to a pool while a caller or worker still uses it or
an alias into its storage. Reusing the pooled buffer can change data observed
through those aliases; pool synchronization alone does not protect borrowers.
Copy retained data or pass an explicit release obligation to its owner. Check
typed-nil values before dereferencing an asserted concrete pointer when the
interface contract permits nil. Make ownership and the end of borrowing clear.

`sync.Pool` is a reuse cache, not persistent storage. Entries can disappear, and
callers must handle a miss. `testing.AllocsPerRun` warms the function once and
measures allocations; it does not force garbage collection. Avoid changing
an allocation budget solely because code uses a pool. Set a justified budget,
keep setup outside the measured work when appropriate, and test cold-pool
behavior separately if it matters to the contract.

When testing weak references or finalization, establish whether a strong
reference remains live and use `runtime.KeepAlive` only where the test requires
it. Do not assume that a fixed number of GC cycles proves arbitrary eviction
or finalizer behavior; test the documented contract and record timing limits.

## Concurrency, File Replacement, and Cancellation

For input-dependent fan-out, bound concurrency using a worker pool, semaphore,
or an appropriate group limit. Choose the bound from memory, I/O, downstream
capacity, and the workload. A small fixed set of goroutines needs no arbitrary
extra limiter. Ensure cancellation reaches workers and any blocked acquisition,
wait for owned workers to finish, and propagate their relevant errors.

When overlapping writes can replace a destination, create a unique temporary
file in its directory, handle write and close errors, and clean up failure paths.
Rename behavior and replacement atomicity are platform dependent; do not
promise cross-platform atomicity or durability solely from `os.Rename`.
Temporary file uniqueness does not serialize publication to the same destination;
use the required ownership, ordering, or version check as well.

For eviction, avoid removing a newly recreated file under the same name. Unique
object identities can permit removal after unlocking; otherwise coordinate the
index update and deletion with writers. Use in-flight deduplication only when
requests with the same key are genuinely interchangeable. For early termination,
a bounded streaming/prefetch pipeline may avoid wasted work; choose it only
when cancellation, result ordering, and resource ownership remain correct.

## Integer Sizes and Bounds

Go allows integer-typed `make` lengths and slice indices, including `int64` and
`uint64`. Runtime lengths must still be representable as `int`, nonnegative,
and valid for the allocation or slice. Validate external sizes against the
resource budget and bounds before use. Check arithmetic before overflow, and
convert to `int` only when the receiving API requires it, after checking the
conversion's target range. Zero is invalid only when the contract says so.

```go
func allocate(n, maxBytes int64) ([]byte, error) {
    if n < 0 || n > maxBytes || uint64(n) > uint64(^uint(0)>>1) {
        return nil, fmt.Errorf("invalid allocation length %d", n)
    }
    return make([]byte, n), nil // int64 length is valid Go
}

func section(data []byte, offset, length int64) ([]byte, error) {
    size := int64(len(data))
    if offset < 0 || length < 0 || offset > size || length > size-offset {
        return nil, fmt.Errorf("invalid section offset=%d length=%d", offset, length)
    }
    return data[offset : offset+length], nil // bounds checked before addition
}
```

Before decoding a length field, check that the input contains the field bytes.
Validate signed subtraction, multiplication, and offset/length addition before
allocating or slicing; checking an already overflowed result is insufficient.

## Errors and Evidence

Treat absence as a cache miss only when the storage contract identifies it as
absence, commonly through `errors.Is(err, fs.ErrNotExist)`. Surface permission,
corruption, short-read, and other meaningful failures. Handle write and close
errors; preserve the original failure if cleanup also fails. Add error context
where it helps callers identify the failed operation, preserving error identity
when the API requires it. An intentionally ignored cleanup error needs a reason,
not a universal logging mandate.

Bound external reads and detect truncation when exceeding the bound is an error;
`io.LimitReader` alone can return a silently truncated prefix. Use the relevant
input and I/O contract to decide whether a partial result is usable.

For changed shared state, exercise the relevant read/write combinations under
`go test -race` when supported. A race-free run does not prove liveness, ordering,
or lifetime safety. Use focused behavioral evidence for those hazards. General
TDD and test-design policy belongs to the repository and `bob-matt-tdd`; this reference
does not impose extra interview, style, or per-commit checklist requirements.
