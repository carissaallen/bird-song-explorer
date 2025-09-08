# Webhook Issue Analysis

## Current Problem
When you play the card, you're getting the bird from the daily scheduler (Yellow-legged Thrush from São Paulo), not a regional bird from Bend, Oregon.

## Root Cause
**The webhook is not being triggered when you play the card.**

Looking at the logs:
- Daily update ran at 03:49:00 and updated the card with Yellow-legged Thrush
- No webhook calls have been received since deployment
- The card is showing the bird from the daily update, not from playing the card

## Why This Happens

### Current Flow (Problematic):
1. **Daily Scheduler (midnight)** → Directly updates the card content via Yoto API
2. **User plays card** → Should trigger webhook → Should update with regional bird
   - But webhook is NOT being called by Yoto

### The Issue:
- Yoto may not support webhooks for card.played events for MYO cards
- Or the webhook needs to be configured in the Yoto developer portal
- The daily scheduler overwrites any regional selection

## Proposed Solution

### Option 1: Remove Daily Scheduler
- Only update when webhook is triggered
- Each play gets a fresh bird based on location
- Problem: If webhook doesn't work, card never updates

### Option 2: Smart Daily Update
- Daily update checks for user's last known location
- Updates with regional bird if location known
- Falls back to global bird if no location history

### Option 3: Manual Trigger Only
- Remove automatic updates entirely  
- Add a web interface or app to manually trigger updates
- User controls when bird changes

## Recommended Approach

Since webhooks aren't working, we should:

1. **Keep daily scheduler but make it smarter**:
   - Store last known user location
   - Use that location for daily updates
   - Fall back to global rotation if no location known

2. **Add manual refresh endpoint**:
   - `/api/v1/refresh-card` that can be called manually
   - Detects caller's location and updates accordingly

3. **Debug webhook setup**:
   - Check Yoto developer portal for webhook configuration
   - Verify if MYO cards support webhooks
   - Test with different event types