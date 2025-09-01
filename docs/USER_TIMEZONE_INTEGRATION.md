# User Timezone Integration for Nature Sounds

## How It Works

The system can determine the user's local time through multiple methods:

### 1. Primary Method: Yoto Device Timezone
When a Yoto card is played, we receive:
- `deviceID` in the webhook payload
- Call `GetDeviceConfig(deviceID)` to get device configuration
- Extract `GeoTimezone` field (e.g., "America/New_York", "Europe/London")
- This is the actual timezone configured on the user's Yoto device

### 2. Fallback: IP-based Location
- Get user's IP address from webhook request
- Convert to geographic location
- Derive timezone from coordinates

## Implementation in Webhook Handler

```go
// In webhook_handler.go, when processing card.played event:

// Get device timezone
var userTimezone string
if webhook.DeviceID != "" {
    deviceConfig, err := h.yotoClient.GetDeviceConfig(webhook.DeviceID)
    if err == nil && deviceConfig != nil {
        userTimezone = deviceConfig.Device.Config.GeoTimezone
        log.Printf("User timezone from device: %s", userTimezone)
    }
}

// Use timezone for nature sound selection
timeHelper := services.NewUserTimeHelper()
userContext := timeHelper.GetUserTimeContext(userTimezone)

// Mix intro with appropriate nature sounds for user's time
mixer := services.NewIntroMixer()
mixedIntro, err := mixer.MixIntroWithNatureSoundsForUser(
    introData, 
    "", // Empty string = auto-select based on time
    userTimezone,
)
```

## Nature Sound Schedule (User's Local Time)

| Time Period | Hours | Nature Sound | Description |
|------------|-------|--------------|-------------|
| Early Morning | 5am-9am | morning_birds | Dawn chorus, birds waking up |
| Late Morning | 9am-12pm | forest | Active forest sounds |
| Afternoon | 12pm-5pm | meadow | Open field ambience |
| Evening | 5pm-8pm | gentle_rain | Calming rain sounds |
| Late Evening | 8pm-10pm | stream | Peaceful water sounds |
| Night | 10pm-5am | night | Crickets, owls, night sounds |

## Benefits

1. **Contextual Audio**: Morning users hear dawn chorus, evening users hear crickets
2. **Global Support**: Works across all timezones
3. **Automatic**: No user configuration needed
4. **Fallback**: Uses server time if timezone unavailable

## Testing

```bash
# Test with specific timezone
go run cmd/test_user_time/main.go -timezone "America/New_York"
go run cmd/test_user_time/main.go -timezone "Europe/London"
go run cmd/test_user_time/main.go -timezone "Australia/Sydney"
```

## Curated Library Approach

Instead of fetching from Xeno-canto API each time:

1. **Pre-select Quality Sounds**: 
   - Manually curate 10-15 high-quality ambient recordings
   - Ensure consistent volume levels
   - Verify they're actually ambient (not single bird calls)

2. **Store Locally**:
   ```
   assets/nature_sounds/
   ├── morning_birds_1.mp3
   ├── morning_birds_2.mp3
   ├── forest_day.mp3
   ├── gentle_rain.mp3
   ├── stream_flowing.mp3
   ├── meadow_afternoon.mp3
   └── night_crickets.mp3
   ```

3. **Runtime Mixing**:
   - Select sound based on user's local time
   - Mix with intro in real-time
   - Cache mixed result for performance

## Privacy Note

- We only use timezone data to enhance audio experience
- No personal location data is stored
- Timezone is derived from device settings, not GPS