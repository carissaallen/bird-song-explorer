#!/bin/bash

# GCP Cloud Run Deployment Script
# Prerequisites: gcloud CLI installed and authenticated

PROJECT_ID="yoto-bird-song-explorer"
REGION="us-central1"
SERVICE_NAME="bird-song-explorer"

echo "🚀 Deploying Bird Song Explorer to Google Cloud Run"
echo "=================================================="

# 1. Set project
echo "Setting GCP project..."
gcloud config set project $PROJECT_ID

# 2. Enable required APIs
echo "Enabling required APIs..."
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com

# 3. Build and push container
echo "Building container..."
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME

# 4. Deploy to Cloud Run with environment variables
echo "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --set-env-vars="ENV=production,\
BASE_URL=https://bird-song-explorer-362662614716.us-central1.run.app,\
YOTO_API_BASE_URL=https://api.yotoplay.com,\
USE_STREAMING=true,\
BIRD_FACT_GENERATOR=enhanced,\
YOTO_CARD_ID=ipHAS" \
  --set-secrets="EBIRD_API_KEY=ebird-api-key:latest,\
YOTO_ACCESS_TOKEN=yoto-access-token:latest,\
YOTO_REFRESH_TOKEN=yoto-refresh-token:latest,\
ELEVENLABS_API_KEY=elevenlabs-api-key:latest,\
XENOCANTO_API_KEY=xenocanto-api-key:latest,\
YOTO_CLIENT_ID=yoto-client-id:latest"

# 5. Get the URL
echo ""
echo "✅ Deployment complete!"
echo ""
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --platform managed --region $REGION --format 'value(status.url)')
echo "🌐 Your app is live at: $SERVICE_URL"
echo ""
echo "📝 Yoto Webhook URL: ${SERVICE_URL}/api/v1/yoto/webhook"
echo ""
echo "Test endpoints:"
echo "  curl ${SERVICE_URL}/health"
echo "  curl ${SERVICE_URL}/api/v1/bird-of-day"