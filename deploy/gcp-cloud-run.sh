#!/bin/bash

# GCP Cloud Run Deployment Script
# Prerequisites: gcloud CLI installed and authenticated

PROJECT_ID="${GCP_PROJECT_ID:-your-gcp-project-id}"
REGION="${GCP_REGION:-us-central1}"
SERVICE_NAME="bird-song-explorer"

# Check if PROJECT_ID is still the default
if [ "$PROJECT_ID" = "your-gcp-project-id" ]; then
    echo "⚠️  GCP_PROJECT_ID not set!"
    echo ""
    echo "Please set your Google Cloud Project ID:"
    echo "  export GCP_PROJECT_ID='your-actual-project-id'"
    echo ""
    echo "Or edit this script and replace 'your-gcp-project-id' with your actual project ID"
    echo ""
    echo "To find your project ID:"
    echo "  gcloud projects list"
    echo ""
    exit 1
fi

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
echo -n "$EBIRD_API_KEY" | gcloud secrets create ebird-api-key --data-file=- 2>/dev/null || echo "ebird-api-key already exists"
echo -n "$XENOCANTO_API_KEY" | gcloud secrets create xenocanto-api-key --data-file=- 2>/dev/null || echo "xenocanto-api-key already exists"
echo -n "$ELEVENLABS_API_KEY" | gcloud secrets create elevenlabs-api-key --data-file=- 2>/dev/null || echo "elevenlabs-api-key already exists"
if [ ! -z "$YOTO_CLIENT_SECRET" ]; then
    echo -n "$YOTO_CLIENT_SECRET" | gcloud secrets create yoto-client-secret --data-file=- 2>/dev/null || echo "yoto-client-secret already exists"
else
    echo "Skipping yoto-client-secret (not set)"
fi
echo -n "$YOTO_ACCESS_TOKEN" | gcloud secrets create yoto-access-token --data-file=- 2>/dev/null || echo "yoto-access-token already exists"
echo -n "$YOTO_REFRESH_TOKEN" | gcloud secrets create yoto-refresh-token --data-file=- 2>/dev/null || echo "yoto-refresh-token already exists"

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
  --memory 1Gi \
  --set-env-vars="ENV=production,YOTO_CLIENT_ID=qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU,YOTO_API_BASE_URL=https://api.yotoplay.com,YOTO_CARD_ID=$YOTO_CARD_ID,USE_ENHANCED_FACTS=true" \
  --set-secrets="EBIRD_API_KEY=ebird-api-key:latest,XENOCANTO_API_KEY=xenocanto-api-key:latest,ELEVENLABS_API_KEY=elevenlabs-api-key:latest,YOTO_ACCESS_TOKEN=yoto-access-token:latest,YOTO_REFRESH_TOKEN=yoto-refresh-token:latest"

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