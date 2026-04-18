# Changelog

All notable changes to this project are documented in this file.

## v0.1.3 - 2026-04-18

### Added
- New command groups for additional API surfaces:
  - `exchange-rates`
  - `withdrawals`
  - `bulk-transfers`
- New payout commands:
  - `payouts list`
  - `payouts get <id>`
  - `payouts account-lookup`
  - `payouts banks <country-code>`
- New trading commands:
  - `trading quotes get <id>`
  - `trading orders cancel <id>`
  - `trading prices get <pair>`
  - scheduled and target order command trees
- New cards commands:
  - `cards encryption-key`
  - `cards team-members`
  - `cards transactions`
  - `cards transactions card <card-id>`
  - `cards transactions get <card-id> <transaction-id>`
  - `cards balance <card-id>`
  - `cards status <card-id>`
- New customers commands:
  - `customers countries`
  - `customers countries states <country-code>`

### Changed
- Aligned multiple service routes with the latest API collection, including updates in addresses, lightning, payouts, trading, and cards.
- Improved payout response compatibility for newer response shapes.
- Expanded CLI tests to cover new command surfaces and validation paths.

## v0.1.2 - 2026-04-18

### Fixed
- Installer checksum verification and cleanup behavior.
