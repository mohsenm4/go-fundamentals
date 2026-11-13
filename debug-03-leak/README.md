# Debug #3 — headwatch

**Bug report**: "headwatch gets slower and uses more memory the longer it
runs. CPU is low and the heap profile shows nothing big. A restart fixes it
for a while."

```sh
go run ./debug-03-leak        # Ctrl-C to stop
```

No profiling is built in: add what you need.

Rules: AI-free · write the guess before running any tool · find the line with
the tool first, read the code after · fix + a test that fails before the fix.
