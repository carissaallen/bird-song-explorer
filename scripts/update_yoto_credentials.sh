#!/bin/bash

# Update Cloud Run service with new Yoto confidential client credentials
# Make sure to set the tokens first by running get_yoto_token

echo "🔄 Updating Bird Song Explorer with new Yoto credentials..."
echo "============================================"

# New confidential client credentials
YOTO_CLIENT_ID="a0PBKaHZnoB02Jpe5RZ9znZV5e1OXPr7"
YOTO_CLIENT_SECRET="7EU77IABOjxVDfl666QeC9uxO9Vt_FZvAEfFhNZ1TBXEC81P9XtPa7RInzaIiPZt"

# Check if tokens are provided as arguments or in environment
if [ -n "$1" ] && [ -n "$2" ]; then
    YOTO_ACCESS_TOKEN="$1"
    YOTO_REFRESH_TOKEN="$2"
    echo "Using tokens from command line arguments"
elif [ -n "$YOTO_ACCESS_TOKEN" ] && [ -n "$YOTO_REFRESH_TOKEN" ]; then
    echo "Using tokens from environment variables"
else
    echo "❌ Error: Tokens not provided"
    echo ""
    echo "Usage: $0 [ACCESS_TOKEN] [REFRESH_TOKEN]"
    echo "Or set YOTO_ACCESS_TOKEN and YOTO_REFRESH_TOKEN environment variables"
    echo ""
    echo "To get new tokens, run:"
    echo "  cd ../cmd/get_yoto_token && go run main.go"
    exit 1
fi

# Update Cloud Run service
echo "📦 Updating Cloud Run service..."
gcloud run services update bird-song-explorer \
  --region us-central1 \
  --update-env-vars \
    "YOTO_CLIENT_ID=${YOTO_CLIENT_ID},\
YOTO_CLIENT_SECRET=${YOTO_CLIENT_SECRET},\
YOTO_ACCESS_TOKEN=${YOTO_ACCESS_TOKEN},\
YOTO_REFRESH_TOKEN=${YOTO_REFRESH_TOKEN}" \
  --quiet

if [ $? -eq 0 ]; then
    echo "✅ Successfully updated Cloud Run service!"
    echo ""
    echo "🔍 Verifying deployment..."
    gcloud run services describe bird-song-explorer \
      --region us-central1 \
      --format="value(status.url)"
    
    echo ""
    echo "📝 Test the service:"
    echo "  curl https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/daily-update"
else
    echo "❌ Failed to update Cloud Run service"
    exit 1
fi