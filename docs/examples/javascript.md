# JavaScript

## Path Traversal Detection

Detects potential path traversal vulnerabilities in Node.js filesystem operations.

### What it does

Checks for unsafe usage of filesystem operations like `fs.readFile`, `fs.writeFile`, etc., where user input might be used without proper sanitization.

### Why it matters

Without proper path validation, attackers could access files outside intended directories using `../` in file paths.

### Examples

**Dangerous:**
```javascript
// Don't do this
const userFile = req.query.filename;
fs.readFile(userFile, (err, data) => {
    res.send(data);
});
```

**Safe:**
```javascript
// Do this instead
const path = require('path');
const safePath = path.join(__dirname, 'uploads', fileName);
fs.readFile(safePath, (err, data) => {
    res.send(data);
});
```

### Prevention
1. Always sanitize file paths
2. Use `path.join()` for combining paths
3. Keep files in a specific directory
4. Check if paths are valid before using them

### Rule Configuration
```yml
language: js
name: js_path_traversal
message: Potential path traversal vulnerability detected
pattern: |
  (call_expression
    function: (member_expression
      object: (identifier) @obj
      property: (property_identifier) @method)
    (#match? @obj "^(fs|path)$")
    (#match? @method "^(readFile|readFileSync|writeFile|writeFileSync|unlink|readdir|access)$"))
```

## Insecure Eval Usage

Detects usage of `eval()` and `Function` constructor with potentially untrusted input.

### What it does

Identifies calls to `eval()` or the `Function` constructor that could execute arbitrary JavaScript code.

### Why it matters

Using `eval()` or `Function` constructor with untrusted input can lead to code injection vulnerabilities.

### Examples

**Dangerous:**
```javascript
// Don't do this
eval(userInput);
new Function(userInput)();
```

**Safe:**
```javascript
// Do this instead
const result = JSON.parse(jsonString);
// Use proper data parsing/validation methods
```

### Prevention
1. Avoid using `eval()` completely
2. Use JSON.parse() for JSON data
3. Implement proper input validation
4. Use safer alternatives like template literals

### Rule Configuration
```yml
language: js
name: js_no_insecure_eval
message: Avoid using eval or the Function constructor with untrusted input
pattern: |
  (call_expression
  function: (identifier) @js_no_insecure_eval
  (#match? @js_no_insecure_eval "^(eval|Function)$"))
```

## Weak Cryptography

Detects usage of cryptographically weak hash functions.

### What it does

Identifies usage of deprecated hash functions like MD5 and SHA-1.

### Why it matters

MD5 and SHA-1 are cryptographically broken and vulnerable to collision attacks.

### Examples

**Dangerous:**
```javascript
// Don't do this
const hash = crypto.createHash('md5');
const hash = CryptoJS.MD5(data);
```

**Safe:**
```javascript
// Do this instead
const hash = crypto.createHash('sha256');
const hash = crypto.createHash('sha512');
```

### Prevention
1. Use strong hash functions (SHA-256, SHA-512)
2. Keep cryptographic libraries updated
3. Follow current cryptographic best practices
4. Regularly review and update hash functions

### Rule Configuration
```yml
language: js
name: js_weak_crypto
message: Using cryptographically weak hash function
pattern: |
  (call_expression
    function: (member_expression
      object: (identifier) @obj
      property: (property_identifier) @method)
    (#match? @obj "^(crypto|CryptoJS|md5|sha1)$")
    (#match? @method "^(MD5|SHA1|sha1|md5)$"))
```

### Further Reading
- [OWASP Path Traversal](https://owasp.org/www-community/attacks/Path_Traversal)
- [OWASP Input Validation Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html)
- [NIST Cryptographic Standards](https://www.nist.gov/cryptography)
