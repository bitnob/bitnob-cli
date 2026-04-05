# Bitnob CLI Project Plan

## Overview

Bitnob CLI is a developer-focused command-line tool for interacting with Bitnob services in local development, testing, and automation workflows. The product direction is similar to Stripe CLI: optimize the developer loop for authentication, API access, webhook testing, and profile switching.

The first release should prioritize developer productivity over broad API coverage. The CLI does not need to expose every Bitnob capability on day one. It needs to make common integration tasks fast, predictable, and scriptable.

## Product Goal

Enable a developer to install the Bitnob CLI, authenticate, make API requests, and test webhook-driven integrations locally in under 5 minutes.

## Product Principles

- Developer-first: optimize for local integration workflows, not end-user account management.
- Fast feedback loops: shorten the path from command execution to useful result.
- Scriptable by default: every important workflow should work in both human-readable and machine-readable modes.
- Secure enough for real use: interactive auth should be safe, and automation should have a clear credential path.
- Narrow initial scope: ship a credible foundation before expanding into many resource-specific commands.

## Primary User Workflows

1. Install the CLI and verify it is working.
2. Authenticate to Bitnob under one or more named profiles.
3. Inspect current account, project, or workspace context.
4. Make API requests without manually crafting headers.
5. Listen for and forward Bitnob webhooks to a local application.
6. Trigger sandbox or test events for integration testing.
7. Switch active profiles without re-entering credentials.
8. Use the CLI in CI or scripts with non-interactive credentials.

## Non-Goals for V1

- Full parity with every Bitnob dashboard action.
- A large set of resource-specific commands before usage patterns are known.
- Local wallet operations intended for end users.
- Complex project scaffolding or code generation.
- Replacing the Bitnob dashboard.

## Recommended V1 Command Surface

### Core commands

- `bitnob login`
- `bitnob login --profile <name> --client-id --secret-key`
- `bitnob login --interactive`
- `bitnob whoami`
- `bitnob logout`
- `bitnob switch <profile>`
- `bitnob balances`
- `bitnob config set`
- `bitnob config get`
- `bitnob config list`
- `bitnob api get`
- `bitnob api post`
- `bitnob api delete`
- `bitnob listen`
- `bitnob trigger`
- `bitnob version`

### Optional early additions

- `bitnob logs`
- `bitnob doctor`
- `bitnob completion`

## Authentication Strategy

Bitnob CLI should support two authentication modes:

### 1. Interactive login

Intended for developers working locally.

Desired flow:

1. User runs `bitnob login`.
2. CLI prompts for missing credential values securely in the terminal.
3. User enters `client_id` and `secret_key`.
4. CLI stores the credential securely under the selected profile.
5. CLI verifies those credentials against the Bitnob API.

Recommended implementation direction:

- Use secure terminal prompts for interactive login first.
- Store credentials under named profiles such as `sandbox` and `live`.
- Keep browser-based or Kratos-backed login as a later enhancement if the backend exposes a dedicated CLI auth flow.

Open design questions:

- When Bitnob exposes a dedicated `whoami` endpoint, what identity fields should the CLI surface directly?
- Will a future browser-based login replace manual credential entry, or coexist with it?
- How should credential verification failures during interactive login be handled: save-then-fail or fail-without-saving?

### 2. Non-interactive HMAC credential login

Intended for CI, automation, and fallback local use.

Desired flow:

- User runs `bitnob login --profile <name> --client-id <id> --secret-key <key>`.
- CLI stores the client ID and secret key for the selected profile.

Requirements:

- Clear separation between interactive credentials and HMAC credentials.
- Good support for headless environments.
- Explicit warnings and help text for real credentials.

## Profile Model

Profiles should allow users to keep multiple credential sets without rewriting config. A profile should contain:

- auth method
- credential reference
- output preferences

Example desired behavior:

- `bitnob login --profile sandbox --interactive`
- `bitnob login --profile live --client-id <id> --secret-key <key>`
- `bitnob switch sandbox`
- `bitnob config list`

Important distinction:

- the CLI should not invent or locally configure the Bitnob environment
- environment should be treated as server-derived identity data once the real `whoami` endpoint exists
- profile names are local CLI concepts; they may happen to match `sandbox` and `live`, but that is a user convention, not a transport rule

