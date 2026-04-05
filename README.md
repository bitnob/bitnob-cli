# Bitnob CLI

Bitnob CLI is a developer-focused command-line tool for API testing and integration workflows.

## Commands

The CLI currently includes:

- `bitnob version`
- `bitnob api get|post|put|patch|delete`
- `bitnob addresses list|get|create|validate|supported-chains`
- `bitnob balances`
- `bitnob wallets`
- `bitnob beneficiaries list|get|create|update|delete`
- `bitnob cards create|get|details|list|fund|withdraw|freeze|unfreeze|terminate|spending-limits|customer`
- `bitnob config list|get|set|unset|edit`
- `bitnob customers list|get|create|update|delete|blacklist`
- `bitnob doctor`
- `bitnob lightning invoices|decode|payments`
- `bitnob listen`
- `bitnob payouts quotes|initialize|beneficiary-lookup|simulate-deposit|fetch|country-requirements|supported-countries|limits`
- `bitnob trading quotes|orders|prices`
- `bitnob transfers create`
- `bitnob transactions list|get`
- `bitnob wait card-active|lightning-paid|transaction-settled`
- `bitnob login --client-id --secret-key`
- `bitnob login --interactive`
- `bitnob login --profile <name>`
- `bitnob whoami`
- `bitnob logout`
- `bitnob switch <profile>`
- `bitnob shell --mode command`
- `bitnob help`

The CLI command layer is implemented with Cobra.

Login verifies credentials with `/api/whoami` before saving them. If `--profile` is omitted, the profile name defaults to the environment returned by `whoami`, for example `live` or `sandbox`.

## Paths and Overrides

By default, the CLI stores configuration in the user config directory:

- macOS: `~/Library/Application Support/bitnob/config.json`
- Linux: `~/.config/bitnob/config.json`

If `os.UserConfigDir()` is unavailable, the fallback path is:

- `~/.bitnob/config.json`

You can override the default paths with:

- `BITNOB_CONFIG_PATH`
- `BITNOB_STATE_DIR`
- `BITNOB_SECRET_BACKEND` (`auto`, `keyring`, or `file`)

`BITNOB_STATE_DIR` is available for local CLI state. By default (`BITNOB_SECRET_BACKEND=auto`), credentials are stored in OS keyring when available, and fall back to a file secret store inside `BITNOB_STATE_DIR` when keyring is unavailable.

Examples:

```bash
BITNOB_CONFIG_PATH=/tmp/bitnob/config.json go run ./cmd/bitnob config list
BITNOB_STATE_DIR=/tmp/bitnob-state go run ./cmd/bitnob whoami
```

## Build

```bash
go build ./cmd/bitnob
bash scripts/build.sh
```

`scripts/build.sh` injects build metadata (`version`, `commit`, `date`) into the binary.

## Staging Smoke Suite

Run the staging smoke checks with isolated temp config/state:

```bash
BITNOB_CLIENT_ID=your-client-id \
BITNOB_SECRET_KEY=your-secret-key \
./scripts/staging_smoke.sh
```

Options:

- `BIN` (default `./bitnob`)
- `PROFILE` (default `staging`)
- `RUN_WRITE=1` to enable write-path probe (`customers create`)

## Run

