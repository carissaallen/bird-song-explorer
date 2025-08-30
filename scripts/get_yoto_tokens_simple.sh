#!/bin/bash

# Simple Yoto OAuth Token Script
# This version just exchanges a code you provide

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}   Simple Yoto Token Exchange        ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Load .env
if [ -f .env ]; then
    source .env
fi

# Set client ID
CLIENT_ID="${YOTO_CLIENT_ID:-qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU}"
REDIRECT_URI="https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"
TOKEN_URL="https://login.yotoplay.com/oauth/token"

echo -e "${YELLOW}Step 1: Get your authorization code${NC}"
echo ""
echo "Visit this URL in your browser:"
echo ""

# Generate a simple state
STATE="state123"
# URL-encode the redirect URI
ENCODED_REDIRECT_URI=$(echo -n "$REDIRECT_URI" | sed 's/:/%3A/g' | sed 's/\//%2F/g')
AUTH_URL="https://login.yotoplay.com/oauth/authorize?client_id=${CLIENT_ID}&redirect_uri=${ENCODED_REDIRECT_URI}&response_type=code&state=${STATE}&scope=offline_access"

echo -e "${GREEN}${AUTH_URL}${NC}"
echo ""
echo "After authorizing, you'll be redirected to:"
echo "https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook?code=XXXXX&state=XXXXX"
echo ""
echo "The page will show an error (that's OK!) - look at the URL bar"
echo "Copy JUST the code value (everything between 'code=' and '&state=')"
echo ""
read -p "Paste your authorization code here: " CODE

if [ -z "$CODE" ]; then
    echo -e "${RED}Error: No code provided${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 2: Exchanging code for tokens...${NC}"
echo ""

# Exchange the code
echo "Sending request to Yoto..."

RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=authorization_code" \
    -d "client_id=${CLIENT_ID}" \
    -d "code=${CODE}" \
    -d "redirect_uri=${REDIRECT_URI}" 2>&1)

echo "Response received:"
echo "$RESPONSE" | head -200

# Check for access_token
if echo "$RESPONSE" | grep -q "access_token"; then
    # Extract tokens using sed
    ACCESS_TOKEN=$(echo "$RESPONSE" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
    REFRESH_TOKEN=$(echo "$RESPONSE" | sed -n 's/.*"refresh_token":"\([^"]*\)".*/\1/p')
    
    if [ -z "$ACCESS_TOKEN" ]; then
        echo -e "${RED}Could not extract access token${NC}"
        exit 1
    fi
    
    echo ""
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
    
    # Remove old tokens
    grep -v "^YOTO_ACCESS_TOKEN=" .env > .env.tmp 2>/dev/null || true
    grep -v "^YOTO_REFRESH_TOKEN=" .env.tmp > .env.new 2>/dev/null || true
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
    echo "gcloud run services update bird-song-explorer \\"
    echo "  --region us-central1 \\"
    echo "  --update-env-vars \\"
    echo "    \"YOTO_ACCESS_TOKEN=${ACCESS_TOKEN},\\"
    echo "YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}\" \\"
    echo "  --quiet"
    echo ""
    
else
    echo -e "${RED}Error: Failed to get tokens${NC}"
    echo "Make sure you:"
    echo "1. Copied the code correctly (just the code, not the whole URL)"
    echo "2. Pasted it quickly (codes expire in ~30 seconds)"
    echo "3. Haven't used this code before (codes are single-use)"
fi