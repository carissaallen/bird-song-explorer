#!/bin/bash

# Ambience Continuity Test Script
# Tests the seamless ambience transition between Track 1 and Track 2

echo "🎵 Ambience Continuity Test"
echo "============================"
echo

# Check if .env file exists and source it
if [ -f .env ]; then
    echo "Loading environment variables from .env..."
    set -a
    source .env
    set +a
else
    echo "⚠️  Warning: .env file not found"
fi

# Check if ELEVENLABS_API_KEY is set
if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo "❌ Error: ELEVENLABS_API_KEY is not set"
    exit 1
fi

# Check if sound_effects directory exists
if [ ! -d "sound_effects" ]; then
    echo "❌ Error: sound_effects directory not found"
    exit 1
fi

# Run the test
echo "Testing ambience continuity between tracks..."
echo
go run test_ambience_continuity.go

# List generated test files
echo
echo "Generated test files:"
ls -la test_track*.mp3 test_combined*.mp3 2>/dev/null || echo "No test files found"