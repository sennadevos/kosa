# kosa CLI

Personal finance from the terminal. Fast data entry, balance queries, loan tracking.

## Transactions

```bash
# Spend
kosa spend 4.50 "Coffee" --cat dining_out
kosa spend 87.30 "Albert Heijn" --cat groceries

# Income
kosa income 3200 "Salary" --cat salary

# Short aliases
kosa s 4.50 "Coffee"
kosa i 3200 "Salary"

# Transfer between accounts
kosa transfer 500 --from "ING Checking" --to "ING Savings"

# Refund (linked to original transaction)
kosa refund 80 "Zara jacket return" --cat clothing --of 142

# Foreign currency
kosa income 460 "Client invoice" --cat freelance --foreign 500 USD
kosa spend 58 "UK bookshop" --foreign 50 GBP

# Tags
kosa spend 120 "Dinner" --cat dining_out --tag date-night --tag anniversary
```

Default account is the primary checking account (configurable). Override with `--account`.

## Loans

```bash
# Quick shortcuts (no interest)
kosa owe 10 "Pizza" --to "Bas"
kosa lent 150 "Bike repair" --to "Jan"

# Full loan creation
kosa loan new "Personal loan" 5000 --from "ABN AMRO" \
  --interest periodic --rate 1.5 --period monthly --due 2027-04-01

# Record a payment
kosa loan pay 3 50          # loan ID 3, pay 50

# List loans
kosa loans                  # all unsettled
kosa loans --payable        # what you owe
kosa loans --receivable     # what others owe you

# Loan details (shows replay: payments, interest accrual, remaining)
kosa loan show 3
```

## Splitting Costs

```bash
# Equal split — total divided evenly among you + friends
kosa split 400 "Airbnb Amsterdam" --cat travel --with "Bas,Jan,Lisa"

# Unequal split
kosa split 400 "Dinner" --cat dining_out --with "Bas:150,Jan:120" --mine 130

# You fronted but didn't participate
kosa split 300 "Gift for Tom" --with "Bas:150,Jan:150" --mine 0
```

Creates your expense + receivable loans + loan-linked transactions in one command.

## Recurring Rules

```bash
# Add
kosa recurring add "Spotify" 9.99 --type expense --freq monthly --day 15 --cat subscriptions
kosa recurring add "Salary" 3200 --type income --freq monthly --day 25 --cat salary

# List
kosa recurring list

# Pause / resume
kosa recurring pause 5
kosa recurring resume 5
```

## Balance

```bash
# Current balance (all accounts)
kosa balance

# Projected balance on a future date
kosa balance --on 2026-05-15
kosa balance --on 2026-05-15 --account "ING Checking"

# Record a balance snapshot
kosa snapshot 1234.56 --account "ING Checking"
kosa snapshot 5000.00 --account "ABN Savings" --source bank_import
```

## Reconciliation

```bash
# Compare actual transactions to recurring rule projections
kosa reconcile --month 2026-04

#  Rule             Expected   Actual   Delta    Status
#  Salary            3200.00  3000.00  -200.00   linked
#  Rent               850.00   850.00     0.00   linked
#  Netflix              9.99     9.99     0.00   linked
#  Spotify              9.99        -        -   missing
```

## Spending Summary

```bash
kosa summary --month
kosa summary --month 2026-05
kosa summary --from 2026-04-01 --to 2026-04-30
```

## Querying

```bash
# Recent transactions
kosa list
kosa list --limit 50
kosa list --cat groceries
kosa list --tag date-night
kosa list --from 2026-04-01 --to 2026-04-07

# Search by description
kosa search "Albert Heijn"

# Output formats
kosa list --json
kosa list --toon            # token-efficient for LLM context
```

## Configuration

```toml
# ~/.config/kosa/config.toml
[backend]
type = "teable"

[backend.teable]
url = "http://localhost:3000"
token = "your-personal-access-token"

[defaults]
account = "ING Checking"
currency = "EUR"
```

## Output Formats

All commands support `--json` and `--toon` flags.

- **default**: human-readable table output
- **--json**: full JSON for programmatic consumption
- **--toon**: token-optimized notation for LLM agents (~40% fewer tokens than JSON)
