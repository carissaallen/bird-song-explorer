# Webhook Implementation Summary

## Changes Made

### 1. Routing Updates (internal/api/router.go)
- **Active webhook endpoint:** `/api/v1/yoto/webhook` → `HandleYotoWebhookV3`
- **Daily update endpoint:** `/api/v1/daily-update` → `DailyUpdateHandler`
- **Test endpoint:** `/api/v1/test-webhook` → `TestWebhookHandler`
- **Commented out deprecated endpoints:**
  - `/webhook` (duplicate)
  - `/api/v1/smart-update` (redundant)
  - `/api/v1/update-card/:cardId` (use test-webhook instead)

### 2. Handler Selection
**Using HandleYotoWebhookV3** because it has:
- Sophisticated fallback logic for location detection failures
- Global bird caching for consistency
- Smart cache handling (doesn't cache fallback birds)
- Better logging for debugging

### 3. Deprecated Code Moved
Moved to `internal/api/deprecated/`:
- `webhook_handler_v2.go` (old V2 handler)
- `webhook_unified.go` (unified handler)
- `smart_update.go` (redundant with daily handler)

Commented out in `handlers.go`:
- `HandleYotoWebhook()` (lines 98-202)
- `UpdateCardManually()` (lines 254-315)

### 4. Utility Functions Added
Created `internal/api/timezone_utils.go`:
- Extracted `GetTimezoneFromLocation()` for shared use

## How Location Detection Works

### Location Detection Priority (in HandleYotoWebhookV3)

1. **Primary Method:** IP-based geolocation
   - Uses the device's IP address from the webhook request
   - Most accurate method when available
   - Provides precise city-level location

2. **Secondary Method:** Device timezone via `GetDeviceConfig(deviceID)`
   - Falls back to this when IP detection fails
   - Maps timezone to approximate location
   - Less precise but reliable for configured devices

3. **Final Fallback:** Daily global bird
   - Used when both IP and timezone detection fail
   - Selected from rotating world locations
   - Cached for consistency throughout the day
   - Not location-cached (allows proper update when location detected)

### Cache Strategy:
- **Regional birds** (IP or timezone detected): Cached by location
- **Global/fallback birds**: NOT cached by location
- This allows the card to update when proper location is detected later

## Testing the Implementation

### Test Webhook with Device ID (will fetch timezone):
```bash
curl "https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/test-webhook?deviceId=YOUR_DEVICE_ID"
```

### Test Webhook without Device ID (will use global bird):
```bash
curl "https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/test-webhook"
```

### Test Daily Update (scheduler):
```bash
curl -X POST "https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/daily-update" \
  -H "X-Scheduler-Token: YOUR_SCHEDULER_TOKEN"
```

## Next Steps

1. **Deploy the changes** to Cloud Run
2. **Test the webhook** with actual Yoto card plays
3. **Monitor logs** to verify location detection is working
4. **Verify daily scheduler** updates the global bird correctly

## Key Insights

1. **IP-based location is the most accurate** when the device's IP is available in the webhook
2. **Device timezone provides reliable fallback** when IP detection fails
3. **Global bird fallback is essential** for users without location data
4. **Smart caching prevents spam** while allowing location-based updates

## Configuration Required

Ensure these environment variables are set:
- `YOTO_CARD_ID` - Your MYO card ID
- `YOTO_CLIENT_ID` - Your Yoto API client ID
- `YOTO_ACCESS_TOKEN` - Valid access token
- `YOTO_REFRESH_TOKEN` - Valid refresh token
- `SCHEDULER_TOKEN` - Token for scheduler authentication (optional)

## Webhook URL for Yoto Dashboard
```
https://bird-song-explorer-362662614716.us-central1.run.app/api/v1/yoto/webhook
```

This is the URL configured in your Yoto developer dashboard.