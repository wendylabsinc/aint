# swift-processinfo-environment

**Flags:** direct reads of `ProcessInfo.processInfo.environment`.

**Why it matters:** reading the process environment directly spreads
stringly typed configuration access throughout the application. Call sites
must each decide how to parse values, provide defaults, report missing keys,
and handle secrets. They are also harder to test or compose with configuration
from files and other sources.

**Fix:** use
[swift-configuration](https://github.com/apple/swift-configuration) and set up
a `ConfigReader` with an `EnvironmentVariablesProvider` at the application's
entry point. Pass the reader, or a scoped reader, to code that needs
configuration.

```swift
// Flags this:
let environment = ProcessInfo.processInfo.environment
let port = Int(environment["HTTP_PORT"] ?? "8080")!

// Prefer this:
import Configuration

let config = ConfigReader(provider: EnvironmentVariablesProvider())
let port = config.int(forKey: "http.port", default: 8080)
```

Add the package and product dependencies:

```swift
.package(
    url: "https://github.com/apple/swift-configuration",
    from: "1.0.0"
)

.product(name: "Configuration", package: "swift-configuration")
```

`EnvironmentVariablesProvider` maps configuration keys to conventional
uppercase, underscore-separated environment names, so `http.port` reads
`HTTP_PORT`. A `ConfigReader` can also combine environment variables with
file, command-line, and in-memory providers in a defined precedence order.

**Known exceptions:** low-level libraries that expose a process-environment
abstraction, test utilities that intentionally capture the complete
environment, and code that must pass the environment unchanged to a child
process may need direct access. Disable this convention rule for those targets
or isolate the access behind one adapter.
