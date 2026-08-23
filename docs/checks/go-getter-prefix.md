# go-getter-prefix

**Flags:** a zero-argument, single-result method named `GetX`.

**Why it matters:** Go APIs conventionally name getters after the property itself. Following that convention keeps APIs compact and predictable.

**Fix:** rename `GetName()` to `Name()`. Keep `Get` when it describes a real lookup operation with arguments.

```go
func (u User) Name() string { return u.name }
```
