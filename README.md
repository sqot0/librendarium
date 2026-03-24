# Librendarium

Librendarium is a powerful tool to synchronize homework assignments from Librus Synergia to Google Calendar. It keeps your study schedule up-to-date by automatically creating, updating (skipping if exists), and cleaning up stale events.

## Features

- **Sync Homework**: One-way sync from Librus to Google Calendar.
- **Concurrent Execution**: Uses Go's concurrency primitives (`errgroup`) to fetch subject and category details in parallel for faster sync.
- **Deduplication**: Intelligently identifies existing events to prevent duplicates.
- **Cleanup**: Automatically removes future calendar events that have been deleted or modified in Librus.
- **Timezone Aware**: Specifically handles Polish educational schedules (Europe/Warsaw timezone).

## Setup

### Prerequisites

1.  **Librus Account**: Standard Synergia credentials.
2.  **Google Cloud Service Account**:
    - Create a project in [Google Cloud Console](https://console.cloud.google.com/).
    - Enable **Google Calendar API**.
    - Create a **Service Account** and download its JSON credentials or copy the private key.
    - Grant "Make changes to events" permission to the Service Account email in your Google Calendar settings.

### Configuration

Create a `.env` file in the root directory:

```env
LIBRUS_LOGIN=your_login
LIBRUS_PASSWORD=your_password
GOOGLE_CALENDAR_ID=your_calendar_id@group.calendar.google.com
SYNC_INTERVAL=1h # Optional: Set to run as a daemon (e.g., 30m, 1h, 24h)
GOOGLE_CALENDAR_CLIENT_EMAIL=service-account@project.iam.gserviceaccount.com
GOOGLE_CALENDAR_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n..."
```

## Running

### Locally

```bash
go run ./cmd/librendarium
```

### Docker

```bash
docker build -t librendarium .
docker run --env-file .env librendarium
```

## Contributing

- Follow standard Go idioms.
- Use `context` for timeouts and cancellation.
- Keep abstractions to a minimum for better maintainability.

## License

MIT
