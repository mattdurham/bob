---
name: bob:go-coding
description: Go coding guidelines for production-quality code — pool lifetimes, concurrency safety, numeric type boundaries, error handling, and test discipline
user-invocable: true
category: reference
---

# Go Coding Guidelines

These guidelines address patterns that appear in code review again and again. Read this before writing Go code that touches pooled resources, concurrent I/O, or external data sources.

Also injected into fix cycles via `.bob/state/fix-prompt.md` so `workflow-coder` follows these rules during every repair iteration.

---

## 1. Pool and Resource Lifetime

**The rule:** Return a pooled object to the pool only when ALL data derived from it has been consumed.

### ❌ Wrong — released while derived data is still live
```go
buf := pool.Get().(*[]byte)
results := process(buf) // results contains slices or references into *buf
pool.Put(buf)           // RACE: results still alive, any access reads freed memory
return results
```

### ✅ Right — caller owns the release
```go
buf := pool.Get().(*[]byte)
results := process(buf)
// Document: caller must call pool.Put(buf) after the last use of results.
return results, buf
```

**Ask before every `pool.Put` / `Release` call:**
- Does this function return anything derived from the pooled object?
- Can any caller still read from that returned data?
- If yes: don't release here — pass ownership to the caller.

**Typed-nil guard:** Any release function that accepts an interface must guard against typed-nil — a `(*T)(nil)` passes a type assertion but panics on dereference:
```go
func ReleaseWidget(w Widget) {
    if c, ok := w.(*concreteWidget); ok && c != nil {
        widgetPool.Put(c)
    }
}
```

**Pool + `testing.AllocsPerRun`:** `sync.Pool` drops entries at GC, which `AllocsPerRun` triggers. Don't assert exactly 0 allocations on pooled code paths — accept ≤1, or warm the pool inside the measurement closure.

---

## 2. Concurrency Safety

### File writes: use unique temp paths
```go
// ❌ Wrong — two concurrent writers for the same destination corrupt each other
tmp := dest + ".tmp"
os.WriteFile(tmp, data, 0o600)
os.Rename(tmp, dest)

// ✅ Right — each writer gets a unique temp name; Rename is atomic
f, err := os.CreateTemp(filepath.Dir(dest), filepath.Base(dest)+".tmp-*")
if err != nil { return err }
tmp := f.Name()
if _, err := f.Write(data); err != nil {
    f.Close()
    os.Remove(tmp)
    return err
}
f.Close()
return os.Rename(tmp, dest)
```

### Eviction / index update: keep lock held across file removal, or use unique names
```go
// ❌ Wrong — new writer can recreate the path between Unlock and Remove
mu.Lock()
delete(index, key)
mu.Unlock()
os.Remove(path) // may delete a newly written file, not the evicted one

// ✅ Safe option A: unique file names per write (from CreateTemp)
//    The evicted path can never alias a new write, so removal is safe after unlock.

// ✅ Safe option B: hold the lock during short removals
mu.Lock()
delete(index, key)
os.Remove(path) // deterministic path — hold lock so no new writer can race
mu.Unlock()
```

### Duplicate writes: use singleflight
```go
// ❌ Wrong — two goroutines both pass the "already cached?" check and both write
if _, ok := index[key]; !ok {
    write(key, value) // both goroutines reach here; loser cleanup deletes winner's file
}

// ✅ Right — deduplicate in-flight writes with singleflight
var sf singleflight.Group
sf.Do(key, func() (any, error) {
    return nil, writeAndIndex(key, value)
})
```

### Goroutine fan-out: always cap concurrency
```go
// ❌ Wrong — unbounded goroutines; can exhaust fds, memory, or downstream rate limits
for _, item := range items {
    go func(item Item) { process(item) }(item)
}

// ✅ Right — bounded worker pool with context cancellation
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8) // tune to downstream capacity; document the rationale
for _, item := range items {
    item := item
    g.Go(func() error { return process(ctx, item) })
}
if err := g.Wait(); err != nil { return err }
```

### Early-stop: cancel the context, don't just return
```go
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()

for _, item := range items {
    result, err := fetch(ctx, item)
    if err != nil { return err }
    if done(result) {
        cancel() // stop in-flight workers
        break
    }
}
```

### Pre-fetch: pipeline, don't batch-all-then-process
```go
// ❌ Wrong — fetches everything before processing; defeats early-stop; spikes memory
all := fetchAll(items)
for _, r := range all { process(r) }

// ✅ Right — bounded prefetch pipeline
sem := make(chan struct{}, 4)
for _, item := range items {
    if ctx.Err() != nil { break }
    sem <- struct{}{}
    go func(item Item) {
        defer func() { <-sem }()
        r := fetch(ctx, item)
        process(r) // cancel ctx on limit reached
    }(item)
}
```

---

## 3. Numeric Type Boundaries

Go's `make`, slice indices, and most standard library functions require `int`, not `int64`. Sizes from external sources (disk, network, protobuf) are often `int64` or `uint64` — always validate and convert explicitly.

```go
// ❌ Won't compile — make requires int
size := computeSize() // returns int64
buf := make([]byte, size)
chunk := buf[offset : offset+length] // offset, length are int64

// ✅ Validate then convert
if size < 0 || size > maxAllowed {
    return fmt.Errorf("invalid size %d", size)
}
buf := make([]byte, int(size))
// same for offset and length
```

