#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Warning: .env file not found!"
    echo "Please ensure environment variables are set."
fi

# Verify required variables
if [ -z "$EBIRD_API_KEY" ]; then
    echo "Error: EBIRD_API_KEY is not set!"
    exit 1
fi

if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo "Error: ELEVENLABS_API_KEY is not set!"
    exit 1
fi

echo "✅ All required API keys found"
echo "🚀 Starting deployment..."

# Run the actual deployment script
./deploy/gcp-cloud-run.sh