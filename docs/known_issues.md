# Known Issues

## Timezone Support for Global Families

**Date Discovered**: 2024-12-24
**Status**: Open
**Priority**: Medium

### Description
The Bird Song Explorer currently updates at a fixed time (midnight UTC) for all users globally. This means families in different timezones don't get a new bird at the start of their local day.

For example:
- When it's January 1st in Sydney, Australian families still hear December 31st's bird
- The bird changes at midnight UTC, which is:
  - 4 PM PST (previous day)
  - 7 PM EST (previous day)  
  - 11 AM Sydney (next day)

### Ideal Behavior
Each family should get a new bird when their local day begins, based on their location.

### Proposed Solution (Not Yet Implemented)
Use real-time webhook updates when the card is played:
1. Detect player's location from IP address
2. Calculate local date/time for that location
3. Select appropriate bird for that location and date
4. Update card during intro playback (no perceived delay)

### Current Workaround
Card updates daily at midnight UTC for all users. This provides a consistent global update time, though not aligned with local days.

### Implementation Notes for Future
- Already created `webhook_handler.go` with timezone detection logic
- Would need to configure Yoto webhook URL in dashboard
- Consider caching birds per location+date to avoid redundant selections
- Use two-stage playback: generic intro plays immediately while updating bird in background

### Related Files
- `/internal/api/webhook_handler.go` - Webhook handler with timezone logic (ready but unused)
- `/internal/services/timezone_bird_selector.go` - Timezone-aware bird selection (ready but unused)

---

## Duplicate "Introduction" Track on Web Interface

**Date Discovered**: 2024-12-24
**Status**: Open
**Priority**: Low (cosmetic issue, app works correctly)

### Description
When viewing the Bird Song Explorer MYO card on the Yoto web interface (my.yotoplay.com), the "Introduction" track appears twice in the playlist. However, the Yoto mobile app displays the tracks correctly with:
1. Introduction
2. [Bird Name]

### Expected Behavior
The web interface should show the same track listing as the mobile app:
- Track 1: Introduction
- Track 2: [Current Bird Name]

### Actual Behavior
The web interface shows:
- Track 1: Introduction
- Track 2: Introduction

### Investigation Notes
- The mobile app displays correctly, indicating the data structure is correct
- This appears to be a display/rendering issue specific to the web UI
- Possible causes:
  - Caching issue on the web interface
  - Web UI incorrectly rendering chapter/track hierarchy
  - Stale data being displayed from a previous update

### Workaround
Use the mobile app to verify correct track listings. The Yoto player itself plays the correct content.

### To Investigate Later
1. Check if the issue persists after 24 hours (cache expiry)
2. Examine the exact JSON structure being sent to the API
3. Test if creating a fresh playlist exhibits the same behavior
4. Check if other MYO cards show similar issues on the web interface

### Screenshots
- Web interface: Shows duplicate "Introduction" tracks
- Mobile app: Shows correct track listing (Introduction, Bird Name)