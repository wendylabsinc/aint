# node-eval

**Flags:** use of the `eval()` function, which executes arbitrary strings as JavaScript code.

**Why it matters:** `eval()` runs strings as code without any validation or sandboxing. If the string is constructed from user input, configuration, or any untrusted source, an attacker can execute arbitrary code in your application's context, potentially stealing data, modifying state, or compromising the entire server.

**Fix:** avoid `eval()` entirely. Use safer alternatives like JSON parsing, a templating engine, or a focused expression parser library. If you must dynamically evaluate code, isolate it in a worker thread or a separate process with restricted permissions.

```javascript
// Flags this:
const userInput = request.query.code;
const result = eval(userInput);  // dangerous!

// Prefer this:
const userInput = request.query.code;
const result = JSON.parse(userInput);  // safe for JSON

// Or use a simple expression evaluator:
const math = require("mathjs");
const result = math.evaluate(userInput);  // limited, safe scope
```
