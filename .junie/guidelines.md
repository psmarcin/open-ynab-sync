# Project Guidelines for Open YNAB Sync

## Project Overview

Open YNAB Sync is a Go application that automatically synchronizes transactions between GoCardless (a banking data provider) and YNAB (You Need A Budget). The application fetches transactions from GoCardless accounts and uploads them to corresponding YNAB accounts on a scheduled basis.

### Key Features

- Automatic transaction synchronization between GoCardless and YNAB
- Support for multiple synchronization jobs (multiple accounts)
- Scheduled execution using cron expressions
- Rate limiting handling for API requests
- Optional New Relic integration for monitoring

## Project Structure

The project is organized as follows:

- **Root Directory**: Contains the main application code
  - `main.go`: Entry point and scheduler setup
  - `job.go`: Job configuration and parsing
  - `gocardless.go`: GoCardless API integration
  - `ynab.go`: YNAB API integration
  - `*_test.go`: Test files for the corresponding Go files

- **cmd/link/**: Contains a utility for linking GoCardless accounts
  - `main.go`: Command-line tool for linking bank accounts with GoCardless

- **requests/**: Contains HTTP request templates for testing
  - `token.http`: HTTP requests for token-related operations

## Development Guidelines

### Building the Project

The project can be built using standard Go commands:

```bash
go build -o open-ynab-sync
```

### Running Tests

Tests can be run using the standard Go test command:

```bash
go test ./...
```

The tests use mocking for external dependencies, particularly for HTTP requests to the GoCardless and YNAB APIs.

### Code Style

The project follows standard Go code style conventions:

- Use `gofmt` or `go fmt` to format code
- Follow Go naming conventions (camelCase for unexported, PascalCase for exported)
- Use meaningful variable and function names
- Include comments for exported functions and types

### Error Handling

The project uses the following approach for error handling:

- Use the `errors` package for wrapping errors with context
- Log errors with appropriate context information
- Return errors to the caller when appropriate
- Use structured logging with `slog` package

### Configuration

The application is configured using environment variables:

- `GC_SECRET_ID`: GoCardless API Secret ID
- `GC_SECRET_KEY`: GoCardless API Secret Key
- `YNAB_TOKEN`: YNAB Personal Access Token
- `JOBS`: Configuration for synchronization jobs
- `CRON_SCHEDULE`: Cron schedule for synchronization

### Deployment

The application can be deployed using Docker:

```bash
docker build -t open-ynab-sync .
docker run -d \
  -e GC_SECRET_ID=your_gocardless_secret_id \
  -e GC_SECRET_KEY=your_gocardless_secret_key \
  -e YNAB_TOKEN=your_ynab_token \
  -e JOBS=your_gocardless_account_id,your_ynab_budget_id,your_ynab_account_id \
  -e CRON_SCHEDULE="0 6,18 * * *" \
  --name open-ynab-sync \
  open-ynab-sync
```

Alternatively, Docker Compose can be used for easier deployment:

```bash
docker-compose up -d
```

## Contributing Guidelines

When contributing to this project, please follow these guidelines:

1. Write tests for new functionality
2. Ensure all tests pass before submitting a pull request
3. Follow the existing code style and error handling patterns
4. Update documentation as needed
5. Use meaningful commit messages

## Troubleshooting

Common issues and their solutions:

1. **Authentication failures**: Ensure that the GoCardless and YNAB credentials are correct and have the necessary permissions.
2. **Rate limiting**: The application handles rate limiting from the GoCardless API, but you may need to adjust your synchronization schedule if you're consistently hitting limits.
3. **Missing transactions**: Check that the date range for transaction fetching (currently 20 days in the past) is appropriate for your needs.