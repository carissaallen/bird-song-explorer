# Voice Transitions with ElevenLabs

## Overview

The Bird Song Explorer uses ElevenLabs' `previous_text` feature to create smooth, natural transitions between different audio tracks. This ensures consistent voice characteristics and natural flow as the narrator moves from one topic to another.

## How It Works

### Track Flow
1. **Track 1: Introduction** - Pre-recorded intro (8 variations)
2. **Track 2: Today's Bird** - Announces the bird name
3. **Track 3: Bird Song** - Actual bird recording (no narration)
4. **Track 4: Bird Explorer's Guide** - Educational facts about the bird
5. **Track 5: See You Tomorrow!** - Outro with jokes/wisdom

### Transition Implementation

Each TTS-generated track includes the `previous_text` field from the preceding narration:

- **Track 2** uses `previous_text` from Track 1 (Introduction)
- **Track 4** uses `previous_text` from Track 2 (Today's Bird announcement)
- **Track 5** uses `previous_text` from Track 4 (Bird Explorer's Guide)

### Code Structure

```go
// In generateBirdAnnouncement (Track 2)
requestBody := map[string]interface{}{
    "text":          announcement,
    "model_id":      "eleven_monolingual_v1",
    "previous_text": cm.lastIntroText,  // From Track 1
    // ... voice settings
}

// In generateBirdDescription (Track 4)
requestBody := map[string]interface{}{
    "text":          descriptionText,
    "model_id":      "eleven_monolingual_v1",
    "previous_text": cm.lastAnnouncementText,  // From Track 2
    // ... voice settings
}

// In generateOutro (Track 5)
requestBody := map[string]interface{}{
    "text":          outroText,
    "model_id":      "eleven_monolingual_v1",
    "previous_text": cm.lastDescriptionText,  // From Track 4
    // ... voice settings
}
```

### Pre-recorded Intro Handling

Since Track 1 uses pre-recorded audio files, the system extracts the corresponding text based on the filename pattern:

```go
// intro_00_Amelia.mp3 → "Welcome, nature detectives! Time to discover..."
// intro_01_Antoni.mp3 → "Hello, bird explorers! Today's special bird is..."
```

This extracted text is then used as `previous_text` for Track 2.

## Benefits

1. **Natural Flow**: The voice maintains consistent prosody and tone across tracks
2. **Context Awareness**: Each track "remembers" what was said before
3. **Smooth Transitions**: Reduces jarring changes in voice characteristics
4. **Better User Experience**: Creates a more cohesive listening experience for children

## Debugging

Look for `[TRANSITION]` log entries in the server output:

```
[TRANSITION] Extracted intro text from URL for intro 3: "Welcome back, little listeners..."
[TRANSITION] Track 2 - Using previous_text from Introduction: "Welcome back, little listeners..."
[TRANSITION] Track 4 - Using previous_text from Track 2: "Today's bird is the American Robin..."
[TRANSITION] Outro - Using previous_text from Track 4: "Did you know? The American Robin..."
```

## Testing

Run the transition test:
```bash
./scripts/test_transitions.sh
```

This will trigger a daily update and show all transition logging to verify the feature is working correctly.