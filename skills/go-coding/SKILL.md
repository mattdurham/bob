---
name: bob:go-coding
description: Go coding guidelines targeting patterns that cause bugs in real production code — pool lifetime contracts, concurrency safety, int64/int type boundaries, error handling, spec accuracy, and test discipline
user-invocable: true
category: reference
---

# Go Coding Guidelines

These guidelines exist because the same class of bugs appears in code review again and again. Read this before writing Go code that touches pooled resources, concurrent file I/O, storage-backed paths, or spec-driven modules.

This is also invoked automatically during fix cycles via `.bob/state/fix-prompt.md` so that workflow-coder follows these rules during every repair iteration.

---

## 1. Pool and Resource Lifetime

**The rule:** Return a pooled object to the pool only at the true end-of-life of all data derived from it.

### ❌ Wrong — released too early
```go
internMap := pool.AcquireInternMap()
rows := streamSortedRows(block, internMap) // rows hold references into internMap
pool.ReleaseInternMap(internMap)           // RACE: rows still alive, lazy decode will read internMap
return rows
```

### ✅ Right — caller owns the release
```go
internMap := pool.AcquireInternMap()
rows := streamSortedRows(block, internMap)
// Document: caller must call pool.ReleaseInternMap(internMap) after
// the last GetField/IterateFields call on any row in `rows`.
return rows, internMap
```

