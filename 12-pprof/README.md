# 12 — pprof: finding bugs with the tools

Each debug: the broken program, the guess written **before** any tool, what the
tool showed (real output), the fix, and the numbers before/after.

Profiling is added with one extra file per program (`pprof.go`): an `init()`
that starts `net/http/pprof` on `127.0.0.1:6060` (local only, never a public
port).

## Debug #2 — memory bloat (`../debug-02-heap/`)

**Bug report**: memory keeps climbing during the initial sync; only 64 raw
blocks are cached, so it should need a few MB.

**Guess**: heap profile, `inuse_space` (live memory that is never freed). A map
or slice of processed blocks that is never cleaned.

**Tool** (sync finished, height ≈ 10,000):

```
curl -s -o /tmp/heap-before.pb.gz 'http://127.0.0.1:6060/debug/pprof/heap?gc=1'
go tool pprof -top -inuse_space /tmp/heap-before.pb.gz

Showing nodes accounting for 522.56MB, 99.71% of 524.06MB total
      flat  flat%   sum%        cum   cum%
  518.92MB 99.02% 99.02%   518.92MB 99.02%  main.fetchBlock
    3.65MB   0.7% 99.71%     3.65MB   0.7%  main.(*index).add
```

```
go tool pprof -list 'main.fetchBlock' /tmp/heap-before.pb.gz

  518.92MB   518.92MB     27:   raw := make([]byte, headerSize+ntx*txSize)
```

```
go tool pprof -top -inuse_objects /tmp/heap-before.pb.gz

     49190 75.67% 75.67%      49190 75.67%  main.(*index).add
      9841 15.14% 90.81%       9841 15.14%  main.fetchBlock
```

**Reading it**: ~9,800 raw blocks are alive (518.92 MB / 9,841 ≈ 53 KB each),
but the cache keeps 64. The heap profile shows the **allocation site**, not who
**retains** the memory, so the rest came from following `raw` in the code:

- `rc.put(height, raw)`: the cache evicts after 64. Not the bug.
- `parseHeader(raw)` keeps `Hash: raw[0:32]` and `Parent: raw[32:64]` in the
  index forever. A subslice shares the **backing array**, and the GC cannot
  free part of an array: 64 useful bytes keep the whole ~50 KB block alive, for
  every block.
- `string(h.Hash)` as the map key is fine: the conversion copies.

`index.add` has the most **objects** (map buckets + keys) but little
**space**: always check both indexes.

**Fix** (`index.go`): `bytes.Clone(raw[0:32])` and `bytes.Clone(raw[32:64])`.
Better still: make `Hash` and `Parent` `[32]byte`. No pointer, no extra
allocation, and the array can be the map key directly (go-ethereum's
`common.Hash` is a `[32]byte` for the same reason).

| | `inuse_space` total | `fetchBlock` |
|---|---|---|
| Before (height ≈ 10,000) | **524.06 MB** | 518.92 MB |
| After (height ≈ 8,300) | **9.3 MB** | 4.2 MB (the 64 cached blocks) |

The after profile was taken a bit earlier in the sync, but the shape is what
matters: before, memory grows ~50 KB per block; after, it is flat at the cache
size plus 64 bytes per header. `-base` shows the drop directly:

```
go tool pprof -top -base /tmp/heap-before.pb.gz /tmp/heap-after.pb.gz
Showing nodes accounting for -514.71MB, 96.50% of 533.36MB total
```

The numbers are estimates: the heap profile is a sample (`MemProfileRate`,
one sample per ~512 KB allocated), scaled up by pprof.

## Debug #3 — goroutine leak (`../debug-03-leak/`)

**Bug report**: headwatch gets slower and uses more memory the longer it runs;
CPU is low and the heap profile shows nothing big.

**Guess**: none written.

**Tool 1 — is there a leak at all?** `runtime.NumGoroutine()` every 3 s
(added in `pprof.go`):

```
goroutines=291    (20:02:51)
goroutines=579
goroutines=857
...
goroutines=10410  (20:04:39)   ≈ 94 new goroutines a second, never going down
```

**Tool 2 — which goroutines?** `debug=1` groups goroutines with the same stack,
count first. Two snapshots 10 s apart:

```
curl -s 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=1'

goroutine profile: total 3717
3686 @ ...
#	main.(*client).fastestHead.func1+0x57	debug-03-leak/client.go:26
---- 10 s later
goroutine profile: total 4661
4626 @ ...
#	main.(*client).fastestHead.func1+0x57	debug-03-leak/client.go:26
```

**Tool 3 — stuck on what?** `debug=2` prints every goroutine with its state;
counted by state:

```
6230 chan send       ← the leak
  16 sleep           peers answering (time.Sleep)
   8 chan receive    the 8 idle verifiers: normal, constant
   3 select          main loop + tickers

goroutine 4 [chan send]:
main.(*client).fastestHead.func1()
	debug-03-leak/client.go:26 +0x58
created by main.(*client).fastestHead in goroutine 1
	debug-03-leak/client.go:21 +0x98
```

**Cause**: `fastestHead` starts one goroutine per peer (6) and sends each answer
on an **unbuffered** channel, but the `select` receives only the first answer
(or times out) and returns. The other senders block on `answers <- h` forever:
nobody will ever receive. 20 polls a second × ~5 stuck senders ≈ 95 a second.
`cancel()` does not help: the send does not watch `ctx`. The heap profile
missed it because the cost is mostly goroutine **stacks**, which are not in the
heap profile.

**Test** (`client_test.go`): count goroutines, call `fastestHead` 10 times with
4 fast peers, wait 100 ms for the slower peers, count again.

```
before the fix:  client_test.go:23: goroutines: 4 before, 30 after   --- FAIL
after the fix:   --- PASS · go test -count=30 -race: ok
```

**Fix** (`client.go`): `make(chan head, len(c.peers))`. Every sender has a slot,
sends and exits; the channel and the unread answers are garbage once
`fastestHead` returns. (Also correct: `select { case answers <- h: case
<-ctx.Done(): }`.)

| | goroutines |
|---|---|
| Before | 291 → 10,410 in 108 s, only up |
| After | 26–35 for 45 s (900 polls), flat |

**Side lesson**: `pprof.go` uses `init()`, so it also runs in `go test`. With
the program still running, the test binary logged `listen tcp 127.0.0.1:6060:
bind: address already in use`. In a real service, start pprof from `main()`
behind a flag or an env var (blockchain-insight: `ENABLE_PPROF=true`).
