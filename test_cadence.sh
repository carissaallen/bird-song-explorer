#!/bin/bash

# Speech Cadence Test Script
# Tests the improved speech cadence with pauses

echo "🎵 Speech Cadence Improvement Test"
echo "==================================="
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

# Run the test
echo "Testing speech cadence improvements..."
echo
go run test_cadence_improvements.go

# List generated test files
echo
echo "Generated test files:"
ls -la test_announcement_*.mp3 test_description_*.mp3 2>/dev/null || echo "No test files found"

echo
echo "🎧 Compare the audio files to hear the difference:"
echo "   • OLD versions: Run-on sentences without pauses"
echo "   • NEW versions: Natural pauses between sentences"