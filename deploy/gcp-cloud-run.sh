#!/bin/bash

# GCP Cloud Run Deployment Script
# Prerequisites: gcloud CLI installed and authenticated

PROJECT_ID="your-gcp-project-id"
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
gcloud services enable secretmanager.googleapis.com

# 3. Create secrets in Secret Manager
echo "Creating secrets..."
echo -n "$EBIRD_API_KEY" | gcloud secrets create ebird-api-key --data-file=-
echo -n "$XENOCANTO_API_KEY" | gcloud secrets create xenocanto-api-key --data-file=-
echo -n "$ELEVENLABS_API_KEY" | gcloud secrets create elevenlabs-api-key --data-file=-
echo -n "$YOTO_CLIENT_SECRET" | gcloud secrets create yoto-client-secret --data-file=-

# 4. Build and push container
echo "Building container..."
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME

# 5. Deploy to Cloud Run
echo "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --set-env-vars="PORT=8080,ENV=production,YOTO_CLIENT_ID=qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU,YOTO_API_BASE_URL=https://api.yotoplay.com" \
  --set-secrets="EBIRD_API_KEY=ebird-api-key:latest,XENOCANTO_API_KEY=xenocanto-api-key:latest,ELEVENLABS_API_KEY=elevenlabs-api-key:latest,YOTO_CLIENT_SECRET=yoto-client-secret:latest"

# 6. Get the URL
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