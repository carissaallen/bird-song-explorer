#!/bin/bash

# Yoto Token Refresh Script
# This script refreshes your Yoto access token using the refresh token

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}     Yoto Token Refresh Script       ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    exit 1
fi

# Load environment variables
source .env

# Check required variables
if [ -z "$YOTO_CLIENT_ID" ] || [ -z "$YOTO_REFRESH_TOKEN" ]; then
    echo -e "${RED}Error: YOTO_CLIENT_ID and YOTO_REFRESH_TOKEN must be set in .env${NC}"
    echo "Run ./scripts/get_yoto_tokens.sh first to get initial tokens"
    exit 1
fi

TOKEN_URL="https://login.yotoplay.com/oauth/token"

echo -e "${YELLOW}Refreshing access token...${NC}"
echo ""

# Refresh the token
REFRESH_REQUEST="grant_type=refresh_token&client_id=${YOTO_CLIENT_ID}&refresh_token=${YOTO_REFRESH_TOKEN}"

RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "${REFRESH_REQUEST}")

# Check if response contains access_token
if echo "$RESPONSE" | grep -q "access_token"; then
    # Extract new tokens
    NEW_ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    NEW_REFRESH_TOKEN=$(echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    
    # If no new refresh token provided, keep the old one
    if [ -z "$NEW_REFRESH_TOKEN" ]; then
        NEW_REFRESH_TOKEN=$YOTO_REFRESH_TOKEN
    fi
    
    echo -e "${GREEN}✓ Successfully refreshed tokens!${NC}"
    echo ""
    
    # Update .env file
    echo -e "${YELLOW}Updating .env file...${NC}"
    
    # Create backup
    cp .env .env.backup
    
    # Remove old tokens
    grep -v "^YOTO_ACCESS_TOKEN=" .env > .env.tmp || true
    grep -v "^YOTO_REFRESH_TOKEN=" .env.tmp > .env.tmp2 || true
    mv .env.tmp2 .env
    rm -f .env.tmp
    
    # Add new tokens
    echo "" >> .env
    echo "# Yoto OAuth Tokens (refreshed $(date))" >> .env
    echo "YOTO_ACCESS_TOKEN=${NEW_ACCESS_TOKEN}" >> .env
    echo "YOTO_REFRESH_TOKEN=${NEW_REFRESH_TOKEN}" >> .env
    
    echo -e "${GREEN}✓ Updated .env file${NC}"
    echo ""
    
    # Show update command for Cloud Run
    echo -e "${YELLOW}Update Google Cloud Run:${NC}"
    echo ""
    cat << EOF
gcloud run services update bird-song-explorer \\
  --region us-central1 \\
  --update-env-vars \\
    "YOTO_ACCESS_TOKEN=${NEW_ACCESS_TOKEN},\\
YOTO_REFRESH_TOKEN=${NEW_REFRESH_TOKEN}" \\
  --quiet
EOF
    
    echo ""
    echo -e "${GREEN}✓ Token refresh complete!${NC}"
    
else
    echo -e "${RED}Error: Failed to refresh token${NC}"
    echo "Response from Yoto API:"
    echo "$RESPONSE"
    echo ""
    echo "You may need to get new tokens using ./scripts/get_yoto_tokens.sh"
    exit 1
fi