#!/bin/bash

# Yoto OAuth Token Helper Script
# This script helps you get new access and refresh tokens for the Yoto API

echo "=== Yoto OAuth Token Helper ==="
echo ""
echo "Step 1: Open this URL in your browser to log in to Yoto:"
echo ""

CLIENT_ID="qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU"
REDIRECT_URI="https://localhost:8080/callback"

# URL encode the redirect URI
ENCODED_REDIRECT_URI="https%3A%2F%2Flocalhost%3A8080%2Fcallback"

AUTH_URL="https://login.yotoplay.com/authorize?audience=https://api.yotoplay.com&scope=offline_access%20openid%20profile&response_type=code&client_id=${CLIENT_ID}&redirect_uri=${ENCODED_REDIRECT_URI}"

echo "$AUTH_URL"
echo ""
echo "Step 2: After logging in, you'll be redirected to a URL that looks like:"
echo "https://localhost:8080/callback?code=XXXXXX"
echo ""
echo "Your browser will show an error (can't connect to localhost) - that's OK!"
echo "Copy the 'code' value from the URL bar."
echo ""
read -p "Paste the authorization code here: " AUTH_CODE

echo ""
echo "Step 3: Exchanging code for tokens..."
echo ""

# Exchange the code for tokens
RESPONSE=$(curl -s -X POST https://login.yotoplay.com/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "client_id=${CLIENT_ID}" \
  -d "code=${AUTH_CODE}" \
  -d "redirect_uri=${REDIRECT_URI}")

# Check if we got a valid response
if echo "$RESPONSE" | grep -q "access_token"; then
    echo "Success! Here are your tokens:"
    echo ""
    echo "$RESPONSE" | python3 -m json.tool
    
    # Extract tokens
    ACCESS_TOKEN=$(echo "$RESPONSE" | python3 -c "import json, sys; print(json.load(sys.stdin)['access_token'])")
    REFRESH_TOKEN=$(echo "$RESPONSE" | python3 -c "import json, sys; print(json.load(sys.stdin)['refresh_token'])")
    
    echo ""
    echo "=== Copy these commands to update Cloud Run: ==="
    echo ""
    echo "gcloud run services update yoto-bird-song-explorer \\"
    echo "  --region us-central1 \\"
    echo "  --update-env-vars YOTO_ACCESS_TOKEN=\"${ACCESS_TOKEN}\",YOTO_REFRESH_TOKEN=\"${REFRESH_TOKEN}\""
    echo ""
    echo "=== Or update locally in .env: ==="
    echo ""
    echo "YOTO_ACCESS_TOKEN=${ACCESS_TOKEN}"
    echo "YOTO_REFRESH_TOKEN=${REFRESH_TOKEN}"
else
    echo "Error getting tokens. Response:"
    echo "$RESPONSE"
    echo ""
    echo "Common issues:"
    echo "- The authorization code may have expired (they're only valid for a few minutes)"
    echo "- Make sure you copied the entire code value from the URL"
    echo "- Try the process again from Step 1"
fi