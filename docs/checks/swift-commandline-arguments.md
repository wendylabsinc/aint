# swift-commandline-arguments

**Flags:** reading `CommandLine.arguments`, `CommandLine.argc`, or
`CommandLine.unsafeArgv` directly.

**Why it matters:** hand-rolled argv handling starts as a two-line
`if arguments.count > 1` and grows into a bespoke parser: no `--help`, no
`--version`, positional/flag ordering bugs, missing validation, and error
messages that don't tell the user what they got wrong. Every CLI reinvents
the same thing slightly worse.

**Fix:** use **swift-argument-parser**. Declare the interface as a type and
let the library generate parsing, validation, and help output.

```swift
// Flags this:
let args = CommandLine.arguments
guard args.count > 1 else {
    print("usage: tool <path>")
    exit(1)
}
let path = args[1]

// Prefer this:
import ArgumentParser

@main
struct Tool: ParsableCommand {
    @Argument(help: "Path to the file to process.")
    var path: String

    @Option(name: .shortAndLong, help: "Output directory.")
    var output: String = "."

    @Flag(help: "Print each file as it is processed.")
    var verbose = false

    mutating func run() throws {
        // ...
    }
}
```

Add the dependency:

```swift
.package(url: "https://github.com/apple/swift-argument-parser", from: "1.5.0")
```

**Known exception:** explicitly forwarding argv into ArgumentParser itself
(e.g. `Tool.parseAsRoot(Array(CommandLine.arguments.dropFirst()))` in a test
harness) is legitimate. `@main` + `ParsableCommand` needs no argv access at
all, so prefer that form where possible.
