# swift-dispatchsource-signal

**Flags:** installing a signal handler with
`DispatchSource.makeSignalSource(signal:queue:)`.

**Why it matters:** a `DispatchSource` signal handler fires on a GCD queue,
outside structured concurrency. From there you have no ordered way to shut
the process down: nothing cancels in-flight tasks, nothing waits for the
HTTP server to drain before the database pool closes, and the handler
typically ends in a bare `exit(0)` that kills work mid-flight. The
lifecycle also has to be paired with `signal(SIGTERM, SIG_IGN)` to work at
all — easy to forget, and silent when you do.

**Fix:** use **swift-service-lifecycle**. `ServiceGroup` owns signal
handling, starts services in declaration order, and shuts them down in
reverse order, propagating cancellation through structured concurrency.

```swift
// Flags this:
let source = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)
source.setEventHandler { server.stop(); exit(0) }
source.resume()

// Prefer this:
import ServiceLifecycle

let group = ServiceGroup(
    services: [databasePool, httpServer],
    gracefulShutdownSignals: [.sigterm, .sigint],
    logger: logger
)
try await group.run()
```

Services opt into graceful shutdown by iterating with
`cancelOnGracefulShutdown()` or awaiting `withGracefulShutdownHandler`,
so each one finishes its own work rather than being killed by an `exit`.

Add the dependency:

```swift
.package(url: "https://github.com/swift-server/swift-service-lifecycle", from: "2.0.0")
```
