

### 📋 AI Agent Prompt & Project Requirements: Gophish Modernization & Security Hardening

**1. Project Overview & Objectives**

* **Goal:** Port and modernize an older fork of the open-source project "Gophish" into a secure, maintainable, and up-to-date codebase.
* **Focus:** Absolute priority is **Security** (Security First). The code must be hardened against common web vulnerabilities (e.g., OWASP Top 10).
* **Workflow:** The AI agent must work incrementally, ensuring all code changes remain locally testable and reviewable.

**2. Git Workflow & Commit Guidelines (Mandatory)**

* **Pull Requests (PRs) Only:** The agent must never push directly to the `main` or `master` branch. Every feature, update, or phase must be submitted as a separate Pull Request.
* **PR Descriptions:** Every PR must include a clear summary of the problem, the solution, and any potential breaking changes or required manual testing steps.
* **Commit Messages:** For every major change, the agent must provide descriptive commit messages following the Conventional Commits specification (e.g., `feat: update GORM to v2`, `fix: secure session cookie flags`, `refactor: upgrade go.mod to 1.22`).

**3. Technology Stack (Current State ➔ Target State)**

* **Language:** Go `1.13` ➔ Go `1.22` (or newer).
* **ORM (Database):** `jinzhu/gorm v1` ➔ `gorm.io/gorm v2` (Expect and handle breaking changes!).
* **Migrations:** `liamstask/goose` ➔ `pressly/goose/v3`.
* **Frontend-Build:** Outdated Node/Gulp ➔ Modern Node LTS (via Docker Multi-Stage Build).
* **Infrastructure:** Legacy Dockerfile ➔ Hardened, distroless/slim Docker image without source code leaks.

**4. Security Requirements (Guardrails)**
The agent must adhere to the following standards for every code generation:

* **Database:** Never use direct string concatenation for SQL queries. Use parameterized queries/binding exclusively to prevent SQL Injection.
* **Authentication:** Verify and upgrade the work factor/cost for `bcrypt` password hashing to current standards.
* **Session Management:** Cookies must strictly enforce `Secure`, `HttpOnly`, and `SameSite=Strict` (or `Lax`) flags.
* **Network:** Disable outdated TLS versions (allow only TLS 1.2 and TLS 1.3). Implement secure HTTP headers (HSTS, CSP, X-Frame-Options).
* **Dependencies:** Do not introduce deprecated packages. Replace vulnerable dependencies with actively maintained alternatives.

**5. Project Phases (Milestones)**
The agent should process the modernization in the following order (creating PRs for each):

* **Phase 1: Infrastructure & Baseline** (Update `go.mod`, rewrite `Dockerfile`, setup linters/scanners like `gosec`).
* **Phase 2: Database & ORM** (Migrate from GORM v1 to v2 and update Goose. Focus on the `models/` directory).
* **Phase 3: Core Security Refactoring** (Audit and refactor `auth.go`, TLS settings, and middleware security).
* **Phase 4: API & Webserver** (Refactor Gorilla Mux routers/handlers, introduce proper `context.Context` management).
* **Phase 5: Frontend & Assets** (Update JS dependencies, mitigate potential XSS vulnerabilities in templates).

**6. Agent Instructions (Rules of Engagement)**

* *Think step-by-step:* Always analyze the existing code context before suggesting refactorings.
* *Explain Breaking Changes:* If an update breaks existing code, clearly list which files will be affected and require modifications.
* *Scope Limit:* Do not generate code for 10 files at once. Provide changes file-by-file or function-by-function to keep PRs reviewable.

* *Testing:* For every change, suggest how to test it locally (e.g., `go test ./...`, manual testing steps for web UI).