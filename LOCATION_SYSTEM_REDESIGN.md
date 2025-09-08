# Location System Redesign Summary

## Key Changes Implemented

### 1. Removed London Default Fallback
- **Old:** Always returned London when location detection failed
- **New:** Returns error/nil when location detection fails
- **Impact:** System now properly distinguishes between "no location" and "London location"

### 2. Expanded Global Bird Selection
The daily scheduler now rotates through 40+ locations with emphasis on:
- **USA:** 10 major cities (NY, LA, Chicago, Houston, Phoenix, Denver, Seattle, Miami, Boston, SF)
- **Canada:** 6 major cities (Toronto, Montreal, Vancouver, Calgary, Edmonton, Ottawa)
- **UK:** 6 major cities (London, Manchester, Edinburgh, Birmingham, Bristol, Leeds)
- **Mexico:** 6 major cities (Mexico City, Guadalajara, Monterrey, Cancun, Tijuana, Ciudad Juárez)
- **Plus:** Other global locations for diversity

### 3. Regional Bird Detection System
New `BirdRegionalChecker` service that:
- Checks if the daily bird has been spotted within **160km** of user's location
- Looks for sightings in the last **30 days**
- Uses eBird API for real observation data

### 4. Smart Content Adaptation
The system now provides three types of content based on location detection:

#### A. Regional Bird (bird spotted nearby)
- Location-specific facts
- Message: "The [bird] has been recently spotted near [city]! Listen carefully - you might hear one nearby."
- Uses local coordinates for fact generation

#### B. Non-Regional Bird (location known, bird not local)
- Generic bird facts
- Message: "While the [bird] hasn't been spotted recently in [city], it's still a fascinating bird to learn about!"
- Uses generic facts without location bias

#### C. No Location Detected
- Generic bird facts
- No location-specific messaging
- Universal content that works globally

## How It Works Now

### Daily Update Flow (Scheduler)
1. Scheduler triggers `/api/v1/daily-update`
2. Selects bird from rotating global locations (40+ cities)
3. Stores as daily bird with audio URL
4. All users start with this bird

### Card Play Flow (Webhook)
1. User opens card → webhook triggered
2. Try to detect location:
   - **Primary:** IP geolocation
   - **Fallback:** Device timezone
3. Get the daily bird (same for everyone)
4. Check if bird is regional to user (160km, 30 days)
5. Customize content based on regionality:
   - Regional → location-specific facts + "spotted nearby" message
   - Non-regional → generic facts + "not spotted in your area" message  
   - No location → generic facts only

## Benefits of New Approach

### 1. Single Card, Personalized Experience
- Everyone gets the same daily bird (single cardId works)
- Content is personalized based on user's location
- Regional users get enhanced, location-aware experience

### 2. No Fake Defaults
- No more pretending everyone is in London
- Honest messaging when location can't be detected
- Clear distinction between "no location" and "actual location"

### 3. Educational Value
- Users learn if birds are actually in their area
- Encourages local birdwatching when bird is regional
- Still educational even when bird isn't local

### 4. Scalable Design
- Works with single shared card ID
- Daily bird selection covers major user markets
- Graceful degradation when location unavailable

## Configuration Required

### Environment Variables
```bash
YOTO_CARD_ID=         # Your MYO card ID
YOTO_CLIENT_ID=       # Yoto API client
EBIRD_API_KEY=        # For regional checking
XENO_CANTO_API_KEY=   # For bird audio
ELEVEN_LABS_API_KEY=  # For TTS
```

### Testing the New System

#### Test with location (should check regionality):
```bash
curl "https://your-api/api/v1/test-webhook?deviceId=DEVICE_ID"
```

#### Test without location (generic facts):
```bash
curl "https://your-api/api/v1/test-webhook"
```

#### Response includes:
- `isRegional`: true/false
- `regionalMessage`: Contextual message about sightings
- `location`: Detected city or "Unknown"
- `locationSource`: "ip", "timezone", or "none"

## Files Modified

### New Files
- `internal/services/bird_regional_checker.go` - Regional detection service
- `internal/api/webhook_handler_v4.go` - New webhook handler without defaults

### Updated Files
- `internal/services/location.go` - Removed London default
- `internal/api/daily_update.go` - Expanded location list
- `internal/api/router.go` - Using V4 handler
- `internal/api/test_webhook.go` - Updated to V4

### Deprecated Files (moved to `internal/api/deprecated/`)
- `webhook_handler_v2.go`
- `webhook_unified.go`
- `smart_update.go`

## Next Steps

1. **Deploy** the updated system
2. **Monitor** logs to verify location detection and regional checking
3. **Test** with real devices in different locations
4. **Verify** eBird API rate limits are sufficient
5. **Consider** caching species codes to reduce API calls