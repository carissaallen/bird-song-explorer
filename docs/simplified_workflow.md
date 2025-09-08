# Simplified Bird Selection Workflow

## Overview
The bird selection system has been simplified to provide better regional coverage while reducing complexity.

## Key Changes

### 1. Cascading Regional Search
When a user plays the card and we detect their location, we now try multiple search strategies:

1. **Near & Recent** (50km radius, 30 days) - Best chance of relevant local birds
2. **Wider Area** (100km radius, 30 days) - Expands search area if nothing nearby  
3. **Regional** (150km radius, 60 days) - Captures seasonal/regional patterns
4. **Global Fallback** - Diverse mix of birds from around the world

### 2. Improved Fallback Birds
Instead of defaulting to only North American birds, the fallback now includes:
- Birds from 6 continents
- Mix of common and interesting species
- Rotates daily for variety

### 3. Simplified Caching
- Removed complex global bird caching between webhook and scheduler
- Each location gets its own daily bird
- Cache prevents repeated API calls for same location/day

### 4. Better Timezone Detection
- Improved coordinate-based timezone mapping
- Covers more regions accurately
- Properly handles Oregon (and other western US states)

## Workflow

### When User Plays Card:

1. **Detect Location**
   - Try IP geolocation first
   - Fall back to device timezone if needed
   - Use location to determine local date/time

2. **Check Cache**
   - If already updated today for this location, return cached bird
   - Prevents excessive API calls

3. **Select Bird**
   - Try 50km/30days search
   - If no results, try 100km/30days
   - If no results, try 150km/60days
   - If still no results, use global fallback

4. **Update Card**
   - Generate intro with random voice
   - Update card content with bird info
   - Cache the result

### Daily Scheduler:

1. **Pick Global Location**
   - Rotates through 10 diverse cities daily
   - Ensures variety in daily birds

2. **Select Bird**
   - Uses same cascading logic
   - Updates the default card

## Benefits

1. **Better Regional Coverage**: More likely to find local birds with relaxed constraints
2. **Simpler Code**: Removed complex global bird caching between services
3. **More Variety**: Global fallback includes birds from all continents
4. **Reliable Fallbacks**: Multiple fallback levels ensure we always have a bird
5. **Efficient**: Caching prevents repeated searches for same location/day

## Technical Details

- **eBird API**: Searches for recent observations with configurable radius/timeframe
- **XenoCanto API**: Provides bird audio recordings
- **Location Detection**: IP geolocation → Device timezone → Default
- **Timezone Mapping**: Coordinate-based IANA timezone detection