#!/bin/bash

# Browser-Based Authentication for Yoto with PKCE
# This uses the OAuth2 Authorization Code flow with PKCE

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}   Yoto Browser Authentication (PKCE)${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Load .env if exists
if [ -f .env ]; then
    source .env
fi

# Set client ID
CLIENT_ID="${YOTO_CLIENT_ID:v72VWYW4o59HgXrgAdiPPQCww1NJwJBS}"

echo -e "${YELLOW}Using Client ID:${NC} $CLIENT_ID"
echo ""

# Check if Go tool exists
if [ ! -f "../cmd/browser_auth/main.go" ]; then
    echo -e "${RED}Error: browser_auth tool not found${NC}"
    echo "Expected at: ../cmd/browser_auth/main.go"
    exit 1
fi

echo -e "${GREEN}Starting browser-based authentication...${NC}"
echo "This will:"
echo "1. Open your browser for Yoto login"
echo "2. Start a local server on port 8081 for callback"
echo "3. Exchange the authorization code for tokens"
echo "4. Update your .env file with new tokens"
echo ""

# Run the browser auth
cd ../cmd/browser_auth && go run main.go

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Authentication successful!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. The tokens have been saved to your .env file"
    echo "2. Deploy to Cloud Run using the command shown above"
else
    echo -e "${RED}❌ Authentication failed${NC}"
    exit 1
fi
