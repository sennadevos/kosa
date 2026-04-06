# Finance

Personal kosaance tracking system backed by a no-code database (visual UI) with full API access for CLI tools, agents, and automation.

## Goals

- Track all money movement: day-to-day spending, recurring subscriptions, loans
- Calculate expected balance on any future date from data alone
- Fast data entry: adding a spend or loan should take seconds
- Full programmatic access: CLI, agents, and automation via REST API
- Visual dashboard via the database UI (no separate frontend needed)

## Design Principles

1. **Single source of truth** — all kosaancial data lives in the database; no spreadsheets, no local files
2. **Calculable** — expected future balance = current balance + sum(future income) - sum(future expenses) - sum(loan payments due)
3. **Low friction** — entering a transaction should be one CLI command or one row in the UI
4. **Automatable** — recurring transactions are generated from schedules, not entered manually each time

## Tech Stack

- **Database + UI**: Teable (PostgreSQL-backed, Airtable-like UI, REST API)
- **CLI tool**: Custom Go CLI for quick data entry and queries
- **Agents**: Claude Code / custom agents that can read and write kosaancial data
- **Automation**: Scheduled jobs for balance alerts, reconciliation prompts

The CLI uses a database abstraction layer so the backend can be swapped (e.g., to Directus, Mathesar, or raw Postgres) without changing application logic.
