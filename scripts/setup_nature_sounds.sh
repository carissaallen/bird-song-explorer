#!/bin/bash

# Setup nature sounds for intro mixing
# This script creates the nature sounds directory and provides guidance on obtaining sounds

NATURE_SOUNDS_DIR="assets/nature_sounds"

echo "🌿 Setting up nature sounds directory..."

# Create directory if it doesn't exist
mkdir -p "$NATURE_SOUNDS_DIR"

echo "📁 Created directory: $NATURE_SOUNDS_DIR"
echo ""
echo "🎵 You'll need to add the following nature sound files:"
echo "  - forest_ambience.mp3 (general forest sounds)"
echo "  - morning_birds.mp3 (dawn chorus)"
echo "  - gentle_rain.mp3 (soft rain sounds)"
echo "  - wind_through_trees.mp3 (gentle wind)"
echo "  - babbling_brook.mp3 (stream sounds)"
echo "  - meadow_sounds.mp3 (open field ambience)"
echo "  - night_crickets.mp3 (evening sounds)"
echo ""
echo "🆓 Free nature sound sources:"
echo "  - Freesound.org (CC licensed sounds)"
echo "  - Zapsplat.com (free with account)"
echo "  - BBC Sound Effects (16,000+ free sounds)"
echo "  - Pixabay (royalty-free sounds)"
echo ""
echo "💡 Tips:"
echo "  - Keep files under 1MB for intro mixing"
echo "  - Use 30-60 second loops"
echo "  - Normalize audio levels before adding"
echo "  - Consider seasonal variations"

# Create a sample README in the directory
cat > "$NATURE_SOUNDS_DIR/README.md" << 'EOF'
# Nature Sounds for Bird Song Explorer

This directory contains ambient nature sounds used for mixing with introduction tracks.

## Required Files

- `forest_ambience.mp3` - General forest/woodland sounds
- `morning_birds.mp3` - Dawn chorus, morning bird songs
- `gentle_rain.mp3` - Soft rain, no thunder
- `wind_through_trees.mp3` - Gentle wind sounds
- `babbling_brook.mp3` - Stream or creek sounds
- `meadow_sounds.mp3` - Open field ambience
- `night_crickets.mp3` - Evening/night sounds

## Audio Specifications

- Format: MP3
- Duration: 30-60 seconds (will be looped)
- Bitrate: 128-192 kbps
- Sample Rate: 44.1 kHz
- Channels: Stereo preferred

## Usage

These sounds are automatically selected based on time of day or can be specified manually when mixing intros.

## Licensing

Ensure all sounds are properly licensed for use in your application.
Recommended sources: Freesound.org (CC), BBC Sound Effects, Pixabay
EOF

echo "📝 Created README in $NATURE_SOUNDS_DIR"
echo "✅ Setup complete!"