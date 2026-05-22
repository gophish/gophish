Awarenox Roadmap
================

Awarenox is a fork of [Gophish](https://github.com/gophish/gophish) that aims
to become a modern, multi-tenant SaaS platform for security awareness — going
beyond phishing simulation into training, integrations, and enterprise
identity.

This document is the source of truth for what we're building and in what
order. It is intentionally opinionated. Anything not listed here is out of
scope until promoted into a phase.

---

## Vision

> **Awarenox is a multi-tenant SaaS for security awareness.**
> Organizations sign up, isolate their data, run phishing simulations,
> deliver microlearning, and pipe everything to their identity, chat, and
> SIEM stacks — without ever touching a YAML file.

Non-goals (explicitly):

- We are **not** building an offensive red-team C2 / payload framework.
  Awareness training and authorized phishing simulation only.
- We are **not** keeping API/database backward compatibility with upstream
  Gophish past Phase 1. Migrations are one-way.
- We are **not** trying to be a full LMS. The training module is a focused
  microlearning experience, not a Moodle replacement.

---

## Baseline: what we inherit from Gophish

What's already here and works (Gophish 0.12.1):

- Go backend with Gorilla mux, GORM v1 (`github.com/jinzhu/gorm`), SQLite /
  MySQL via goose migrations (`db/db_sqlite3`, `db/db_mysql`).
- Domain models: `User`, `Role`/`Permission` (`models/rbac.go`),
  `Campaign`, `Template`, `Page`, `SMTP`, `Group`, `Result`, `MailLog`,
  `IMAP`, `Webhook`.
- Two HTTP servers: admin (`controllers/route.go`) and phishing landing
  (`controllers/phish.go`).
- Outbound webhooks (`webhook/webhook.go`), basic RBAC (admin/user only,
  global — no org scoping).
- jQuery + Bootstrap 3 + Gulp frontend in `static/js/src`, server-rendered
  Go templates in `templates/`.
- Docker build (multi-stage, debian-slim runtime) and GitHub Actions CI
  building against Go 1.21–1.23.

What's missing and we will fix:

- No tenant model. `User`, `Campaign`, etc. have no `OrganizationID`.
- RBAC is global, two-role. No org-scoped roles, no fine-grained perms.
- Auth is username/password + API key. No SSO, no SAML, no OIDC, no MFA.
- GORM v1 is EOL. Many deps are 5+ years old (`gorilla/mux 1.7`,
  `logrus 1.4`, `goquery 1.5`, `goose` from bitbucket, etc.).
- Webhook payloads are minimal and not signed in a modern way.
- No structured event bus → SIEM/SOAR integration is bolt-on per customer.
- Frontend is server-rendered HTML + jQuery; no API consumer story for
  external apps; CKEditor 4 is past EOL.
- Single tenant, single database. No row-level isolation, no per-tenant
  encryption, no billing, no quotas.

---

## Phases

Each phase ships behind its own set of PRs and ends with a tagged release
(`v0.x.0-awarenox`). No phase starts until the previous one is merged to
`main` and has a passing migration path.

### Phase 0 — Rebrand and modernization base (week 1–2)

Goal: rename the project, modernize the toolchain, set up the dev loop.
**Zero new features.** This must be boring and reversible.

Concrete deliverables:

- [ ] Rename Go module `github.com/gophish/gophish` →
      `github.com/wagnerbocchi/awarenox`. Update every import.
- [ ] Bump `go.mod` to Go 1.23, drop `// indirect` clutter, run
      `go mod tidy`.
- [ ] Replace EOL deps:
      - `bitbucket.org/liamstask/goose` → `github.com/pressly/goose/v3`
      - `github.com/jinzhu/gorm` → `gorm.io/gorm` (v2) + `gorm.io/driver/...`
      - `github.com/gorilla/context` → stdlib `context`
      - `gopkg.in/alecthomas/kingpin.v2` → `github.com/alecthomas/kingpin/v2`
      - Bump `gorilla/mux`, `gorilla/csrf`, `gorilla/sessions`, `logrus`,
        `goquery`, `go-imap`, `go-message`.
- [ ] Replace `ioutil` calls (deprecated since Go 1.16).
- [ ] Rewrite Dockerfile: `golang:1.23-alpine` builder, `gcr.io/distroless`
      or `alpine:3.20` runtime, multi-arch (amd64/arm64) via buildx, image
      published to `ghcr.io/wagnerbocchi/awarenox`.
- [ ] CI: add `golangci-lint`, `gosec`, `govulncheck`, `staticcheck` as
      required checks. Add coverage upload (Codecov or in-repo badge).
- [ ] Replace gulp pipeline with Vite (the React migration in Phase 3
      will use it anyway — we set it up now for the legacy assets too).
- [ ] Rebrand: README, LICENSE attribution (keep MIT, credit Jordan
      Wright), logo SVG, favicon, default `<title>`, `templates/base.html`,
      `templates/login.html`, all references in error messages and log
      lines. New default config file `awarenox.json` (with
      backwards-compat fallback to `config.json` reading for one release).
- [ ] Env var rename: `GOPHISH_INITIAL_ADMIN_PASSWORD` →
      `AWARENOX_INITIAL_ADMIN_PASSWORD` (accept both for one release with
      a deprecation log line).
- [ ] `make` targets: `make dev`, `make test`, `make lint`, `make docker`.
- [ ] `docker-compose.yml` for local dev with MySQL 8 + MailHog +
      MinIO (S3-compatible, used in Phase 2).

Exit criteria: `go build`, `go test ./...`, `docker compose up`, and
`make lint` all pass on a fresh clone. CI green on Linux + macOS.

### Phase 1 — Multi-tenant core (week 3–6)

Goal: turn Awarenox into a real multi-tenant data model. Every read and
write is scoped to an organization. No cross-tenant leaks possible at the
ORM layer.

Concrete deliverables:

- [ ] New model `Organization` (`id`, `slug`, `name`, `plan`,
      `created_at`, `deleted_at`, `settings_json`).
- [ ] New model `Membership` (`user_id`, `organization_id`, `role_id`,
      `invited_at`, `accepted_at`).
- [ ] Add `OrganizationID` (indexed, `NOT NULL`) to: `Campaign`,
      `Template`, `Page`, `SMTP`, `Group`, `Target` (group members),
      `Result`, `Event`, `MailLog`, `IMAP`, `Webhook`, `Attachment`.
- [ ] One-shot migration that creates a `default` organization and
      backfills all existing rows to it. Migration is idempotent and
      logs the row counts it touched.
- [ ] **Tenant-scoped GORM scope**: a `WithOrg(ctx)` helper that injects
      `WHERE organization_id = ?` automatically. Refactor every query in
      `models/` to require an `OrgID` (or accept a `Querier` interface
      that already has it bound). Code review checklist: any new query
      must go through `WithOrg`.
- [ ] Middleware `middleware.RequireOrg` that resolves the org from
      either (a) the authenticated user's active membership or (b) the
      `X-Awarenox-Org` header for API tokens scoped to one org.
- [ ] Org-aware RBAC. Replace the global `admin`/`user` roles with:
      - `owner` — full control of the org, billing, delete org
      - `admin` — manage org settings, users, campaigns
      - `operator` — create/run campaigns, no settings
      - `analyst` — read-only, can export reports
      - `auditor` — read-only, no PII (emails redacted)
      Plus a `super_admin` global role for platform operators (no org
      scope; can impersonate read-only with an audit log entry).
- [ ] Per-org API keys. `api_key` moves off `User` and onto a new
      `APIToken` model with `org_id`, `scope`, `expires_at`,
      `last_used_at`. Old user-level keys keep working for one release.
- [ ] Audit log table (`audit_events`): who, when, org, action,
      target_type, target_id, ip, user_agent, before/after JSON. Wired
      into every mutating handler via a single middleware.
- [ ] Per-org quotas (configurable per plan): max users, max recipients
      per campaign, max sends per day, max templates. Enforced at the
      controller layer.
- [ ] Tests: a `tenant_isolation_test.go` that fuzzes every list/get
      endpoint with a foreign org's token and asserts 404 / empty list.

Exit criteria: a script that creates 2 orgs, runs a campaign in each,
and proves no API call can see the other org's data, even by ID
enumeration. Audit log captures every mutation.

### Phase 2 — Storage, queues, observability (week 7–8)

Goal: production-grade infra so multi-tenant scale doesn't fall over.

Concrete deliverables:

- [ ] Asset storage abstraction (`storage.Provider`): local FS (dev),
      S3-compatible (prod). Attachments and report exports move to it.
- [ ] Background job queue. Today campaign sending runs inside the
      single-process `worker/`. Replace with [`river`][river] (PostgreSQL
      jobs, transactional, no Redis dependency) or `asynq` if we stay on
      Redis. Decision recorded in `docs/adr/0002-job-queue.md`.
- [ ] PostgreSQL support added as the recommended production database.
      SQLite stays for dev/eval. MySQL stays for upstream-compat users
      but is deprioritized.
- [ ] Structured logging: `logrus` → `log/slog` (Go 1.21+ stdlib), JSON
      output, with `org_id`, `user_id`, `request_id` on every line.
- [ ] OpenTelemetry: traces + metrics, OTLP exporter, dashboards for
      Grafana shipped in `deploy/grafana/`.
- [ ] Health endpoints: `/healthz` (liveness), `/readyz` (DB + queue +
      storage), `/metrics` (Prometheus).
- [ ] Rate limiting per org (not per IP) via `middleware/ratelimit`.

[river]: https://riverqueue.com

### Phase 3 — New frontend (week 9–14)

Goal: replace the jQuery + Bootstrap 3 + server-rendered admin with a
modern SPA consuming the API. Keep the phishing landing pages
server-rendered (they need to be fast, cookie-light, and template-driven).

Concrete deliverables:

- [ ] OpenAPI 3.1 spec generated from Go handlers (`oapi-codegen` or
      hand-written). Becomes the contract.
- [ ] Frontend: **React + TypeScript + Vite + TanStack Query +
      shadcn/ui + Tailwind**. Lives in `web/`. Built artifacts embedded
      via `embed.FS` in the Go binary.
- [ ] Rewrite each page: Dashboard, Campaigns (list + wizard + results),
      Templates (with a modern WYSIWYG — TipTap, replacing CKEditor 4),
      Landing Pages, Sending Profiles, Groups, Users, Settings, Webhooks.
- [ ] Dark mode, responsive, keyboard accessible (axe-core in CI).
- [ ] Real-time campaign progress via SSE (`/api/campaigns/:id/events`)
      instead of polling.
- [ ] Localization: PT-BR (default), EN, ES.
- [ ] Per-org branding: logo, primary color, custom domain for both the
      admin app and the phishing landing pages.

### Phase 4 — Identity integrations (week 15–17)

Goal: enterprise IdP integration for both login and target-group sync.

Concrete deliverables:

- [ ] **Login (SSO)**:
      - SAML 2.0 (`crewjam/saml`) — per-org IdP config.
      - OIDC (`coreos/go-oidc`) — generic, plus first-party shortcuts
        for Azure AD / Entra ID, Google Workspace, Okta.
      - MFA fallback for password users: TOTP (already partially
        possible with `pquerna/otp`) + WebAuthn (`go-webauthn/webauthn`).
- [ ] Just-in-time provisioning: first SSO login creates the membership
      with a default role chosen by the org admin.
- [ ] **Directory sync** (targets / groups):
      - Microsoft Graph (Azure AD) — pull users/groups, filter by group.
      - Google Workspace Directory API.
      - Okta SCIM 2.0.
      - Generic LDAP / Active Directory (with paged search).
      Scheduled via the Phase 2 job queue, with incremental sync and a
      dry-run preview.
- [ ] SCIM 2.0 inbound endpoint so IdPs can push users at us.

### Phase 5 — Notifications and chat (week 18)

Goal: real-time alerts and reports in the tools security teams live in.

Concrete deliverables:

- [ ] Notification framework (`notify.Channel` interface) with backends:
      Slack (Block Kit), Microsoft Teams (Adaptive Cards), Discord,
      generic webhook, email digest.
- [ ] Per-org rules engine (simple, declarative): "when event ∈ {clicked,
      submitted_data, reported} and campaign.tags ∈ {executives} → send
      to #soc-alerts". No DSL — JSON rules edited in the UI.
