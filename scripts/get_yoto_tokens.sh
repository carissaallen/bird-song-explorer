#!/bin/bash

# Yoto OAuth Token Generation Script
# This script helps you get access and refresh tokens for the Yoto API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}     Yoto OAuth Token Generator      ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    echo "Please create a .env file with your Yoto credentials"
    exit 1
fi

# Load environment variables
source .env

# Check required variables
if [ -z "$YOTO_CLIENT_ID" ]; then
    echo -e "${RED}Error: YOTO_CLIENT_ID must be set in .env${NC}"
    exit 1
fi

# Yoto OAuth endpoints
AUTH_URL="https://login.yotoplay.com/oauth/authorize"
TOKEN_URL="https://login.yotoplay.com/oauth/token"
REDIRECT_URI="http://localhost:8080/callback"

echo -e "${YELLOW}Step 1: Generate Authorization URL${NC}"
echo ""
echo "You need to visit this URL in your browser to authorize the application:"
echo ""

# Generate state for security
STATE=$(openssl rand -hex 16)

# Build authorization URL
AUTH_FULL_URL="${AUTH_URL}?client_id=${YOTO_CLIENT_ID}&redirect_uri=${REDIRECT_URI}&response_type=code&state=${STATE}&scope=offline_access"

echo -e "${GREEN}${AUTH_FULL_URL}${NC}"
echo ""
echo -e "${YELLOW}Step 2: Get Authorization Code${NC}"
echo ""
echo "1. Open the URL above in your browser"
echo "2. Log in to your Yoto account"
echo "3. Authorize the application"
echo "4. You'll be redirected to: http://localhost:8080/callback?code=XXXXX&state=XXXXX"
echo "5. Copy the 'code' parameter from the URL"
echo ""
read -p "Enter the authorization code: " AUTH_CODE

if [ -z "$AUTH_CODE" ]; then
    echo -e "${RED}Error: Authorization code cannot be empty${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 3: Exchange Code for Tokens${NC}"
echo ""

# Exchange authorization code for tokens
# Build the request - include client_secret only if it exists
TOKEN_REQUEST="grant_type=authorization_code&client_id=${YOTO_CLIENT_ID}&code=${AUTH_CODE}&redirect_uri=${REDIRECT_URI}"
if [ ! -z "$YOTO_CLIENT_SECRET" ]; then
    TOKEN_REQUEST="${TOKEN_REQUEST}&client_secret=${YOTO_CLIENT_SECRET}"
fi

RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "${TOKEN_REQUEST}")

# Check if response contains access_token
if echo "$RESPONSE" | grep -q "access_token"; then
    # Extract tokens
    ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$REFRESH_TOKEN" ]; then
        echo -e "${RED}Error: Failed to extract tokens from response${NC}"
        echo "Response: $RESPONSE"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Successfully obtained tokens!${NC}"
    echo ""
    echo -e "${BLUE}Access Token:${NC}"
    echo "$ACCESS_TOKEN" | head -c 50
    echo "..."
    echo ""
    echo -e "${BLUE}Refresh Token:${NC}"
    echo "$REFRESH_TOKEN" | head -c 50
    echo "..."
    echo ""
    
    # Update .env file
    echo -e "${YELLOW}Step 4: Update Configuration${NC}"
    echo ""
    read -p "Do you want to update your .env file with these tokens? (y/n): " UPDATE_ENV
    
    if [ "$UPDATE_ENV" = "y" ] || [ "$UPDATE_ENV" = "Y" ]; then
        # Remove old tokens if they exist
        grep -v "^YOTO_ACCESS_TOKEN=" .env > .env.tmp || true
        grep -v "^YOTO_REFRESH_TOKEN=" .env.tmp > .env.tmp2 || true
        mv .env.tmp2 .env
        rm -f .env.tmp
        
        # Add new tokens
        echo "" >> .env
        echo "# Yoto OAuth Tokens (generated $(date))" >> .env
        echo "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN}" >> .env
        echo "YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}" >> .env
        
        echo -e "${GREEN}✓ Updated .env file${NC}"
    fi
    
    echo ""
    echo -e "${YELLOW}Step 5: Update Google Cloud Run${NC}"
    echo ""
    echo "Run these commands to update your Cloud Run service:"
    echo ""
    echo -e "${BLUE}# Option 1: Update just the tokens${NC}"
    cat << EOF
gcloud run services update bird-song-explorer \\
  --region us-central1 \\
  --update-env-vars \\
    "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN},\\
YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}" \\
  --quiet
EOF
    
    echo ""
    echo -e "${BLUE}# Option 2: Update all environment variables from .env${NC}"
    cat << 'EOF'
source .env && gcloud run services update bird-song-explorer \
  --region us-central1 \
  --update-env-vars \
    "YOTO_ACCESS_TOKEN=${YOTO_ACCESS_TOKEN},\
YOTO_REFRESH_TOKEN=${YOTO_REFRESH_TOKEN},\
YOTO_CLIENT_ID=${YOTO_CLIENT_ID},\
YOTO_CLIENT_SECRET=${YOTO_CLIENT_SECRET},\
YOTO_API_BASE_URL=${YOTO_API_BASE_URL},\
YOTO_CARD_ID=${YOTO_CARD_ID},\
EBIRD_API_KEY=${EBIRD_API_KEY},\
XENOCANTO_API_KEY=${XENOCANTO_API_KEY},\
ELEVENLABS_API_KEY=${ELEVENLABS_API_KEY}" \
  --quiet
EOF
    
    echo ""
    echo -e "${GREEN}✓ Token generation complete!${NC}"
    echo ""
    echo -e "${YELLOW}Note:${NC} Access tokens expire after a period. Use the refresh token to get new access tokens when needed."
    
else
    echo -e "${RED}Error: Failed to get tokens${NC}"
    echo "Response from Yoto API:"
    echo "$RESPONSE"
    echo ""
    echo "Common issues:"
    echo "1. Authorization code has already been used"
    echo "2. Authorization code has expired (they expire quickly)"
    echo "3. Client credentials are incorrect"
    echo "4. Redirect URI doesn't match what's configured in Yoto"
    exit 1
fi