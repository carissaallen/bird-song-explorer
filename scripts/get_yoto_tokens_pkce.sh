#!/bin/bash

# Yoto OAuth Token Generation Script with PKCE
# This script implements the proper PKCE flow as required by Yoto API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Yoto OAuth Token Generator (PKCE)  ${NC}"
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
AUTH_URL="https://login.yotoplay.com/authorize"
TOKEN_URL="https://login.yotoplay.com/oauth/token"
REDIRECT_URI="http://localhost:8080/callback"

echo -e "${YELLOW}Step 1: Generate PKCE Parameters${NC}"
echo ""

# Generate PKCE code verifier (43-128 characters, URL-safe base64)
# This is a random string that will be used to verify the token exchange
CODE_VERIFIER=$(openssl rand -base64 96 | tr -d '\n' | tr '+/' '-_' | tr -d '=')

# Generate code challenge (SHA256 hash of verifier, base64url encoded)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -sha256 -binary | base64 | tr '+/' '-_' | tr -d '=')

echo "PKCE Code Verifier generated (first 20 chars): ${CODE_VERIFIER:0:20}..."
echo "PKCE Code Challenge generated (first 20 chars): ${CODE_CHALLENGE:0:20}..."
echo ""

echo -e "${YELLOW}Step 2: Generate Authorization URL${NC}"
echo ""
echo "You need to visit this URL in your browser to authorize the application:"
echo ""

# Build authorization URL with PKCE parameters
AUTH_FULL_URL="${AUTH_URL}?audience=https://api.yotoplay.com&scope=offline_access&response_type=code&client_id=${YOTO_CLIENT_ID}&code_challenge=${CODE_CHALLENGE}&code_challenge_method=S256&redirect_uri=${REDIRECT_URI}"

echo -e "${GREEN}${AUTH_FULL_URL}${NC}"
echo ""
echo -e "${YELLOW}Step 3: Get Authorization Code${NC}"
echo ""
echo "1. Open the URL above in your browser"
echo "2. Log in to your Yoto account"
echo "3. Authorize the application"
echo "4. You'll be redirected to: http://localhost:8080/callback?code=XXXXX"
echo "5. Copy the 'code' parameter from the URL"
echo ""
read -p "Enter the authorization code: " AUTH_CODE

if [ -z "$AUTH_CODE" ]; then
    echo -e "${RED}Error: Authorization code cannot be empty${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 4: Exchange Code for Tokens (with PKCE)${NC}"
echo ""

# Exchange authorization code for tokens using PKCE
# Note: No client_secret is needed for public clients with PKCE
RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=authorization_code" \
    -d "client_id=${YOTO_CLIENT_ID}" \
    -d "code_verifier=${CODE_VERIFIER}" \
    -d "code=${AUTH_CODE}" \
    -d "redirect_uri=${REDIRECT_URI}")

# Check if response contains access_token
if echo "$RESPONSE" | grep -q "access_token"; then
    # Extract tokens using jq if available, otherwise use grep
    if command -v jq &> /dev/null; then
        ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
        REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')
        EXPIRES_IN=$(echo "$RESPONSE" | jq -r '.expires_in')
    else
        ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        REFRESH_TOKEN=$(echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
        EXPIRES_IN=$(echo "$RESPONSE" | grep -o '"expires_in":[0-9]*' | cut -d':' -f2)
    fi
    
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
    if [ ! -z "$EXPIRES_IN" ]; then
        echo -e "${BLUE}Expires In:${NC} $EXPIRES_IN seconds"
        echo ""
    fi
    
    # Update .env file
    echo -e "${YELLOW}Step 5: Update Configuration${NC}"
    echo ""
    read -p "Do you want to update your .env file with these tokens? (y/n): " UPDATE_ENV
    
    if [ "$UPDATE_ENV" = "y" ] || [ "$UPDATE_ENV" = "Y" ]; then
        # Backup current .env
        cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
        
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
        
        echo -e "${GREEN}✓ Updated .env file (backup saved as .env.backup.*)${NC}"
    fi
    
    echo ""
    echo -e "${GREEN}✓ Token generation complete!${NC}"
    echo ""
    echo -e "${YELLOW}To update Google Cloud Run deployment:${NC}"
    echo ""
    echo -e "${BLUE}Step 1: Update the secrets in Google Secret Manager:${NC}"
    echo "echo -n \"${ACCESS_TOKEN}\" | gcloud secrets versions add yoto-access-token --data-file=-"
    echo "echo -n \"${REFRESH_TOKEN}\" | gcloud secrets versions add yoto-refresh-token --data-file=-"
    echo ""
    echo -e "${BLUE}Step 2: Force Cloud Run to use the new secret versions:${NC}"
    echo "gcloud run services update bird-song-explorer \\"
    echo "  --region us-central1 \\"
    echo "  --set-secrets=\"YOTO_ACCESS_TOKEN=yoto-access-token:latest,YOTO_REFRESH_TOKEN=yoto-refresh-token:latest\""
    echo ""
    echo -e "${YELLOW}Important Notes:${NC}"
    echo "1. Access tokens expire after ${EXPIRES_IN:-3600} seconds"
    echo "2. Refresh tokens are SINGLE-USE - each refresh gives you a new refresh token"
    echo "3. Always store the new refresh token after using it"
    echo ""
    echo -e "${YELLOW}To test your tokens locally:${NC}"
    echo "curl -H \"Authorization: Bearer \$YOTO_ACCESS_TOKEN\" https://api.yotoplay.com/user/devices"
    
else
    echo -e "${RED}Error: Failed to get tokens${NC}"
    echo "Response from Yoto API:"
    echo "$RESPONSE" | python -m json.tool 2>/dev/null || echo "$RESPONSE"
    echo ""
    echo "Common issues:"
    echo "1. Authorization code has already been used (they're single-use)"
    echo "2. Authorization code has expired (they expire quickly)"
    echo "3. Redirect URI doesn't match exactly"
    echo "4. PKCE code verifier doesn't match the challenge"
    exit 1
fi