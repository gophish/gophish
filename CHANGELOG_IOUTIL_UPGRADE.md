# Changelog - Deprecated ioutil Package Upgrade

## Date: 2025-07-16

### Overview
Upgraded all deprecated `ioutil` package calls to use the modern `io` and `os` packages as recommended since Go 1.16. This improves code maintainability and follows current Go best practices.

### Changes Made

#### 1. **controllers/phish_test.go**
- Replaced `io/ioutil` import with `io` and `os`
- Updated function calls:
  - `ioutil.ReadAll(resp.Body)` → `io.ReadAll(resp.Body)` (2 occurrences)
  - `ioutil.ReadFile("static/images/pixel.png")` → `os.ReadFile("static/images/pixel.png")`

#### 2. **controllers/api/import_test.go**
- Removed `io/ioutil` import and added `os`
- Updated function calls:
  - `ioutil.ReadFile(tt.htmlFile)` → `os.ReadFile(tt.htmlFile)`

#### 3. **webhook/webhook_test.go**
- Replaced `io/ioutil` import with `io`
- Updated function calls:
  - `ioutil.ReadAll(r.Body)` → `io.ReadAll(r.Body)`

#### 4. **util/util.go**
- Removed `io/ioutil` import (already had `io` imported)
- Updated function calls:
  - `ioutil.ReadAll(m.Body)` → `io.ReadAll(m.Body)`

#### 5. **models/attachment_test.go**
- Removed `io/ioutil` import (already had `io` and `os` imported)
- Updated function calls:
  - `ioutil.ReadDir("testdata")` → `os.ReadDir("testdata")`
  - `ioutil.ReadAll(reader)` → `io.ReadAll(reader)`

#### 6. **config/config_test.go**
- Fixed undefined `io` error
- Updated function calls:
  - `io.TempFile("", "gophish-config")` → `os.CreateTemp("", "gophish-config")`

### Summary of Replacements

| Deprecated Function | Replacement | Package |
|-------------------|-------------|---------|
| `ioutil.ReadFile` | `os.ReadFile` | `os` |
| `ioutil.WriteFile` | `os.WriteFile` | `os` |
| `ioutil.ReadAll` | `io.ReadAll` | `io` |
| `ioutil.ReadDir` | `os.ReadDir` | `os` |
| `ioutil.TempFile` | `os.CreateTemp` | `os` |

### Benefits
- Follows Go 1.16+ best practices
- Removes deprecated package usage
- Improves code maintainability
- Prepares codebase for future Go versions

### Testing
All changes maintain the same functionality as before. The only modifications were to replace deprecated function calls with their modern equivalents.