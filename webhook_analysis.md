# Webhook Handler Analysis & Recommendations

## Current State

### Active Endpoints (router.go)
1. **`/api/v1/yoto/webhook`** - Points to `HandleYotoWebhookUnified`
2. **`/webhook`** - Also points to `HandleYotoWebhookUnified` (alternative path)
3. **`/api/v1/test-webhook`** - Test endpoint for simulating webhooks
4. **`/api/v1/smart-update`** - Points to `SmartUpdateHandler`
5. **`/api/v1/daily-update`** - Points to `SmartUpdateHandler` (same handler!)
6. **`/api/v1/update-card/:cardId`** - Points to `SmartUpdateHandler` (manual trigger)

### Multiple Webhook Handlers Found
1. **HandleYotoWebhookUnified** (webhook_unified.go) - Currently active
2. **HandleYotoWebhookV2** (webhook_handler.go) - Not used
3. **HandleYotoWebhookV3** (webhook_handler_v3.go) - Not used, but has better fallback logic
4. **HandleYotoWebhook** (handlers.go) - Old version, not used

### Handler Comparison

#### HandleYotoWebhookUnified (ACTIVE)
- Basic location detection from IP
- Falls back to device timezone
- Uses update cache to prevent rapid updates
- Missing: Global bird fallback logic

#### HandleYotoWebhookV3 (NOT USED - BUT BETTER!)
- Has sophisticated fallback to global birds
- Better location detection with explicit fallback handling
- Caches global birds for consistency
- Skips cache for fallback locations (allows proper update later)
- Better logging and error handling

#### SmartUpdateHandler
- Used for both daily-update and smart-update endpoints
- Has rotating global locations for scheduler
- Good for scheduled updates but not ideal for webhooks

#### DailyUpdateHandler
- Separate implementation for daily updates
- Stores global bird for fallback use
- Not currently routed to any endpoint

## Current Implementation

### 1. Location Detection Priority
The webhook handler uses a cascading approach for location detection:
- **Primary:** IP-based geolocation (most accurate when device IP is available)
- **Secondary:** Device timezone via API (fallback when IP fails)
- **Tertiary:** Global bird selection (when both methods fail)

### 2. Missing Webhook Data
The Yoto webhook payload only contains:
- `eventType` (e.g., "card.played")
- `cardId`
- `deviceId`
- `userId`
- `timestamp`

**It does NOT contain:**
- User's IP address
- User's location
- Device timezone (must be fetched separately)

### 3. Handler Confusion
- Multiple overlapping handlers with different capabilities
- The best handler (HandleYotoWebhookV3) is not being used
- SmartUpdateHandler is doing double duty for different use cases

## Recommended Solution

### Step 1: Use the Right Handler
Replace `HandleYotoWebhookUnified` with `HandleYotoWebhookV3` because it has:
- Better fallback logic for when location can't be detected
- Global bird caching for consistency
- Smarter cache handling (doesn't cache fallback birds)

### Step 2: Location Detection Strategy
The V3 handler already implements the correct priority:
1. **Primary:** IP-based location from the device (most accurate)
2. **Secondary:** Device timezone from `GetDeviceConfig(deviceID)` (reliable fallback)
3. **Tertiary:** Daily global bird (when no location data available)

### Step 3: Clean Up Unused Code
Remove or comment out:
- `HandleYotoWebhook` (handlers.go lines 98-201)
- `HandleYotoWebhookV2` (webhook_handler.go - entire file can be removed)
- `HandleYotoWebhookUnified` (webhook_unified.go - after switching to V3)
- `UpdateCardManually` (handlers.go lines 253-313) - redundant with smart-update

### Step 4: Streamline Endpoints
Keep:
- `/api/v1/yoto/webhook` → `HandleYotoWebhookV3` (for Yoto webhooks)
- `/api/v1/daily-update` → `DailyUpdateHandler` (for scheduler)
- `/api/v1/test-webhook` → `TestWebhookHandler` (for testing)

Remove:
- `/webhook` (duplicate)
- `/api/v1/smart-update` (redundant)
- `/api/v1/update-card/:cardId` (use test-webhook instead)

## Implementation Steps

1. **Update router.go** to use HandleYotoWebhookV3
2. **Test the webhook** using the test endpoint
3. **Remove unused handlers** after confirming V3 works
4. **Update daily scheduler** to use DailyUpdateHandler directly

## Location Detection Flow

The webhook provides device information that enables location detection:
```
User Device → Yoto Servers → Your API (with device IP)
```

The handler attempts location detection in order:
1. **IP Geolocation**: Uses the device's IP for precise location
2. **Timezone Mapping**: Falls back to device timezone if IP fails
3. **Global Selection**: Uses a rotating global bird if both fail

This cascading approach ensures the most accurate location is used when available, with reliable fallbacks for edge cases.

## Testing Recommendation

Use the test webhook endpoint to simulate different scenarios:
```bash
# Test with device ID (will fetch timezone)
curl "https://your-api/api/v1/test-webhook?deviceId=DEVICE_ID"

# Test without device ID (will use global bird)
curl "https://your-api/api/v1/test-webhook"
```

This will help verify the location detection and fallback logic works correctly.