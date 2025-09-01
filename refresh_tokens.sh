#!/bin/bash

echo "Refreshing Yoto tokens..."

REFRESH_TOKEN="v1.MfH3ZyzPkOHCfI7hyQuHYlHmnM9-7DWOZhtlmHqQ22khZjQSTY7DpYLUMNZq9eozTda3UN7zmvZWCvyk_FWlg0k"
CLIENT_ID="qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU"

curl -X POST https://api.yotoplay.com/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "client_id=$CLIENT_ID" \
  -d "refresh_token=$REFRESH_TOKEN"