# go-goroutine-in-loop

**Flags:** a `go` statement lexically inside a `for` or `range` loop.

**Why it matters:** the number of goroutines then follows the input size. Large or attacker-controlled inputs can exhaust memory, sockets, or downstream services.

**Fix:** use a bounded worker pool, semaphore, or `errgroup` with an explicit limit. Propagate cancellation and make the owner wait for every worker.

```go
group, ctx := errgroup.WithContext(ctx)
group.SetLimit(8)
for _, item := range items {
    group.Go(func() error { return process(ctx, item) })
}
return group.Wait()
```