**Sizes derived from external data (files, headers, wire format):**
```go
// Subtraction can produce negative results — always check before using
dataSize := totalSize - int64(headerLen)
if dataSize < 0 {
    return fmt.Errorf("corrupt header: negative data size %d", dataSize)
}
if dataSize > maxDataSize {
    return fmt.Errorf("data too large: %d bytes", dataSize)
}
```

**Length fields from untrusted sources:**
```go
n := binary.LittleEndian.Uint32(buf[0:4])
if n == 0 || n > maxLen {
    return fmt.Errorf("invalid length field %d", n)
}
data := make([]byte, int(n))
```

---

## 4. Error Handling

### Distinguish "not found" from "broken"
```go
// ❌ Wrong — treats all failures as a clean miss; hides real I/O errors
data, err := readFromStore(key)
if err != nil {
    return nil, false, nil // permission denied? disk full? both silently ignored
}

// ✅ Right — only absence is a miss; everything else propagates
data, err := readFromStore(key)
if errors.Is(err, fs.ErrNotExist) {
    return nil, false, nil
}
if err != nil {
    return nil, false, fmt.Errorf("store read %q: %w", key, err)
}
```

### Validate untrusted input before use
```go
// Bound reads from external sources
r := io.LimitReader(src, maxBytes)
data, err := io.ReadAll(r)

// Validate length fields before allocation
if n > maxAllowed {
    return fmt.Errorf("length %d exceeds maximum %d", n, maxAllowed)
}
```

### Wrap errors with context
```go
// ❌ Loses call site information
return err

// ✅ Caller knows where and why
return fmt.Errorf("load config from %s: %w", path, err)
```

### Don't swallow errors with blank identifier
```go
// ❌ Silent failure
_ = file.Close()

// ✅ At minimum log; for write paths, return or join the error
if err := file.Close(); err != nil {
    return fmt.Errorf("close %s: %w", file.Name(), err)
}
```

---

## 5. Test Discipline

### Name tests to match what they actually assert
```go
// ❌ Misleading — name promises zero allocs; body allows several
func TestProcessZeroAllocs(t *testing.T) {
    allocs := testing.AllocsPerRun(100, fn)
    assert.Less(t, allocs, 5.0)
}

// ✅ Name reflects the actual regression guard
func TestProcessNoMapAllocPerCall(t *testing.T) {
    allocs := testing.AllocsPerRun(100, fn)
    assert.Less(t, allocs, 2.0)
}
```

### GC-safe tests for weak references and sync.Pool
```go
//go:noinline
func storeValue(c *Cache[Thing]) {
    v := &Thing{id: 1}
    c.Put("key", v)
    // v goes out of scope when storeValue returns
}

func TestEvictedAfterGC(t *testing.T) {
    c := NewCache[Thing]()
    storeValue(c)   // strong ref dropped on return
    runtime.GC()
    runtime.GC()    // two cycles reduces scheduler nondeterminism
    _, ok := c.Get("key")
    assert.False(t, ok)
}
```

### Keep strong references alive until after the assertion
```go
v := &Thing{id: 1}
c.Put("key", v)
got, ok := c.Get("key")
require.True(t, ok)
assert.Equal(t, 1, got.id)
runtime.KeepAlive(v) // prevent v from being collected before Get returns
```

### Concurrent tests must exercise actual concurrent access
```go
// ❌ Only reads are concurrent — misses write/write and read/write races
c.Put("key", v)
var wg sync.WaitGroup
for range 10 {
    wg.Add(1)
    go func() { defer wg.Done(); c.Get("key") }()
}

// ✅ Mix of concurrent reads and writes
var wg sync.WaitGroup
for i := range 10 {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        c.Put(fmt.Sprintf("k%d", i%3), &Thing{id: i})
        c.Get(fmt.Sprintf("k%d", i%3))
    }(i)
}
wg.Wait()
```

### Assert panic values, not just panic occurrence
```go
// ❌ Only verifies a panic happens — message is part of the contract too
require.Panics(t, func() { c.Put("k", nil) })

// ✅ Locks in the contract
require.PanicsWithValue(t, "cache: value must be non-nil", func() {
    c.Put("k", nil)
})
```

---

## Quick Checklist (Before Every Commit)

- [ ] Every `pool.Put` / `Release`: have ALL callers finished using data derived from this object?
- [ ] Every file write: using `os.CreateTemp` + `os.Rename`, not a deterministic `.tmp` path?
- [ ] Every index/eviction update that removes files: either holding the lock, or using unique file names?
- [ ] Every goroutine fan-out: `errgroup.SetLimit` or semaphore in place?
- [ ] Every `int64` / `uint64` size or offset: validated and converted to `int` before `make` or slice?
- [ ] Every external length field: validated for zero, negative, and upper-bound before use?
- [ ] Every store/cache read error: only `fs.ErrNotExist` suppressed as a miss?
- [ ] Every `_ = expr`: is there a reason this error is intentionally ignored? Add a comment.
- [ ] New test named `TestFooZeroAllocs`: does the assertion actually require ≤0?
- [ ] GC-dependent test: `//go:noinline` helper to drop the strong ref + `runtime.KeepAlive` after assertions?
