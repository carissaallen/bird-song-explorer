#!/bin/bash

# Helper script to replace a voice in the Bird Song Explorer system
# This updates the configuration and generates new intro files

set -e

echo "🎤 Bird Song Explorer - Voice Replacement Tool"
echo "=============================================="
echo ""

# Show current voices
echo "Current voices configured:"
echo ""
echo "1. Amelia    (British)      - ZF6FPAbjXT4488VcRRnw"
echo "2. Antoni    (American)     - ErXwobaYiN019PkySvjV"
echo "3. Charlotte (Australian)   - 5GZaeOOG7yqLdoTRsaa6"
echo "4. Peter     (Irish)        - E8tAm6nkbW2yKYAJLVXH"
echo "5. Drake     (Canadian)     - HYM6YgFANZinEBanknZK"
echo "6. Sally     (Southern US)  - XHqlxleHbYnK8xmft8Vq"
echo ""

# Instructions
echo "To replace a voice:"
echo ""
echo "STEP 1: Edit internal/config/voices.go"
echo "   Update the DefaultVoices array with new voice info:"
echo "   - ID: The ElevenLabs voice ID"
echo "   - Name: Voice name (used in filenames)"
echo "   - Region: Accent/region description"
echo "   - Language: Language code (en-US, en-GB, etc.)"
echo ""
echo "STEP 2: Generate intro files for the new voice"
echo "   Run: ./scripts/generate_intros_for_voice.sh <VOICE_ID> <VOICE_NAME>"
echo "   Example: ./scripts/generate_intros_for_voice.sh ErXwobaYiN019PkySvjV Antoni"
echo ""
echo "STEP 3: Remove old voice intro files (optional)"
echo "   rm final_intros/intro_*_OldVoiceName.mp3"
echo ""
echo "STEP 4: Mix with nature sounds (optional)"
echo "   go run cmd/mix_intros/main.go -all"
echo ""

# Check if we should proceed with replacement
read -p "Do you want to replace a voice now? (y/n): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Exiting..."
    exit 0
fi

# Get voice to replace
echo ""
read -p "Which voice NUMBER do you want to replace (1-6)? " VOICE_NUM

case $VOICE_NUM in
    1) OLD_NAME="Amelia" ;;
    2) OLD_NAME="Antoni" ;;
    3) OLD_NAME="Charlotte" ;;
    4) OLD_NAME="Peter" ;;
    5) OLD_NAME="Drake" ;;
    6) OLD_NAME="Sally" ;;
    *) echo "Invalid selection"; exit 1 ;;
esac

echo ""
echo "Replacing: $OLD_NAME"
echo ""

# Get new voice information
read -p "Enter new voice NAME (e.g., 'Jessica'): " NEW_NAME
read -p "Enter new voice ID from ElevenLabs: " NEW_ID
read -p "Enter voice region (e.g., 'British', 'American'): " NEW_REGION
read -p "Enter language code (e.g., 'en-US', 'en-GB'): " NEW_LANG

# Show what will be done
echo ""
echo "📋 Summary:"
echo "   Replacing: $OLD_NAME"
echo "   With: $NEW_NAME ($NEW_REGION, $NEW_LANG)"
echo "   Voice ID: $NEW_ID"
echo ""

read -p "Proceed? (y/n): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Update voices.go
echo ""
echo "⚠️  Please manually update internal/config/voices.go:"
echo ""
echo "Replace the entry for $OLD_NAME with:"
echo "{"
echo "    ID:       \"$NEW_ID\","
echo "    Name:     \"$NEW_NAME\","
echo "    Region:   \"$NEW_REGION\","
echo "    Language: \"$NEW_LANG\","
echo "},"
echo ""
echo "Press Enter when you've updated the file..."
read

# Check for API key
if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo ""
    echo "❌ ELEVENLABS_API_KEY not set!"
    echo "Please run: export ELEVENLABS_API_KEY='your_api_key'"
    echo "Then run: ./scripts/generate_intros_for_voice.sh $NEW_ID $NEW_NAME"
    exit 1
fi

# Generate new intros
echo ""
echo "🎵 Generating intro files for $NEW_NAME..."
./scripts/generate_intros_for_voice.sh "$NEW_ID" "$NEW_NAME"

# Ask about removing old files
echo ""
read -p "Remove old intro files for $OLD_NAME? (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing old files..."
    rm -f final_intros/intro_*_"${OLD_NAME}".mp3
    echo "✅ Old files removed"
fi

# Ask about mixing with nature sounds
echo ""
read -p "Mix new intros with nature sounds? (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Mixing with nature sounds..."
    go run cmd/mix_intros/main.go -all
fi

echo ""
echo "✅ Voice replacement complete!"
echo ""
echo "New voice $NEW_NAME is ready to use."
echo "Test with: afplay final_intros/intro_00_${NEW_NAME}.mp3"