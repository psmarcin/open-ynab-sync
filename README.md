# Open YNAB Sync

A Go application that synchronizes transactions between GoCardless and YNAB (You Need A Budget).

## Features

- Automatically fetches transactions from GoCardless
- Synchronizes transactions to YNAB
- Runs on a schedule (every minute by default)
- Handles rate limiting and authentication

## Prerequisites

- Go 1.24.4 or later (for building from source)
- Docker (optional, for containerized deployment)
- GoCardless account with API access
- YNAB account with personal access token

## Installation

### Using Docker

1. Pull the Docker image or build it yourself:

```bash
# Build the Docker image
docker build -t open-ynab-sync .
```

2. Run the container with the required environment variables:

```bash
docker run -d \
  -e GC_SECRET_ID=your_gocardless_secret_id \
  -e GC_SECRET_KEY=your_gocardless_secret_key \
  -e YNAB_TOKEN=your_ynab_token \
  -e JOBS=your_gocardless_account_id,your_ynab_budget_id,your_ynab_account_id \
  -e CRON_SCHEDULE="0 6,18 * * *" \
  --name open-ynab-sync \
  open-ynab-sync
```

### Using Docker Compose (Recommended)

1. Copy the example environment file and edit it with your credentials:

```bash
cp .env.example .env
# Edit .env with your actual credentials
```

2. Start the application using Docker Compose:

```bash
docker-compose up -d
```

3. View the logs:

```bash
docker-compose logs -f
```

4. Stop the application:

```bash
docker-compose down
```

The Docker Compose setup includes:
- Automatic container restart if it crashes
- Log rotation to prevent disk space issues
- Health checks to monitor the application status

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/psmarcin/open-ynab-sync.git
cd open-ynab-sync
```

2. Build the application:

```bash
go build -o open-ynab-sync
```

3. Run the application with the required environment variables:

```bash
GC_SECRET_ID=your_gocardless_secret_id \
GC_SECRET_KEY=your_gocardless_secret_key \
YNAB_TOKEN=your_ynab_token \
JOBS=your_gocardless_account_id,your_ynab_budget_id,your_ynab_account_id \
CRON_SCHEDULE="0 6,18 * * *" \
./open-ynab-sync
```

## Configuration

The application requires the following environment variables:

| Variable | Description |
|----------|-------------|
| `GC_SECRET_ID` | GoCardless API Secret ID |
| `GC_SECRET_KEY` | GoCardless API Secret Key |
| `YNAB_TOKEN` | YNAB Personal Access Token |
| `JOBS` | Configuration for synchronization jobs (see below) |
| `CRON_SCHEDULE` | Cron schedule for synchronization (default: "0 6,18 * * *" - twice daily at 6am and 6pm) |
| `NEW_RELIC_LICENCE_KEY` | New Relic License Key (optional, for monitoring) |
| `NEW_RELIC_USER_KEY` | New Relic User Key (optional, for monitoring) |
| `NEW_RELIC_APP_NAME` | New Relic Application Name (optional, for monitoring) |

### Jobs Configuration

The `JOBS` environment variable allows you to configure multiple synchronization jobs. Each job synchronizes transactions from a specific GoCardless account to a specific YNAB account.

Format: `GCAccountID1,YNABBudgetID1,YNABAccountID1|GCAccountID2,YNABBudgetID2,YNABAccountID2|...`

Each job consists of three parts separated by commas:
1. GoCardless Account ID
2. YNAB Budget ID
3. YNAB Account ID

Multiple jobs are separated by the pipe character (`|`).

Example:
```
JOBS=gc_acc_123456,ynab_budget_abc123,ynab_account_def456|gc_acc_789012,ynab_budget_ghi789,ynab_account_jkl012
```

This configures two synchronization jobs:
1. From GoCardless account `gc_acc_123456` to YNAB account `ynab_account_def456` in budget `ynab_budget_abc123`
2. From GoCardless account `gc_acc_789012` to YNAB account `ynab_account_jkl012` in budget `ynab_budget_ghi789`

### Getting GoCardless Credentials

1. Sign up for a GoCardless developer account at [GoCardless Developer Portal](https://bankaccountdata.gocardless.com/)
2. Create an application to get your Secret ID and Secret Key
3. Find your Account ID in your GoCardless dashboard

### Getting YNAB Credentials

1. Log in to your YNAB account
2. Go to Account Settings > Developer Settings
3. Generate a new Personal Access Token
4. Find your Account ID in the URL when viewing your budget account

## How It Works

1. The application authenticates with GoCardless using your Secret ID and Secret Key
2. It fetches transactions from the past 20 days from your GoCardless account
3. It converts these transactions to YNAB format
4. It uploads the transactions to your YNAB account
5. This process repeats according to your CRON_SCHEDULE (default: twice daily at 6am and 6pm)
6. If New Relic monitoring is configured, performance metrics and logs are sent to New Relic

## Development

### Running Tests

```bash
go test ./...
```

### Continuous Integration

This project uses GitHub Actions for continuous integration. The workflow automatically runs all tests on every push to the `main` branch and on pull requests to the `main` branch.

You can see the workflow configuration in `.github/workflows/go-tests.yml`.

### Project Structure

- `main.go` - Entry point and scheduler setup
- `job.go` - Job configuration and parsing
- `gocardless.go` - GoCardless API integration
- `ynab.go` - YNAB API integration
- `Dockerfile` - Container definition
- `docker-compose.yml` - Docker Compose configuration for easy deployment
- `.env.example` - Example environment variables file
- `.github/workflows/` - GitHub Actions workflow configurations
  - `go-tests.yml` - CI workflow for running tests

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
