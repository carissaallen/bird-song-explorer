# Adding New Regional Voices to Bird Song Explorer

## Overview
This guide explains how to add new voices from different regions to provide variety while maintaining consistency within each daily playlist.

## Step 1: Find Voice IDs from ElevenLabs

### Option A: Using ElevenLabs Voice Library
1. Go to https://elevenlabs.io/voice-library
2. Filter by:
   - Language: English
   - Use Case: Children's Stories / Education
   - Accent: Select different regions (British, Australian, Irish, etc.)
3. Test voices by clicking "Play" on samples
4. When you find a suitable voice, click on it to get the voice ID

### Option B: Using ElevenLabs API
```bash
# List available voices
curl -X GET "https://api.elevenlabs.io/v1/voices" \
  -H "xi-api-key: $ELEVENLABS_API_KEY" | jq '.voices[] | {name: .name, voice_id: .voice_id, labels: .labels}'
```

### Option C: Using Pre-made Voices
Common voice IDs for different regions:
- **British**: "ThT5KcBeYPX3keUQqHPh" (Dorothy)
- **Australian**: "XB0fDUnXU5powFXDhCwa" (Charlotte)  
- **Irish**: "mTSvIrm2hmcnOvb21nW2" (Mimi)
- **Scottish**: "CwhRBWXzGAHq8TQ4Fs17" (Alice)
- **Indian**: "SKnFmXtzMJQHaLveShBH" (Neerja)

## Step 2: Test Voice Quality

Before generating all intros, test each voice:

```bash
# Test a single voice
VOICE_ID="YOUR_VOICE_ID"
VOICE_NAME="TestVoice"

curl -X POST "https://api.elevenlabs.io/v1/text-to-speech/$VOICE_ID" \
  -H "xi-api-key: $ELEVENLABS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello, bird explorers! Today we will discover an amazing bird.",
    "model_id": "eleven_multilingual_v2",
    "voice_settings": {
      "stability": 0.50,
      "similarity_boost": 0.80,
      "use_speaker_boost": true,
      "speed": 1.0,
      "style": 0
    }
  }' \
  --output "test_${VOICE_NAME}.mp3"

# Play the test
afplay "test_${VOICE_NAME}.mp3"
```

## Step 3: Generate Pre-recorded Intros

1. Edit `/scripts/generate_regional_intros.sh` to add your voice IDs:
```bash
declare -A VOICES=(
    ["Amelia"]="ZF6FPAbjXT4488VcRRnw"      # British
    ["Antoni"]="ErXwobaYiN019PkySvjV"      # American
    ["YourVoice"]="YOUR_VOICE_ID"          # Region
)
```

2. Run the generation script:
```bash
./scripts/generate_regional_intros.sh
```

3. Review generated files in `final_intros_regional/`

## Step 4: Update Voice Configuration

Edit `/internal/config/voices.go`:

```go
var DefaultVoices = []VoiceProfile{
    {
        ID:       "ZF6FPAbjXT4488VcRRnw",
        Name:     "Amelia",
        Region:   "British",
        Language: "en-GB",
    },
    {
        ID:       "ErXwobaYiN019PkySvjV",
        Name:     "Antoni",
        Region:   "American",
        Language: "en-US",
    },
    // Add your new voices here
    {
        ID:       "YOUR_VOICE_ID",
        Name:     "VoiceName",
        Region:   "Australian",
        Language: "en-AU",
    },
}
```

## Step 5: Deploy Changes

1. Copy approved intros to production directory:
```bash
cp final_intros_regional/*.mp3 final_intros/
```

2. Build and deploy:
```bash
gcloud builds submit --tag gcr.io/yoto-bird-song-explorer/bird-song-explorer
gcloud run deploy bird-song-explorer \
  --image gcr.io/yoto-bird-song-explorer/bird-song-explorer:latest \
  --region us-central1
```

## Voice Selection Criteria

When selecting voices, consider:

1. **Clarity**: Voice should be clear and easy to understand
2. **Age-appropriateness**: Friendly, warm tone suitable for children
3. **Energy Level**: Engaging but not hyperactive
4. **Regional Authenticity**: Natural accent from the region
5. **Consistency**: Similar pacing and tone across different texts

## Testing Voice Rotation

The system automatically rotates voices daily using this formula:
```
voiceIndex = (year*10000 + month*100 + day) % numberOfVoices
```

To test different days:
```bash
# Check which voice will be used on a specific date
curl https://your-app.run.app/api/v1/debug/voice-for-date?date=2024-08-27
```

## Troubleshooting

### Voice Not Found
If a voice ID is invalid, you'll see:
```
Warning: No pre-recorded intros for voice [VoiceName], using Antoni
```

### Rate Limiting
ElevenLabs has rate limits. If you hit them:
- Add delays between API calls
- Use the script's skip feature for existing files
- Consider upgrading your ElevenLabs plan

### Voice Quality Issues
If a voice doesn't sound right:
- Adjust voice_settings (stability, similarity_boost)
- Try different model_id (eleven_turbo_v2 for faster, eleven_multilingual_v2 for quality)
- Test with different intro texts