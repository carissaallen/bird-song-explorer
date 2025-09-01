#!/bin/bash

# Load the current access token
source .env

echo "Full GET request for upload URL:"
echo "================================"
echo ""
echo "URL: https://api.yotoplay.com/media/transcode/audio/uploadUrl"
echo ""
echo "Headers:"
echo "  Authorization: Bearer $YOTO_ACCESS_TOKEN"
echo "  Accept: application/json"
echo ""
echo "Full curl command:"
echo "curl -v -X GET 'https://api.yotoplay.com/media/transcode/audio/uploadUrl' \\"
echo "  -H 'Authorization: Bearer $YOTO_ACCESS_TOKEN' \\"
echo "  -H 'Accept: application/json'"
echo ""
echo "Executing request with verbose output:"
echo "======================================="
echo ""

curl -v -X GET 'https://api.yotoplay.com/media/transcode/audio/uploadUrl' \
  -H "Authorization: Bearer $YOTO_ACCESS_TOKEN" \
  -H 'Accept: application/json' 2>&1 | grep -E '^[<>]|HTTP|{.*}'