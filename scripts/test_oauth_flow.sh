#!/bin/bash

# Simple OAuth flow tester for Yoto API

echo "==================================="
echo "     Yoto OAuth Flow Tester       "
echo "==================================="
echo ""

# Load environment variables
if [ -f ../.env ]; then
    source ../.env
else
    echo "Enter your Yoto Client ID:"
    read YOTO_CLIENT_ID
fi

echo "Current Configuration:"
echo "----------------------"
echo "Client ID: $YOTO_CLIENT_ID"
echo "Redirect URI: https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"
echo ""

# Generate a simple state
STATE="test_$(date +%s)"

# Build the authorization URL
AUTH_URL="https://api.yotoplay.com/authorize"
REDIRECT_URI="https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"

# URL encode the redirect URI
ENCODED_REDIRECT_URI=$(echo -n "$REDIRECT_URI" | jq -sRr @uri)

FULL_AUTH_URL="${AUTH_URL}?client_id=${YOTO_CLIENT_ID}&redirect_uri=${ENCODED_REDIRECT_URI}&response_type=code&state=${STATE}&scope=offline_access"

echo "Authorization URL:"
echo "=================="
echo "$FULL_AUTH_URL"
echo ""
echo "Instructions:"
echo "1. Copy and paste this URL into your browser"
echo "2. Log in to your Yoto account"
echo "3. Click 'Authorize' or 'Allow'"
echo "4. Check the URL you're redirected to"
echo ""
echo "Expected redirect URL format:"
echo "https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook?code=XXXXXX&state=${STATE}"
echo ""
echo "If you see an error or no 'code' parameter:"
echo "- Check that the redirect URI matches EXACTLY what's in your Yoto app settings"
echo "- Make sure your Yoto app is not in 'test mode' or restricted"
echo "- Verify the client ID is correct"
echo ""
echo "What URL did you get redirected to? (paste here):"
read REDIRECT_RESULT

if echo "$REDIRECT_RESULT" | grep -q "code="; then
    CODE=$(echo "$REDIRECT_RESULT" | sed -n 's/.*code=\([^&]*\).*/\1/p')
    echo ""
    echo "✓ Found authorization code: $CODE"
    echo ""
    echo "Now let's try to exchange it for tokens..."
    echo ""
    
    TOKEN_URL="https://api.yotoplay.com/oauth/token"
    
    RESPONSE=$(curl -s -X POST "$TOKEN_URL" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=authorization_code" \
        -d "client_id=${YOTO_CLIENT_ID}" \
        -d "code=${CODE}" \
        -d "redirect_uri=${REDIRECT_URI}")
    
    echo "Token exchange response:"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
    
else
    echo ""
    echo "✗ No authorization code found in the URL"
    echo "This suggests the OAuth flow isn't completing properly"
    echo ""
    echo "Debugging steps:"
    echo "1. Verify in Yoto Developer Dashboard that the redirect URI is EXACTLY:"
    echo "   https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"
    echo ""
    echo "2. Check if there's an error parameter in the URL:"
    if echo "$REDIRECT_RESULT" | grep -q "error="; then
        ERROR=$(echo "$REDIRECT_RESULT" | sed -n 's/.*error=\([^&]*\).*/\1/p')
        ERROR_DESC=$(echo "$REDIRECT_RESULT" | sed -n 's/.*error_description=\([^&]*\).*/\1/p')
        echo "   Error: $ERROR"
        echo "   Description: $ERROR_DESC"
    fi
fi