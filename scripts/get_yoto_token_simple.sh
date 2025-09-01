#!/bin/bash

# Simple Yoto OAuth Token Exchange
# Just paste the authorization code, not the full URL

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}   Yoto OAuth Token Exchange         ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Load .env
if [ -f .env ]; then
    source .env
fi

# Set variables
CLIENT_ID="${YOTO_CLIENT_ID:-qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU}"
REDIRECT_URI="https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"
TOKEN_URL="https://login.yotoplay.com/oauth/token"

echo -e "${YELLOW}Step 1: Get Authorization Code${NC}"
echo ""
echo "Open this URL in your browser:"
echo ""

# Generate authorization URL
STATE="state123"
ENCODED_REDIRECT_URI=$(printf '%s' "$REDIRECT_URI" | jq -sRr @uri)
AUTH_URL="https://login.yotoplay.com/oauth/authorize?client_id=${CLIENT_ID}&redirect_uri=${ENCODED_REDIRECT_URI}&response_type=code&state=${STATE}&scope=offline_access+openid+profile"

echo -e "${GREEN}${AUTH_URL}${NC}"
echo ""
echo "After authorizing, look at the URL in your browser."
echo "It will be something like:"
echo "https://yoto-bird-song-explorer...webhook?code=ABC123&state=..."
echo ""
echo -e "${YELLOW}Copy ONLY the code value (the part after 'code=' and before '&')${NC}"
echo ""
read -p "Paste just the authorization code: " CODE

if [ -z "$CODE" ]; then
    echo -e "${RED}Error: No code provided${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 2: Exchanging code for tokens...${NC}"
echo ""

# Exchange the code - don't include client_secret for public clients
REQUEST_BODY="grant_type=authorization_code&client_id=${CLIENT_ID}&code=${CODE}&redirect_uri=${REDIRECT_URI}"

RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "${REQUEST_BODY}" 2>&1)

echo "Response from Yoto:"
echo "$RESPONSE" | python -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

# Check for access_token
if echo "$RESPONSE" | grep -q "access_token"; then
    # Extract tokens
    ACCESS_TOKEN=$(echo "$RESPONSE" | python -c "import sys, json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null || echo "")
    REFRESH_TOKEN=$(echo "$RESPONSE" | python -c "import sys, json; print(json.load(sys.stdin)['refresh_token'])" 2>/dev/null || echo "")
    
    if [ -z "$ACCESS_TOKEN" ]; then
        # Fallback to sed
        ACCESS_TOKEN=$(echo "$RESPONSE" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
        REFRESH_TOKEN=$(echo "$RESPONSE" | sed -n 's/.*"refresh_token":"\([^"]*\)".*/\1/p')
    fi
    
    if [ -z "$ACCESS_TOKEN" ]; then
        echo -e "${RED}Could not extract tokens${NC}"
        echo "Please copy them manually from the response above"
        exit 1
    fi
    
    echo -e "${GREEN}✅ SUCCESS! Got tokens!${NC}"
    echo ""
    echo "Access Token (first 50 chars):"
    echo "$ACCESS_TOKEN" | cut -c1-50
    echo ""
    echo "Refresh Token (first 50 chars):"
    echo "$REFRESH_TOKEN" | cut -c1-50
    echo ""
    
    # Update .env
    echo -e "${YELLOW}Updating .env file...${NC}"
    
    # Backup
    cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    
    # Remove old tokens
    grep -v "^YOTO_ACCESS_TOKEN=" .env > .env.tmp 2>/dev/null || true
    grep -v "^YOTO_REFRESH_TOKEN=" .env.tmp > .env.new 2>/dev/null || cp .env .env.new
    mv .env.new .env
    rm -f .env.tmp
    
    # Add new tokens
    echo "" >> .env
    echo "# Yoto Tokens (updated $(date))" >> .env
    echo "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN}" >> .env
    echo "YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}" >> .env
    
    echo -e "${GREEN}✓ Updated .env file${NC}"
    echo ""
    
    # Show Cloud Run command
    echo -e "${BLUE}Update Cloud Run with this command:${NC}"
    echo ""
    cat << EOF
gcloud run services update bird-song-explorer \\
  --region us-central1 \\
  --update-env-vars \\
    "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN},\\
YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}" \\
  --quiet
EOF
    
else
    echo -e "${RED}Error: Failed to get tokens${NC}"
    echo ""
    echo "Common issues:"
    echo "1. Code expired (they expire in ~30 seconds)"
    echo "2. Code already used (each code can only be used once)"
    echo "3. Wrong code copied"
    echo ""
    echo "Try again with a fresh authorization code"
fi