# Deployment Notes - Enhanced Bird Explorer's Guide

## Changes Made

### 1. Enhanced Fact Generator (V4)
- **File**: `internal/services/improved_bird_facts_v4.go`
- Integrates eBird API for real-time local sightings
- Location-aware content generation
- Improved transitions and natural language flow
- Size comparisons and vivid descriptions
- Nesting information and baby bird facts
- Conservation messages with local actions

### 2. Content Manager Integration
- **File**: `pkg/yoto/content_update.go`
- Added `generateEnhancedBirdDescription()` function
- Falls back to basic description if location unavailable
- Automatically uses V4 generator when coordinates provided
- Maintains backward compatibility

### 3. Environment Configuration
- **Updated**: `deploy/gcp-cloud-run.sh`
- Increased memory to 1Gi for enhanced processing
- Added `USE_ENHANCED_FACTS=true` environment variable
- eBird API key added to secrets management

## Deployment Steps

### Prerequisites
1. Set the eBird API key environment variable:
   ```bash
   export EBIRD_API_KEY="your-ebird-api-key"
   ```

2. Ensure all other required environment variables are set:
   - `ELEVENLABS_API_KEY`
   - `YOTO_CLIENT_SECRET`
   - `YOTO_ACCESS_TOKEN`
   - `YOTO_REFRESH_TOKEN`
   - `YOTO_CARD_ID`

### Deploy to Google Cloud Run
```bash
cd /path/to/bird-song-explorer
./deploy/gcp-cloud-run.sh
```

### Testing the Deployment
After deployment, test the endpoints:
```bash
# Health check
curl https://your-cloud-run-url.run.app/health

# Bird of the day
curl https://your-cloud-run-url.run.app/api/v1/bird-of-day

# Webhook endpoint (for Yoto)
# https://your-cloud-run-url.run.app/api/v1/yoto/webhook
```

## Feature Flags

### USE_ENHANCED_FACTS
When set to `true`, enables:
- eBird API integration for recent sightings
- Location-specific content
- Enhanced transitions and natural language
- Longer, more comprehensive scripts (90-120 seconds)

### Fallback Behavior
If eBird API is unavailable or location data is missing:
- Falls back to original fact generation
- Still includes improved transitions from V3
- Maintains all core functionality

## API Keys Required

1. **eBird API Key**
   - Sign up at: https://ebird.org/api/keygen
   - Used for recent bird sightings
   - Rate limit: 100 requests per hour

2. **ElevenLabs API Key**
   - For text-to-speech generation
   - Required for all narration tracks

3. **Yoto API Credentials**
   - CLIENT_ID and CLIENT_SECRET
   - ACCESS_TOKEN and REFRESH_TOKEN
   - For card content updates

## Memory Requirements
- Increased from 512Mi to 1Gi
- Needed for:
  - Enhanced fact processing
  - Wikipedia content fetching
  - eBird API data processing
  - Audio generation and mixing

## Monitoring
Check Cloud Run logs for:
- eBird API connection status
- Fact generation success/fallback
- Audio generation timing
- Memory usage patterns

## Rollback Plan
If issues occur:
1. Remove `USE_ENHANCED_FACTS` environment variable
2. Redeploy with previous memory setting (512Mi)
3. System will use original fact generation

## Future Enhancements
- Geocoding service for city/state names
- Historical sighting patterns
- Seasonal migration tracking
- Multi-language support