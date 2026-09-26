# go-fundamentals

Exercises from reading Go's runtime and stdlib source, one folder per topic.
Each one is written from scratch after reading the real implementation,
then compared against it. Notes live in `notes.md`.

| Folder | Source I read | What the exercise shows |
|---|---|---|
| `01-numeric-types/` | spec §Types | overflow and conversion behavior |
| `03-payment/` | `io.go`, `sort.Interface` | small interfaces + mock, avoiding interface pollution |
| `04-hashmap/` | `internal/runtime/maps` (Swiss Tables) | open-addressing map vs stdlib design |
| `linklist/` | `runtime/slice.go` | linked list + iterator |
| `05-goroutine-benchmark/` | `runtime/proc.go` | 1M goroutines: ~2KB stack each vs OS threads |
| `06-counter/` | Go Memory Model | data race → mutex/atomic, verified with `-race` |
| `07-bounded-channel/` | `runtime/chan.go` | bounded channel from scratch (+ a cold rebuild from memory) |
| `07-sync-once/` | `sync/once.go` | why CAS alone is not enough: `BrokenOnce` test proves callers return early |
| `08-select-nil/` | `runtime/select.go`, `runtime/chan.go` | nil channel silences a select case; nil receive with no default = deadlock (`-deadlock` flag) |
| `09-generics/` | `cmp/cmp.go`, `slices/slices.go` | Map/Filter/Reduce/Set from scratch + 3 cases where generics are the wrong tool |
| `cold-rebuild/` | — | weekly rebuilds without looking at code or AI |

```sh
go test -race ./...
```