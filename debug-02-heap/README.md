# Debug #2 — blockidx

**Bug report**: "blockidx's memory keeps climbing during the initial sync and
never comes back down. It only keeps 64 raw blocks in its cache, so it should
need a few MB."

```sh
go run ./debug-02-heap        # Ctrl-C to stop
```

No profiling is built in: add what you need.

Rules: AI-free · write the guess before running any tool · find the line with
the tool first, read the code after · before/after numbers.