## Output Model

All commands should support structured, machine-readable output by default.

Output expectations:

- errors should be normalized and actionable
- request IDs should be surfaced when available
- verbose mode should help with debugging HTTP requests and webhook forwarding
- help output remains human-readable

## Architecture

The CLI should be structured in layers to keep command handlers thin and testable.

### CLI layer

Responsibilities:

- command definitions
- flag parsing
- help text
- output formatting

### Application layer

Responsibilities:

- auth orchestration
- profile resolution
- command execution logic
- webhook forwarding workflow

### API client layer

Responsibilities:

- HTTP transport
- auth injection
- retries and timeouts
- shared base URL handling
- response normalization

### Storage layer

Responsibilities:

- config file management
- secure credential storage via keychain when available
- fallback storage strategy where secure keychain access is not available

### Local services layer

Responsibilities:

- local callback handling for interactive auth
- webhook forwarding
- optional logs streaming

## Technology Recommendation

Recommended language: Go

Reasoning:

- good fit for distributable CLIs
- strong standard library for networking and HTTP
- easy cross-compilation
- simpler implementation path for long-running commands like `listen`

Recommended initial tooling:

- `cobra` for command structure
- standard library `net/http` initially, unless a clear need for another HTTP client emerges
- `goreleaser` for packaging and release automation
- OS keychain integration for secure credential storage

## Delivery Phases

### Phase 0: Foundation

Goals:

- initialize repository structure
- define coding standards and release approach
- establish configuration and profile model
- add base CLI binary with help and version output

Deliverables:

- module setup
- command skeleton
- config loader
- structured error model
- release pipeline outline

### Phase 1: Authentication and Identity

Goals:

- implement profile-aware credential flows
- support non-interactive HMAC credential auth immediately
- implement prompt-based interactive login

Deliverables:

- `login`
- `login --client-id --secret-key`
- `whoami`
- `logout`
- `switch`

Dependencies:

- backend support for CLI token exchange if interactive auth is included in this phase

### Phase 2: API Access

Goals:

- make generic Bitnob API requests from the CLI
- normalize output and error handling

Deliverables:

- `api get`
- `api post`
- `api delete`
- `--json` support
- request tracing and verbose mode

### Phase 3: Local Integration Workflows

Goals:

- enable webhook testing against local applications
- support test flows against the authenticated Bitnob API

Deliverables:

- `listen`
- `trigger`
- webhook signature verification support if applicable

Dependencies:

- backend support for webhook forwarding, event registration, or tunneling

### Phase 4: Developer Experience Polish

Goals:

- improve usability and automation support
- add diagnostics and operability features

Deliverables:

- shell completions
- `doctor`
- optional `logs`
- improved help and examples
- update notification policy if desired

## Risks and Constraints

- Browser-based auth may appear straightforward because Ory/Kratos already exists, but the CLI token exchange and lifecycle still need explicit design if the product goes that way.
- Webhook forwarding may require backend work and should be treated as a product feature, not just a local networking task.
- A large command surface too early will slow down iteration and increase maintenance cost.
- Credential handling and profile switching need careful design to avoid mixing the wrong credentials into the wrong local profile.

## Success Criteria for Initial Release

The first usable release should satisfy all of the following:

- developer can install the CLI and run `bitnob version`
- developer can authenticate with either interactive prompt flow or non-interactive credential flags
- developer can confirm current identity and connection status
- developer can make authenticated API requests
- developer can retrieve a real authenticated resource such as balances

## Immediate Next Planning Tasks

1. Finalize the auth specification for interactive login, HMAC credentials, token storage, and profile behavior.
2. Define the exact command tree, flags, and command semantics.
3. Define the initial repository structure for a Go implementation.
4. Clarify backend requirements for a real `whoami` endpoint and webhook forwarding.
5. Pick the first milestone to implement and define acceptance criteria.

## Proposed First Milestone

Milestone name: `v0.1 foundation`

Scope:

- initialize Go project
- implement `version`
- implement config and profile loading
- implement `login --profile <name> --client-id --secret-key`
- implement `whoami` using an authenticated API probe

Why this milestone:

- it creates a working vertical slice
- it avoids blocking the entire project on interactive auth design
- it gives the team a usable base for real CLI ergonomics and release mechanics
