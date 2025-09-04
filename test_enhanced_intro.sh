#!/bin/bash

# Enhanced Intro Test Script
# This script tests the new enhanced intro generation with local sound effects

echo "🎵 Enhanced Intro Test Script"
echo "============================="
echo

# Check if .env file exists and source it
if [ -f .env ]; then
    echo "Loading environment variables from .env..."
    set -a
    source .env
    set +a
else
    echo "⚠️  Warning: .env file not found"
    echo "Make sure ELEVENLABS_API_KEY is set in your environment"
fi

# Check if ELEVENLABS_API_KEY is set
if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo "❌ Error: ELEVENLABS_API_KEY is not set"
    echo "Please set it in your .env file or environment"
    exit 1
fi

# Check if sound_effects directory exists
if [ ! -d "sound_effects" ]; then
    echo "❌ Error: sound_effects directory not found"
    echo "Please ensure you have the sound_effects folder with:"
    echo "  - ambience/forest-ambience.mp3"
    echo "  - ambience/jungle_sounds.mp3"
    echo "  - ambience/morning-birdsong.mp3"
    echo "  - chimes/sparkle_chime.mp3"
    exit 1
fi

# Run the test
echo "Running enhanced intro test..."
echo
go run test_enhanced_intro.go

# Check if any test files were created and list them
echo
echo "Test files created:"
ls -la test_enhanced_intro_*.mp3 2>/dev/null || echo "No test files found"