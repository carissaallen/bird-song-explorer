#!/bin/bash

echo "Testing Yoto OAuth Redirect Parameter Names"
echo "============================================"
echo ""

CLIENT_ID="qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU"
REDIRECT_URL="https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook"
STATE="test_$(date +%s)"

echo "Test 1: Using redirect_url (per docs)"
echo "--------------------------------------"
URL1="https://api.yotoplay.com/authorize?client_id=${CLIENT_ID}&redirect_url=${REDIRECT_URL}&response_type=code&state=${STATE}&scope=offline_access"
echo "$URL1"
echo ""

echo "Test 2: Using redirect_uri (OAuth standard)"
echo "-------------------------------------------"
URL2="https://api.yotoplay.com/authorize?client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URL}&response_type=code&state=${STATE}&scope=offline_access"
echo "$URL2"
echo ""

echo "Test 3: Using encoded redirect_url"
echo "-----------------------------------"
ENCODED_URL=$(echo -n "$REDIRECT_URL" | jq -sRr @uri)
URL3="https://api.yotoplay.com/authorize?client_id=${CLIENT_ID}&redirect_url=${ENCODED_URL}&response_type=code&state=${STATE}&scope=offline_access"
echo "$URL3"
echo ""

echo "Test 4: Without scope parameter"
echo "--------------------------------"
URL4="https://api.yotoplay.com/authorize?client_id=${CLIENT_ID}&redirect_url=${REDIRECT_URL}&response_type=code&state=${STATE}"
echo "$URL4"
echo ""

echo "Instructions:"
echo "1. Try each URL above in your browser"
echo "2. See which one redirects after authorization"
echo "3. Note any error messages"
echo ""
echo "If none work, check:"
echo "- Is the app in production/active status?"
echo "- Is the exact redirect URL saved in Yoto dashboard?"
echo "- Are there any restrictions on the app?"