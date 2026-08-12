# swift-taskgroup-cancelall-not-deferred

**Flags:** a bare `group.cancelAll()` statement — one that isn't wrapped in
a `defer`.

**Why it matters:** in a "race N tasks, take the first result/error"
pattern, the call that waits for a result (`try await group.next()`,
`try await group.waitForAll()`, etc.) can throw if the first task to
finish failed. A `cancelAll()` written as a plain statement *after* that
call never runs on the throwing path — execution jumps straight past it —
so the group's remaining tasks are never cancelled. They keep running to
completion, wasting work or holding resources (connections, file handles)
longer than intended.

**Fix:** start the block with `defer { group.cancelAll() }` so cancellation
runs unconditionally, regardless of whether the awaited call throws or
returns normally.

```swift
// Flags this:
try await withThrowingTaskGroup(of: Response.self) { group in
    for url in mirrors {
        group.addTask { try await fetch(url) }
    }
    let first = try await group.next()
    group.cancelAll()
    return first
}

// Prefer this:
try await withThrowingTaskGroup(of: Response.self) { group in
    defer { group.cancelAll() }
    for url in mirrors {
        group.addTask { try await fetch(url) }
    }
    return try await group.next()
}
```

**Note:** this check looks for `cancelAll()` written as its own statement
on a line with nothing else. Writing the deferred cancellation as the
one-liner shown above (`defer { group.cancelAll() }`) is both the
recommended fix and what keeps this check quiet — a multi-line `defer { ... }`
block with `cancelAll()` on its own inner line won't be recognized as safe by
this regex-based check; prefer the one-liner form.
