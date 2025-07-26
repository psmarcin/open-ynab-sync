# GoCardless Account Linking Tool

This command-line tool helps users link their bank accounts with GoCardless for use with Open YNAB Sync.

## Overview

The tool guides users through the GoCardless authorization flow:

1. Creates an agreement with the specified institution
2. Creates a requisition for account access
3. Opens a browser for the user to authorize access
4. Waits for the callback from GoCardless
5. Displays the linked account IDs

## Usage

```bash
go run main.go -institution=INSTITUTION_ID [-port=8080] [-auth-timeout=5m] [-http-timeout=20s]
```

### Required Environment Variables

- `GC_SECRET_ID`: GoCardless API Secret ID
- `GC_SECRET_KEY`: GoCardless API Secret Key

These can be set in a `.env` file in the parent directory.

### Command-line Arguments

- `-institution`: Institution ID (required)
- `-port`: Port to listen for callback (default: 8080)
- `-auth-timeout`: Timeout for waiting for authorization callback (default: 5 minutes)
- `-http-timeout`: Timeout for HTTP requests (default: 20 seconds)

## Architecture

The application is organized into the following components:

### Configuration (`config/config.go`)

Handles loading configuration from environment variables and command-line flags.

- Loads `.env` file
- Parses command-line arguments
- Validates required configuration values
- Provides sensible defaults

### GoCardless API Client (`api/gocardless.go`)

Handles communication with the GoCardless API.

- Authentication
- Creating agreements and requisitions
- Checking requisition status
- Listing requisitions

### HTTP Callback Server (`server/callback.go`)

Handles the HTTP server for receiving the OAuth callback from GoCardless.

- Starts and stops the server gracefully
- Handles the callback request
- Provides a channel for signaling when the callback is received

### Authorization Flow (`auth/flow.go`)

Orchestrates the authorization flow.

- Creates agreements and requisitions
- Opens the browser for user authorization
- Waits for the callback
- Checks the requisition status

### Main Application (`main.go`)

Orchestrates the components.

- Initializes the logger
- Sets up signal handling for graceful shutdown
- Loads configuration
- Creates and connects the components
- Executes the authorization flow
- Displays the results

## Improvements

This refactored implementation includes several improvements over the original:

1. **Separation of Concerns**: Each component has a single responsibility, making the code more maintainable.
2. **Improved Error Handling**: Consistent error handling with proper context using the `errors` package.
3. **Graceful Shutdown**: The application handles signals for graceful shutdown.
4. **Structured Logging**: Uses `slog` for structured logging with context.
5. **Testability**: Components are designed with interfaces for better testability.
6. **Configuration Flexibility**: Enhanced configuration options with sensible defaults.
7. **Code Reusability**: Components can be reused in other parts of the application.
8. **Comprehensive Tests**: Each component has thorough tests to ensure it works correctly.