- [ ] Scheduled digests: daily/weekly campaign summary, monthly
      awareness score per department.

### Phase 6 — SIEM / SOAR integrations (week 19–20)

Goal: every event Awarenox produces is consumable by a SOC.

Concrete deliverables:

- [ ] Outbound event bus: signed (HMAC-SHA256 or Ed25519) JSON
      envelopes, at-least-once delivery, retry with exponential backoff,
      dead-letter queue surfaced in the UI.
- [ ] First-party exporters:
      - **Splunk** HEC (HTTP Event Collector).
      - **Microsoft Sentinel** Log Analytics ingestion API + a
        Sentinel solution (workbook + analytic rules) in
        `integrations/sentinel/`.
      - **Elastic** via the Elastic Common Schema (ECS) over the Bulk
        API.
      - **CrowdStrike Falcon LogScale / Humio**.
      - **CEF/Syslog** over TLS for everything else.
- [ ] SOAR playbook hooks: a stable webhook contract documented for
      Tines, XSOAR, Swimlane.
- [ ] STIX 2.1 / TAXII export of phishing indicators (sender domains,
      landing URLs, kit hashes) for sharing with ISACs.

### Phase 7 — Email and AI (week 21–23)

Goal: stop pretending SMTP is the only way to send mail in 2026, and add
AI assist where it actually saves time.

