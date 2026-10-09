# Security Policy

## Supported versions

Botla is developed on a single line. Security fixes go to the `main` branch and are not backported to older commits or tags.

| Version | Supported |
|---|---|
| `main` | Yes |
| Anything older | No |

## Reporting a vulnerability

**Please don't report security issues in public GitHub issues, discussions or pull requests.**

Report them privately through GitHub Security Advisories: open the repository's **Security** tab and click **Report a vulnerability**. Please include:

- the affected component (API endpoint, widget, dashboard, ingestion, etc.) and the commit you tested
- steps to reproduce or a proof of concept
- the impact as you understand it, such as data exposure, cross-tenant access or code execution

If you can't use GitHub Security Advisories, open a public issue that asks for a private contact channel, without any vulnerability details.

## Scope

In scope:

- the Go API (`cmd/`, `internal/`, `pkg/`): authentication and session handling, tenant isolation between organizations and workspaces, plan and quota enforcement, admin authorization
- URL and file ingestion, including SSRF and parser issues
- the embeddable widget and its public chat endpoints: domain allowlist and embed-token bypasses, XSS in rendered answers
- the dashboard (`frontend/`)
- prompt injection that leads to a concrete security impact, such as leaking another tenant's data or calling actions the bot owner didn't configure
- the Docker, Compose and CI configuration in this repository

Out of scope:

- a model giving a wrong or off-topic answer with no security impact (that's a quality issue; please open a normal issue)
- vulnerabilities in third-party services (OpenAI, OpenRouter, Qdrant Cloud, Cloudflare) or in dependencies that are already publicly known, unless Botla uses them in an unsafe way
- denial of service through traffic volume, and findings that need physical access or a compromised developer machine
- missing hardening headers or best-practice suggestions without a demonstrated impact

## What to expect

- **Acknowledgement** within 3 working days.
- **Initial assessment** (severity and whether it's in scope) within 10 working days.
- **Fix**: critical and high-severity issues are prioritised and usually fixed within 30 days. You'll get updates while the work is in progress.
- **Disclosure**: once a fix is on `main`, an advisory is published. Reporters are credited unless they'd rather not be.

Botla is maintained by one person, so these are good-faith targets rather than contractual guarantees. Please give a reasonable window for a fix before disclosing publicly.
