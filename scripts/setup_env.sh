#!/bin/bash

echo "Setting up environment variables for the language bot..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
# Database Configuration
PG_USER=postgres
PG_PASSWORD=pass

# Application Environment
ENV=development

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
EOF
    echo ".env file created. Please update TELEGRAM_BOT_TOKEN with your actual bot token."
else
    echo ".env file already exists."
fi

echo "Environment setup complete!"
echo "To start the application, run: docker-compose up" 