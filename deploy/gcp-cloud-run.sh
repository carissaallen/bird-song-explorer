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
YOTO_CLIENT_ID=$YOTO_CLIENT_ID,\
YOTO_API_BASE_URL=$YOTO_API_BASE_URL,\
YOTO_CARD_ID=$YOTO_CARD_ID,\
YOTO_ACCESS_TOKEN=$YOTO_ACCESS_TOKEN,\
YOTO_REFRESH_TOKEN=$YOTO_REFRESH_TOKEN,\
EBIRD_API_KEY=$EBIRD_API_KEY,\
XENOCANTO_API_KEY=$XENOCANTO_API_KEY,\
ELEVENLABS_API_KEY=$ELEVENLABS_API_KEY"

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