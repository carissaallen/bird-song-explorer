#!/bin/bash

# Yoto OAuth with PKCE for Public Client Apps
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Yoto OAuth with PKCE Flow${NC}"
echo "=========================="
echo ""

# Load environment
if [ -f ../.env ]; then
    source ../.env
fi

CLIENT_ID="${YOTO_CLIENT_ID:-qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU}"
REDIRECT_URI="https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"

# Generate PKCE code verifier (43-128 characters, URL-safe base64)
CODE_VERIFIER=$(openssl rand -base64 96 | tr -d '\n' | tr '+/' '-_' | tr -d '=')

# Generate code challenge (SHA256 hash of verifier, base64url encoded)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -sha256 -binary | base64 | tr '+/' '-_' | tr -d '=')

echo "PKCE Code Verifier: ${CODE_VERIFIER:0:20}..."
echo "PKCE Code Challenge: ${CODE_CHALLENGE:0:20}..."
echo ""

# Build authorization URL with PKCE
AUTH_URL="https://login.yotoplay.com/authorize"
FULL_URL="${AUTH_URL}?audience=https://api.yotoplay.com&scope=offline_access&response_type=code&client_id=${CLIENT_ID}&code_challenge=${CODE_CHALLENGE}&code_challenge_method=S256&redirect_uri=${REDIRECT_URI}"

echo -e "${GREEN}Step 1: Visit this URL to authorize:${NC}"
echo "$FULL_URL"
echo ""
echo "After authorizing, you'll be redirected to your app with a code parameter."
echo ""
read -p "Enter the authorization code from the redirect URL: " AUTH_CODE

if [ -z "$AUTH_CODE" ]; then
    echo -e "${RED}No code provided${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 2: Exchanging code for tokens with PKCE...${NC}"

# Exchange code for tokens using PKCE
RESPONSE=$(curl -s -X POST "https://login.yotoplay.com/oauth/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=authorization_code" \
    -d "client_id=${CLIENT_ID}" \
    -d "code_verifier=${CODE_VERIFIER}" \
    -d "code=${AUTH_CODE}" \
    -d "redirect_uri=${REDIRECT_URI}")

echo "Response: $RESPONSE"

# Check for access_token
if echo "$RESPONSE" | grep -q "access_token"; then
    ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token' 2>/dev/null || echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token' 2>/dev/null || echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    
    echo ""
    echo -e "${GREEN}✓ Success! Tokens obtained${NC}"
    echo ""
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
    
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
else
    echo -e "${RED}Failed to get tokens${NC}"
    echo "Response: $RESPONSE"
fi