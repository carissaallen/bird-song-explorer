#!/bin/bash

# Get new Yoto tokens using the confidential client

echo "🔐 Getting new Yoto tokens for confidential client"
echo "============================================"

# Set the new client credentials
export YOTO_CLIENT_ID="a0PBKaHZnoB02Jpe5RZ9znZV5e1OXPr7"
export YOTO_CLIENT_SECRET="7EU77IABOjxVDfl666QeC9uxO9Vt_FZvAEfFhNZ1TBXEC81P9XtPa7RInzaIiPZt"

echo "Client ID: ${YOTO_CLIENT_ID}"
echo "Client Secret: ${YOTO_CLIENT_SECRET:0:20}..."
echo ""

# Check if Go tool exists
if [ ! -f "../cmd/get_yoto_token/main.go" ]; then
    echo "❌ Error: get_yoto_token tool not found"
    echo "Expected at: ../cmd/get_yoto_token/main.go"
    exit 1
fi

# Run the OAuth flow
echo "📱 Starting OAuth flow..."
echo "This will open your browser for authorization"
echo ""

cd ../cmd/get_yoto_token && go run main.go

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Tokens obtained successfully!"
    echo ""
    echo "Next steps:"
    echo "1. The tokens have been saved to your .env file"
    echo "2. Run: ./update_yoto_credentials.sh"
    echo "   to update Cloud Run with the new credentials"
else
    echo "❌ Failed to obtain tokens"
    exit 1
fi
