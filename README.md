# León Speedcams
![Version](https://img.shields.io/badge/Version-1.0.0-brightgreen.svg)
![Go](https://img.shields.io/badge/Go-1.22-brightgreen.svg)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

León Speedcams is a Go service that checks for new speed camera reports from [ahoraleon.com](https://www.ahoraleon.com) and sends alerts via Telegram for specified streets. The service is designed to run on a schedule using Docker Compose or other schedulers.

## Table of Contents
- [Features](#features)
- [Requirements](#requirements)
- [Configuration](#configuration)
- [Development Setup](#development-setup)
- [Docker](#docker)
- [Continuous Integration](#continuous-integration)
- [Contributing](#contributing)
- [License](#license)

## Features
- Scrapes speed camera data from ahoraleon.com.
- Sends Telegram messages when speed cameras are active on monitored streets.
- Configurable via environment variables.
- Lightweight Docker image suitable for scheduled runs.

## Requirements
- Go 1.24 or later.
- Docker for dev container and/or production deployment.
- Telegram Bot with valid token and chat ID: [Create a Bot](https://core.telegram.org/bots#6-botfather).

## Configuration
Create a `.env` file in the project root with the following variables:

```dotenv
REQUEST_TIMEOUT="30"
BASE_REQUEST_URL="https://www.ahoraleon.com"
MONITORED_STREETS="Street1,Street2,Street3"
TELEGRAM_BOT_TOKEN="1234567890:ABCDEF..."
TELEGRAM_CHAT_ID="123456789"
```

## Development Setup
1. Install Go 1.24 or later.
2. Clone the repository:
   ```bash
   git clone https://github.com/dmarts05/leon-speedcams-go.git
   cd leon-speedcams-go
   ```
3. Run:
   ```bash
   make run
   ```
4. Run tests:
   ```bash
   make test
   ```
5. Run tests with coverage:
   ```bash
   make test/cover
   ```

## Docker
Build and run the Docker image:

```bash
docker build -t leon-speedcams-go .
docker run --rm --env-file .env leon-speedcams-go
```

## Continuous Integration
CI is configured with GitHub Actions:
- Lints code with `golangci-lint`.
- Runs tests and builds binaries.
- Publishes Docker image on successful main branch builds.

See `.github/workflows` for details.

## Contributing
Contributions are welcome. Please open an issue or submit a pull request.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.