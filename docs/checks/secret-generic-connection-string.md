# secret-generic-connection-string

**Flags:** a `postgres://`, `postgresql://`, `mysql://`, `mongodb://`, or
`mongodb+srv://` connection string with a literal username and password
embedded in it (`scheme://user:pass@host`).

**Why it matters:** a connection string with a plaintext credential is a
secret sitting in source control — anyone with read access to the repo (or
its history, even after a later "fix") can connect directly to the
database. Connection strings routinely get copy-pasted into config files,
scripts, and example code, so this pattern shows up more than most.

**Fix:** load the credential from an environment variable, secrets manager,
or config file that's excluded from version control, and interpolate it
into the connection string at runtime.

```go
// Flags this:
dsn := "postgres://admin:s3cr3t@db.example.com:5432/mydb"

// Prefer this:
dsn := fmt.Sprintf("postgres://%s:%s@db.example.com:5432/mydb",
    os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))
```
