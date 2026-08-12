# swift-unstructured-task

**Flags:** creating an unstructured `Task { ... }`, `Task(priority:...) { ... }`,
or `Task.detached { ... }` — as opposed to a structured child task
(`group.addTask { ... }`, `async let`, or a task passed into
`withTaskGroup`/`withThrowingTaskGroup`).

**Why it matters:** an unstructured `Task` has no parent to inherit
cancellation or propagate its error from. If the enclosing scope is
cancelled or throws, the detached task keeps running — it can outlive the
work that spawned it, leak resources, or silently swallow an error nobody
awaits. Structured concurrency (task groups, `async let`) ties a child
task's lifetime to its parent by construction, so cancellation and error
propagation are automatic.

**Fix:** reach for a task group or `async let` instead. If a genuinely
detached, fire-and-forget task is required (e.g. bridging from synchronous
code at a top-level entry point), that's a deliberate exception — consider
narrowing it and leaving a comment explaining why structure isn't possible
here.

```swift
// Flags this:
Task {
    await process(item)
}

// Prefer this (dynamic number of children):
await withTaskGroup(of: Void.self) { group in
    group.addTask { await process(item) }
}

// Or this (fixed, known-at-compile-time set of children):
async let result = process(item)
await result
```