Concrete deliverables:

- [ ] Sending providers (in addition to SMTP):
      - Microsoft Graph `sendMail` (delegated + application perms).
      - Gmail API (service account + domain-wide delegation).
      - Amazon SES, SendGrid, Postmark.
      A `mailer.Provider` interface; the sending profile UI picks one.
- [ ] DKIM/SPF/DMARC alignment hints in the sending-profile wizard
      (warn the user before they send from a misaligned domain).
- [ ] Inbound parser for **reported phishing**: existing IMAP support,
      plus Microsoft Graph (subscription on a shared mailbox) and
      Gmail watch. Reported messages become first-class objects in the
      timeline and can match against past campaigns.
- [ ] LLM assist (Anthropic Claude via the API):
      - "Generate template": prompt → email + landing page + lure SMS.
        Brand-voice profile per org. Reviewable diff before save.
      - "Explain this click": summarize why a target clicked given
        their training history and the template content.
      - "Localize template": translate template + landing page to a
        target locale, preserving tracking pixels and links.
      All LLM calls are opt-in per org, audited, and use prompt caching
      (system prompts + brand profile cached) for cost control.

### Phase 8 — Awareness platform (week 24–28)

Goal: this is what makes Awarenox more than "Gophish with SaaS lipstick".

Concrete deliverables:

