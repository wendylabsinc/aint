# go-large-interface

**Flags:** an interface declaration containing more than five methods or embedded members.

**Why it matters:** large provider-side interfaces couple consumers to methods they do not use and make substitutes and tests harder to write.

**Fix:** define a small interface in the consuming package containing only the behavior that consumer needs. The five-member threshold is a heuristic, so this check has `info` severity by default.

```go
type ConfigReader interface {
    ReadConfig(context.Context) (Config, error)
}
```
