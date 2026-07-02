# Language Learning Bot

A Telegram bot for language learning with word management, translation, and testing features.

## Setup

### 1. Environment Setup

Run the setup script to create the necessary environment file:

```bash
./scripts/setup_env.sh
```

This will create a `.env` file with default values. You need to update the `TELEGRAM_BOT_TOKEN` with your actual bot token.

### 2. Database Configuration

The application uses PostgreSQL. The default configuration is:
- Host: `db` (Docker service name)
- User: `postgres`
- Password: `pass`
- Database: `my_database`
- Port: `5432`

### 3. Running the Application

#### Using Docker Compose (Recommended)

```bash
docker-compose up
```

#### Manual Setup

1. Start PostgreSQL:
```bash
docker run -d \
  --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=pass \
  -e POSTGRES_DB=my_database \
  -p 5432:5432 \
  postgres:15
```

2. Set environment variables:
```bash
export PG_USER=postgres
export PG_PASSWORD=pass
export ENV=development
export TELEGRAM_BOT_TOKEN=your_bot_token_here
```

3. Run the application:
```bash
cd app
go run ./cmd/langbot
```

## Features

- Word management and translation
- Language learning tests
- Book reading functionality
- Custom translation scenarios
- Scheduled word testing (respects per-user quiet hours and timezone)
- Quiet hours: users can configure a time window and days of week when the bot must not send messages (⚙️ Settings)
- German article detection (der/die/das) when saving nouns
- Multi-language support (English, Ukrainian, German, Spanish)

## Troubleshooting

### Database Connection Issues

If you encounter database connection errors:

1. Ensure PostgreSQL is running:
```bash
docker ps
```

2. Check the database logs:
```bash
docker logs go-db
```

3. Verify environment variables are set correctly:
```bash
echo $PG_USER
echo $PG_PASSWORD
```

### Common Issues

- **Password authentication failed**: Ensure `PG_PASSWORD=pass` matches the PostgreSQL container password
- **Connection refused**: Make sure the PostgreSQL container is running and healthy
- **Bot token not set**: Update `TELEGRAM_BOT_TOKEN` in your `.env` file

## Development

The application uses:
- Go 1.24
- PostgreSQL 15
- GORM for database operations
- Telegram Bot API
- Docker for containerization 