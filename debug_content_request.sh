#!/bin/bash

# Load the current access token
source .env

echo "Full GET request for content:"
echo "=============================="
echo ""
echo "URL: https://api.yotoplay.com/content/$YOTO_CARD_ID"
echo ""
echo "Headers:"
echo "  Authorization: Bearer $YOTO_ACCESS_TOKEN"
echo "  Accept: application/json"
echo ""
echo "Full curl command:"
echo "curl -v -X GET 'https://api.yotoplay.com/content/$YOTO_CARD_ID' \\"
echo "  -H 'Authorization: Bearer $YOTO_ACCESS_TOKEN' \\"
echo "  -H 'Accept: application/json'"
echo ""
echo "Executing request with verbose output:"
echo "======================================="
echo ""

curl -v -X GET "https://api.yotoplay.com/content/$YOTO_CARD_ID" \
  -H "Authorization: Bearer $YOTO_ACCESS_TOKEN" \
  -H 'Accept: application/json' 2>&1 | grep -E '^[<>]|HTTP|{.*}'