#!/bin/bash

# Check for .env file
if [ -f .env ]; then
    source .env
    echo "Loaded .env file"
    echo "DATABASE_URL is set to: $DATABASE_URL"
else
    echo "Error: .env file not found in $(pwd)"
    exit 1
fi

# Ensure variables are exported to subprocesses
export DATABASE_URL
export PORT
export SESSION_KEY

# Run the migration
go run migrate/main.go