```bash
go run ./cmd/bitnob version
go run ./cmd/bitnob api get /api/balances
go run ./cmd/bitnob api post /api/customers --data '{"email":"john@example.com"}'
go run ./cmd/bitnob addresses list --page 1 --limit 20
go run ./cmd/bitnob addresses get address_123
go run ./cmd/bitnob addresses create --chain bitcoin --label home --reference wallet_ref_123
go run ./cmd/bitnob addresses validate --address 36FJHVQDuEqRCz8Qcx3ZCkAcHnmC6sgPZjvUVgTUtvHY --chain solana
go run ./cmd/bitnob addresses supported-chains
go run ./cmd/bitnob balances
go run ./cmd/bitnob balances get USDT
go run ./cmd/bitnob wallets
go run ./cmd/bitnob beneficiaries list --limit 10 --offset 0
go run ./cmd/bitnob beneficiaries get beneficiary_123
go run ./cmd/bitnob beneficiaries create --company-id company_123 --type bank --name "Bank Beneficiary" --account-name "John Doe" --account-number 1234567890 --bank-code 058
go run ./cmd/bitnob cards list --status active --card-type virtual --limit 10
go run ./cmd/bitnob cards get card_123
go run ./cmd/bitnob cards details card_123
go run ./cmd/bitnob cards create --customer-id customer_123 --card-type virtual --card-brand visa --currency USD --billing-line1 "123 Main Street" --billing-city Lagos --billing-state Lagos --billing-postal-code 100001 --billing-country NG
go run ./cmd/bitnob cards fund card_123 --amount 500.00 --currency USD --reference fund_001
go run ./cmd/bitnob cards spending-limits card_123 --single-transaction 2000.00 --daily 10000.00 --allowed-categories travel,software
go run ./cmd/bitnob cards customer customer_123
go run ./cmd/bitnob customers list --page 1 --limit 20
go run ./cmd/bitnob customers get customer_123
go run ./cmd/bitnob customers create --customer-type individual --email john@example.com --first-name John --last-name Doe
go run ./cmd/bitnob doctor
go run ./cmd/bitnob lightning invoices create --amount 50000 --description "Payment for order" --reference order_12345
go run ./cmd/bitnob lightning invoices get invoice_123
go run ./cmd/bitnob lightning invoices list --page 1 --page-size 20
go run ./cmd/bitnob lightning invoices verify --payment-hash hash_123
go run ./cmd/bitnob lightning decode --request lnbc500u1...
go run ./cmd/bitnob lightning payments initiate --request lnbc500u1...
go run ./cmd/bitnob lightning payments send --request lnbc500u1... --reference payment_123
go run ./cmd/bitnob listen --secret your-webhook-secret
go run ./cmd/bitnob listen --secret your-webhook-secret --forward-to http://localhost:3000/webhook
go run ./cmd/bitnob payouts quotes create --from-asset BTC --to-currency NGN --source onchain --chain BITCOIN --amount 0.0015 --payment-reason "Vendor payout" --reference offramp-quote-001 --country NG
go run ./cmd/bitnob payouts initialize QT_6158 --customer-id customer_123 --beneficiary '{"account_number":"0123456789","bank_code":"058","account_name":"Ada Okafor"}' --reference offramp-init-001 --payment-reason "Supplier settlement"
go run ./cmd/bitnob payouts initialize QT_6158 --customer-id customer_123 --beneficiary-id beneficiary_123 --reference offramp-init-001 --payment-reason "Supplier settlement"
go run ./cmd/bitnob payouts beneficiary-lookup --country NG --account-number 0123456789 --bank-code 058 --type bank_account
go run ./cmd/bitnob payouts quotes list --order DESC --page 1 --take 10
go run ./cmd/bitnob payouts quotes get QT_120732
go run ./cmd/bitnob payouts quotes finalize QT_6158
go run ./cmd/bitnob payouts fetch id 4d808ea1-0d16-419c-b545-881821c3e904
go run ./cmd/bitnob payouts fetch reference offramp-quote-001
go run ./cmd/bitnob payouts country-requirements NG
go run ./cmd/bitnob payouts supported-countries
go run ./cmd/bitnob payouts limits
go run ./cmd/bitnob trading quotes create --base-currency BTC --quote-currency USDT --side buy --quantity 0.0001
go run ./cmd/bitnob trading orders create --base-currency BTC --quote-currency USDT --side buy --quantity 0.0001 --price 90822.540125 --quote-id quote_123 --reference 324567890
go run ./cmd/bitnob trading orders list
go run ./cmd/bitnob trading orders get order_123
go run ./cmd/bitnob trading prices
go run ./cmd/bitnob transfers create --to-address 0xdaBe4B0Ca57dfBF13763D5f190A2d30B94f1Bf59 --amount 2000000 --currency USDT --chain bsc --reference testing_bsc_013 --description "Testing on USDT to binance"
go run ./cmd/bitnob transactions list --page 1 --limit 20 --status success --type credit
go run ./cmd/bitnob transactions list --cursor your_next_cursor
go run ./cmd/bitnob transactions get txn_123
go run ./cmd/bitnob wait card-active card_123 --timeout 2m --interval 5s
go run ./cmd/bitnob wait lightning-paid invoice_123 --timeout 2m --interval 5s
go run ./cmd/bitnob wait transaction-settled txn_123 --timeout 2m --interval 5s
go run ./cmd/bitnob config list
go run ./cmd/bitnob config get profile.active
go run ./cmd/bitnob config unset profile.active
go run ./cmd/bitnob config edit
BITNOB_CLIENT_ID=your-client-id BITNOB_SECRET_KEY=your-secret-key go run ./cmd/bitnob login
go run ./cmd/bitnob login --interactive
BITNOB_SECRET_KEY=your-secret-key go run ./cmd/bitnob login --interactive --client-id your-client-id
go run ./cmd/bitnob login --profile live --interactive
BITNOB_CLIENT_ID=your-client-id BITNOB_SECRET_KEY=your-secret-key go run ./cmd/bitnob login --profile sandbox
go run ./cmd/bitnob whoami
go run ./cmd/bitnob shell --mode command
go run ./cmd/bitnob help
```
