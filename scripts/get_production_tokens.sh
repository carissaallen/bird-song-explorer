#!/bin/bash

# Yoto OAuth Token Generation Script for Production
# Uses the production callback URL configured in Yoto Developer Dashboard

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Yoto Production Token Generator    ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check if .env file exists
if [ ! -f ../.env ]; then
    echo -e "${RED}Error: ../.env file not found!${NC}"
    echo "Please ensure you're running this from the scripts directory"
    exit 1
fi

# Load environment variables
source ../.env

# Check required variables
if [ -z "$YOTO_CLIENT_ID" ]; then
    echo -e "${RED}Error: YOTO_CLIENT_ID must be set in .env${NC}"
    exit 1
fi

# Correct Yoto OAuth endpoints
AUTH_URL="https://login.yotoplay.com/oauth/authorize"
TOKEN_URL="https://login.yotoplay.com/oauth/token"
REDIRECT_URI="https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"

echo -e "${YELLOW}Step 1: Generate Authorization URL${NC}"
echo ""
echo "You need to visit this URL in your browser to authorize the application:"
echo ""

# Generate state for security
STATE=$(openssl rand -hex 16)

# Build authorization URL - use the old working client ID
YOTO_CLIENT_ID="qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU"
AUTH_FULL_URL="${AUTH_URL}?client_id=${YOTO_CLIENT_ID}&redirect_uri=${REDIRECT_URI}&response_type=code&state=${STATE}&scope=offline_access"

echo -e "${GREEN}${AUTH_FULL_URL}${NC}"
echo ""
echo -e "${YELLOW}Step 2: Get Authorization Code${NC}"
echo ""
echo "1. Open the URL above in your browser"
echo "2. Log in to your Yoto account"
echo "3. Authorize the application"
echo "4. You'll be redirected to the production Bird Song Explorer"
echo "5. The page will show you the tokens and Cloud Run update command"
echo ""
echo -e "${BLUE}Alternative: Manual Token Exchange${NC}"
echo ""
echo "If you want to manually exchange the code:"
echo "1. Copy the 'code' parameter from the redirect URL"
echo "2. Enter it below:"
echo ""
read -p "Enter the authorization code (or press Enter to skip): " AUTH_CODE

if [ ! -z "$AUTH_CODE" ]; then
    echo ""
    echo -e "${YELLOW}Step 3: Exchange Code for Tokens${NC}"
    echo ""
    
    # Exchange authorization code for tokens
    YOTO_CLIENT_ID="qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU"
    TOKEN_REQUEST="grant_type=authorization_code&client_id=${YOTO_CLIENT_ID}&code=${AUTH_CODE}&redirect_uri=${REDIRECT_URI}"
    
    RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "${TOKEN_REQUEST}")
    
    # Check if response contains access_token
    if echo "$RESPONSE" | grep -q "access_token"; then
        # Extract tokens using jq if available, otherwise use grep
        if command -v jq &> /dev/null; then
            ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
            REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')
        else
            ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
            REFRESH_TOKEN=$(echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
        fi
        
        if [ -z "$ACCESS_TOKEN" ] || [ -z "$REFRESH_TOKEN" ]; then
            echo -e "${RED}Error: Failed to extract tokens from response${NC}"
            echo "Response: $RESPONSE"
            exit 1
        fi
        
        echo -e "${GREEN}✓ Successfully obtained tokens!${NC}"
        echo ""
        echo -e "${BLUE}Access Token (first 50 chars):${NC}"
        echo "$ACCESS_TOKEN" | head -c 50
        echo "..."
        echo ""
        echo -e "${BLUE}Refresh Token (first 50 chars):${NC}"
        echo "$REFRESH_TOKEN" | head -c 50
        echo "..."
        echo ""
        
        # Update .env file
        echo -e "${YELLOW}Step 4: Update Configuration${NC}"
        echo ""
        read -p "Do you want to update your .env file with these tokens? (y/n): " UPDATE_ENV
        
        if [ "$UPDATE_ENV" = "y" ] || [ "$UPDATE_ENV" = "Y" ]; then
            # Remove old tokens if they exist
            grep -v "^YOTO_ACCESS_TOKEN=" ../.env > ../.env.tmp || true
            grep -v "^YOTO_REFRESH_TOKEN=" ../.env.tmp > ../.env.tmp2 || true
            mv ../.env.tmp2 ../.env
            rm -f ../.env.tmp
            
            # Add new tokens
            echo "" >> ../.env
            echo "# Yoto OAuth Tokens (generated $(date))" >> ../.env
            echo "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN}" >> ../.env
            echo "YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}" >> ../.env
            
            echo -e "${GREEN}✓ Updated .env file${NC}"
        fi
        
        echo ""
        echo -e "${YELLOW}Step 5: Update Google Cloud Run${NC}"
        echo ""
        echo "Run this command to update your Cloud Run service:"
        echo ""
        echo -e "${BLUE}# Update tokens in Cloud Run${NC}"
        cat << EOF
gcloud run services update bird-song-explorer \\
  --region us-central1 \\
  --update-env-vars \\
    "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN},\\
YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}" \\
  --quiet
EOF
        
        echo ""
        echo -e "${GREEN}✓ Token generation complete!${NC}"
        echo ""
        echo -e "${YELLOW}Note:${NC} The service will automatically refresh tokens when they expire."
        
    else
        echo -e "${RED}Error: Failed to get tokens${NC}"
        echo "Response from Yoto API:"
        echo "$RESPONSE"
        echo ""
        echo "Common issues:"
        echo "1. Authorization code has already been used"
        echo "2. Authorization code has expired (they expire quickly)"
        echo "3. Redirect URI doesn't match what's configured in Yoto"
        exit 1
    fi
else
    echo ""
    echo -e "${YELLOW}Instructions:${NC}"
    echo "1. Visit the authorization URL above"
    echo "2. After authorizing, you'll be redirected to your production service"
    echo "3. The service will display the tokens and update instructions"
    echo ""
    echo -e "${BLUE}To check if tokens are working:${NC}"
    echo "curl https://yoto-bird-song-explorer-362662614716.us-central1.run.app/api/v1/daily-update"
fi