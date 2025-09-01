#!/bin/bash

# Generate intro files for a specific voice using ElevenLabs API
# Usage: ./generate_intros_for_voice.sh <VOICE_ID> <VOICE_NAME>

set -e

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: $0 <VOICE_ID> <VOICE_NAME>"
    echo "Example: $0 ErXwobaYiN019PkySvjV Antoni"
    exit 1
fi

VOICE_ID="$1"
VOICE_NAME="$2"
OUTPUT_DIR="final_intros"

# Check for ElevenLabs API key
if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo "❌ Error: ELEVENLABS_API_KEY environment variable not set"
    echo "Please run: export ELEVENLABS_API_KEY='your_api_key'"
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Intro scripts (8 variations)
declare -a INTROS=(
    "Welcome, nature detectives! Time to discover an amazing bird from your neighborhood."
    "Hello, bird explorers! Today's special bird is waiting to sing for you."
    "Ready for an adventure? Let's meet today's featured bird from your area!"
    "Welcome back, little listeners! A wonderful bird is calling just for you."
    "Hello, young scientists! Let's explore the amazing birds living near you."
    "Calling all bird lovers! Your daily bird discovery is ready."
    "Time for today's bird adventure! Listen closely to nature's music."
    "Welcome to your daily bird journey! Let's discover who's singing today."
)

echo "🎤 Generating intros for voice: $VOICE_NAME (ID: $VOICE_ID)"
echo "📁 Output directory: $OUTPUT_DIR"
echo ""

# Generate each intro
for i in "${!INTROS[@]}"; do
    INTRO_TEXT="${INTROS[$i]}"
    OUTPUT_FILE="${OUTPUT_DIR}/intro_$(printf "%02d" $i)_${VOICE_NAME}.mp3"
    
    # Skip if file already exists
    if [ -f "$OUTPUT_FILE" ]; then
        echo "✅ Exists: $OUTPUT_FILE"
        continue
    fi
    
    echo "🎵 Generating intro $i: ${OUTPUT_FILE##*/}"
    
    # Call ElevenLabs API
    curl -s --request POST \
        --url "https://api.elevenlabs.io/v1/text-to-speech/$VOICE_ID" \
        --header "xi-api-key: $ELEVENLABS_API_KEY" \
        --header "Content-Type: application/json" \
        --data "{
            \"text\": \"$INTRO_TEXT\",
            \"model_id\": \"eleven_monolingual_v1\",
            \"voice_settings\": {
                \"stability\": 0.75,
                \"similarity_boost\": 0.75,
                \"style\": 0.5,
                \"use_speaker_boost\": true
            }
        }" \
        --output "$OUTPUT_FILE"
    
    # Check if file was created successfully
    if [ -f "$OUTPUT_FILE" ] && [ -s "$OUTPUT_FILE" ]; then
        # Get duration for verification
        DURATION=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$OUTPUT_FILE" 2>/dev/null || echo "unknown")
        echo "   ✅ Created: ${OUTPUT_FILE##*/} (${DURATION}s)"
    else
        echo "   ❌ Failed to create: ${OUTPUT_FILE##*/}"
        rm -f "$OUTPUT_FILE" # Remove empty file if created
    fi
    
    # Small delay to avoid rate limiting
    sleep 1
done

echo ""
echo "🎉 Generation complete!"
echo ""
echo "📊 Summary for $VOICE_NAME:"
ls -la "$OUTPUT_DIR"/intro_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l | xargs -I {} echo "   Total files: {}"

# Show file sizes and durations
echo ""
echo "📁 Generated files:"
for file in "$OUTPUT_DIR"/intro_*_"${VOICE_NAME}".mp3; do
    if [ -f "$file" ]; then
        DURATION=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$file" 2>/dev/null || echo "?")
        SIZE=$(ls -lh "$file" | awk '{print $5}')
        echo "   ${file##*/}: ${DURATION}s, ${SIZE}"
    fi
done

echo ""
echo "✅ Done! Intro files for $VOICE_NAME are ready."
echo ""
echo "Next steps:"
echo "1. Test the files: afplay $OUTPUT_DIR/intro_00_${VOICE_NAME}.mp3"
echo "2. Mix with nature sounds: go run cmd/mix_intros/main.go -all"