**Key checks before every pool.Put / Release call:**
- Am I inside a function that returns data derived from this pool object?
- Can any caller of MY function (or my caller's caller) still call GetField / IterateFields / decodeNow on data that touches this pool object?
- If yes: don't release here. Return the pool object to the caller.

**Typed-nil guard:** Any `ReleaseXxx(interface{})` function must guard against typed-nil:
```go
func ReleaseAdapter(p SpanFieldsProvider) {
    if a, ok := p.(*modulesSpanFieldsAdapter); ok && a != nil {
        putAdapter(a)
    }
}
```

**Pool + AllocsPerRun:** `sync.Pool` drops entries at GC. Don't assert exactly 0 allocations in `testing.AllocsPerRun` on pooled code paths. Accept ≤1 or warm the pool inside the closure.

---

## 2. Concurrency Safety

### File writes: use unique temp paths
```go
// ❌ Wrong — concurrent writes to same key corrupt each other
tmp := path + ".tmp"
os.WriteFile(tmp, data, 0600)
os.Rename(tmp, path)

// ✅ Right — each write gets a unique temp name
f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
if err != nil { return err }
tmp := f.Name()
// write to f, then:
f.Close()
os.Rename(tmp, path)
```

### Eviction: don't release the lock before deleting files
```go
// ❌ Wrong — new Put can recreate the file between Unlock and Remove
c.mu.Lock()
delete(c.index, key)
c.mu.Unlock()
os.Remove(filePath)  // removes the NEW file, not the evicted one

// ✅ Option A: hold the lock during removal (simple, correct for short-lived I/O)
c.mu.Lock()
delete(c.index, key)
c.mu.Unlock()
// Only safe if filePath is unique-per-write (from CreateTemp), not deterministic

// ✅ Option B: unique file names — the evicted file's path can never alias a new write
```

### Same-key concurrent Put: use singleflight or per-key lock
```go
// Both goroutines pass the "already exists?" check → both write → loser deletes winner's file
// Use singleflight.Group or mark the key as "in-flight" under the main lock
var sf singleflight.Group
sf.Do(key, func() (interface{}, error) {
    return nil, writeAndIndex(key, value)
})
```

### Goroutine fan-out: always cap concurrency
```go
// ❌ Wrong — one goroutine per group, no limit
for _, grp := range groups {
    go func(g Group) { process(g) }(grp)
}

// ✅ Right — bounded worker pool
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8) // document why 8
for _, grp := range groups {
    grp := grp
    g.Go(func() error { return process(ctx, grp) })
}
```

### Early-stop: cancel, don't just return
```go
// When a query limit is satisfied, cancel the context so workers stop
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()
// ...
if err == errLimitReached {
    cancel()
    break
}
```

### Pre-fetch patterns: pipeline, don't batch-all-then-process
```go
// ❌ Wrong — defeats early-stop, spikes memory
allBytes := fetchAll(groups)  // Phase 1: fetch everything
process(allBytes)              // Phase 2: process (limit may be hit on group 1)

// ✅ Right — pipelined with bounded prefetch
sem := make(chan struct{}, 4) // 4 groups in flight at a time
for _, grp := range groups {
    if ctx.Err() != nil { break }
    sem <- struct{}{}
    go func(g Group) {
        defer func() { <-sem }()
        data := fetch(ctx, g)
        process(data) // if limit → cancel(ctx)
    }(grp)
}
```

---

## 3. Type Safety: int64 and int Boundaries

Go `make`, slice indices, and most library functions require `int`, not `int64`. Sizes read from disk or computed from file offsets are often `int64`. You must convert with a bounds check before use.

```go
// ❌ Won't compile
buf := make([]byte, blockLen)       // blockLen is int64
slice := buf[colStart:colStart+colLen]  // colStart, colLen are int64

// ✅ Convert with bounds check
if blockLen > int64(maxBlockSize) || blockLen < 0 {
    return fmt.Errorf("invalid block length %d", blockLen)
}
buf := make([]byte, int(blockLen))
colStartInt := int(colStart) // after validating colStart >= 0 && colStart <= int64(len(buf))
```

**Pattern for sizes from disk:**
```go
valueSize := info.Size() - int64(fileHeaderLen) - int64(keyLen)
if valueSize < 0 {
    return fmt.Errorf("corrupt file: negative value size %d", valueSize)
}
if valueSize > maxValueSize {
    return fmt.Errorf("value too large: %d bytes", valueSize)
}
```

---

## 4. Error Handling

### Cache/store misses: only treat IsNotExist as a miss
```go
// ❌ Wrong — hides real I/O failures
data, err := readCacheFile(path)
if err != nil {
    return nil, false, nil  // silently pretends it's a miss
}

// ✅ Right — only NotExist is a miss
data, err := readCacheFile(path)
if os.IsNotExist(err) {
    return nil, false, nil
}
if err != nil {
    return nil, false, fmt.Errorf("cache read %s: %w", path, err)
}
```

### Protect against corruption in untrusted data
```go
// When reading length fields from disk:
keyLen := binary.LittleEndian.Uint32(header[4:8])
if keyLen == 0 || keyLen > maxKeyLen {
    return fmt.Errorf("invalid key length %d in %s", keyLen, path)
}

// Use LimitReader on untrusted streams:
r := io.LimitReader(f, maxValueSize)
data, err := io.ReadAll(r)
```

---

## 5. Spec Accuracy

**When you change a contract, update the spec file in the same commit.**

| Change | What to update |
|--------|----------------|
| New public function or type | SPECS.md — add contract, invariants, back-ref |
| Changed function signature | SPECS.md — update entry, add NOTES.md entry |
| New design decision / algorithm | NOTES.md — dated entry with `NOTE-XXX` ID |
| New test function | TESTS.md |
| New benchmark | BENCHMARKS.md |

**Spec IDs in code:** Tag every non-trivial implementation with the relevant spec ID:
```go
// SPEC-007: single I/O per block — never issue per-column reads
func (r *Reader) GetBlockWithBytes(ctx context.Context, id ulid.ULID) ([]byte, error) {
```

**Back-refs in specs must be valid:** When writing `Back-ref: pkg/file.go:FunctionName`, verify the function exists with that exact name. Renames without updating specs are a HIGH-severity finding.

**"Experimental" vs "active use":** If you promote an experimental code path to the hot path, update the spec section that says "retained for future use" to say "in active use as of [date]".

---

## 6. Test Discipline

### Name tests accurately
```go
// ❌ Misleading — allows boxing allocations
func TestIterateFieldsZeroAllocs(t *testing.T) {
    allocs := testing.AllocsPerRun(100, fn)
    assert.Less(t, allocs, 5.0) // "zero allocs" but allows 5?
}

// ✅ Name matches assertion
func TestIterateFieldsNoSeenMapAllocs(t *testing.T) {
    allocs := testing.AllocsPerRun(100, fn)
    assert.Less(t, allocs, 2.0) // guards against the seen-map alloc regressing
}
```

### GC-safe weak.Pointer and sync.Pool tests
```go
//go:noinline
func putValue(c *objectcache.Cache[myStruct]) {
    v := &myStruct{n: 1}
    c.Put("key", v)
    // v goes out of scope when putValue returns
}

func TestGCEviction(t *testing.T) {
    c := objectcache.New[myStruct]()
    putValue(c)          // strong ref dropped when putValue returns
    runtime.GC()
    runtime.GC()         // two cycles reduces nondeterminism
    got, ok := c.Get("key")
    assert.False(t, ok, "expected GC to evict")
    _ = got
}
```

### KeepAlive in ownership tests
```go
v := &myStruct{n: 1}
c.Put("key", v)
got, ok := c.Get("key")
require.True(t, ok)
assert.Equal(t, 1, got.n)
runtime.KeepAlive(v) // prevent v from being collected before Get
```

### Concurrent tests must be actually concurrent
```go
// ❌ Tests concurrent Get but not concurrent Put
func TestConcurrentPutGet(t *testing.T) {
    c.Put("key", v)  // single put before goroutines
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() { defer wg.Done(); c.Get("key") }()  // only Get is concurrent
    }
}

// ✅ Tests both concurrent Put and Get
func TestConcurrentPutGet(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            v := &myStruct{n: i}
            c.Put(fmt.Sprintf("key-%d", i%5), v)
            c.Get(fmt.Sprintf("key-%d", i%5))
        }(i)
    }
    wg.Wait()
}
```

### Assert panic messages, not just panic occurrence
```go
// ❌ Only checks that panic happens
require.Panics(t, func() { c.Put("k", nil) })

// ✅ Checks the contract (message is part of the spec)
require.PanicsWithValue(t, "objectcache: value must be non-nil", func() {
    c.Put("k", nil)
})
```

---

## Quick Checklist (Before Every Commit)

Before submitting Go code, run through:

- [ ] Every `pool.Put` / `Release`: have ALL callers finished using data from this pool object?
- [ ] Every file write path: using `os.CreateTemp` + `os.Rename`, not a deterministic `.tmp` path?
- [ ] Every eviction: file removal happens either under lock, or the file name is unique (non-deterministic)?
- [ ] Every goroutine fan-out: `errgroup.SetLimit` or semaphore in place?
- [ ] Every `int64` size/offset: converted to `int` with bounds check before `make` or slice?
- [ ] Every error from a cache/store: only `os.IsNotExist` suppressed as a miss?
- [ ] Every size read from disk: validated for negative values before arithmetic or cast?
- [ ] Changed a contract? SPECS.md / NOTES.md updated in same commit?
- [ ] New test named `TestFooZeroAllocs`? Does it actually assert ≤0 (or is the name a lie)?
- [ ] New GC-dependent test? `//go:noinline` helper + `runtime.KeepAlive` in place?