- [ ] **Microlearning modules**: short video / Markdown lessons with
      embedded quizzes. SCORM 1.2 / 2004 import for customers with
      existing content libraries.
- [ ] **Auto-enrollment rules**: clicked a campaign → enrolled in the
      matching lesson within 24h. Failed quiz → re-enrolled in 14 days.
- [ ] **Awareness score** per user, group, department, org. Composite
      of campaign behavior + training completion + quiz scores +
      reported-phishing accuracy. Exposed in the API and in dashboards.
- [ ] **Certifications**: signed PDF on lesson-track completion, with
      verifiable QR code linking to a public verification page.
- [ ] **Just-in-time training**: when a user clicks a phishing
      simulation, the landing page (instead of just saying "you were
      phished") drops them into a 90-second lesson immediately.

### Phase 9 — SaaS plane (week 29–32)

Goal: turn Awarenox into something we can actually sell to multiple
customers without a manual onboarding call each time.

Concrete deliverables:

- [ ] Self-serve signup, email verification, org creation wizard.
- [ ] Billing: Stripe subscriptions, per-seat + per-recipient metering.
      Usage events emitted from the job queue, aggregated nightly.
- [ ] Plans: Free (1 org, 50 recipients/month), Pro, Business,
      Enterprise. Quotas from Phase 1 are wired to plan limits.
- [ ] Customer portal: invoices, payment method, plan changes,
      cancellation, GDPR/LGPD export and delete-my-data flow.
- [ ] Status page (`status.awarenox.app`) backed by the `/readyz`
      checks of each region.
- [ ] Region selection at signup (EU, US, BR). One Postgres per region,
      no cross-region replication of customer data.

### Phase 10 — Compliance and trust (continuous, kicks off in parallel)

Not a phase you finish — a track that runs alongside Phases 4+.

- [ ] SOC 2 Type II readiness: control mapping, evidence collection,
      tabletop incident response. Drata or Vanta integration.
- [ ] LGPD / GDPR: DPA template, data residency promise, sub-processor
      list page, in-product consent for AI features.
- [ ] Vulnerability disclosure policy + `SECURITY.md` rewrite + a
      `security.txt`.
- [ ] Quarterly third-party pentest. Findings tracked publicly (severity
      + fix date, not details) on a trust page.

---

## Architecture decisions (recorded in `docs/adr/`)

Each major choice gets a one-page ADR. Initial set we will write in
Phase 0:

- ADR-0001: Multi-tenancy strategy — shared database, shared schema,
  row-level filter (chosen) vs. schema-per-tenant vs. db-per-tenant.
- ADR-0002: Job queue — River (Postgres-native) vs. Asynq (Redis).
- ADR-0003: ORM — GORM v2 vs. sqlc + pgx. (Probably GORM v2 to keep the
  migration manageable; revisit at Phase 9 scale.)
- ADR-0004: Frontend framework — React vs. SvelteKit vs. SolidStart.
- ADR-0005: API style — REST + OpenAPI vs. tRPC vs. GraphQL.
- ADR-0006: Authn — Ory Kratos vs. roll-our-own with `crewjam/saml` +
  `coreos/go-oidc`.

---

## What we are explicitly deferring

Listed here so we don't fall into them by accident:

- Mobile apps. The SPA must be responsive; native apps wait until we
  have paying customers asking for them.
- On-prem "appliance" edition. Self-hosted Docker stays supported, but
  no airgap installer, no hardware appliance.
- Marketplace / third-party plugins. Webhooks + the API are the
  extension surface until Phase 9 closes.
- Generative-AI-driven autonomous campaigns ("agent runs phishing for
  you"). Not building that. Human-in-the-loop only.

---

## How we work

- One PR per deliverable checkbox. PR title format:
  `[phase-N] short description`.
- Every PR closes its checkbox in this file as part of the diff.
- Every phase ends with a tagged release and a short blog post in
  `docs/changelog/`.
- Breaking changes are batched at phase boundaries. Inside a phase,
  the API contract is stable.
- Tests: unit for models, integration for controllers (against a real
  Postgres in CI), E2E for critical flows (Playwright, Phase 3+).
- `main` is always deployable. Feature flags via a tiny in-house
  `features` package (no LaunchDarkly until Phase 9).

---

## Open questions (need an answer before the relevant phase)

- [ ] **Phase 0**: keep MySQL as a first-class DB, or deprecate now and
      tell users to migrate to Postgres in Phase 2?
- [ ] **Phase 1**: do we want a `super_admin` impersonation feature on
      day one, or is that a Phase 9 (SaaS plane) concern?
- [ ] **Phase 3**: PT-BR as default UI, or detect from `Accept-Language`
      with a per-org override?
- [ ] **Phase 4**: do we ship our own SAML/OIDC, or adopt Ory Kratos as
      the auth service? Trade-off: control vs. weeks of work saved.
- [ ] **Phase 7**: which LLM provider(s)? Default to Anthropic Claude
      (we like the long context and prompt caching), but offer
      bring-your-own-key for customers with a strict procurement list?
- [ ] **Phase 9**: launch region — start EU+BR (LGPD/GDPR) and add US
      later, or vice versa?
