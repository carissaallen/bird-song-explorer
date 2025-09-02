<p align="center">
<img src="assets/cover/custom_cover_transparent_bkg.png" alt="Bird Song Explorer Yoto Cover" width="300"/>
</p>

# Bird Song Explorer 

> A smart Yoto card that brings local birds to your child's speaker every day. Wake up to the actual birds singing in your neighborhood, learn fun facts, and develop a connection with nature—all through the magic of a Yoto card.

## What It Does

Every morning, this Yoto card automatically updates with a new bird from your area. Your child will hear:

### Track Structure
1. **Introduction** - A warm welcome with nature sounds from a variety of narrators from different regions
2. **Today's Bird** - Announcement of the featured bird with ambient nature sounds
3. **Bird Song** - The actual bird's song recorded in nature  
4. **Bird Explorer's Guide** - Fun, educational facts about the bird tailored for young explorers
5. **See You Tomorrow!** - A playful outro with jokes, fun facts, or tomorrow's teaser

The birds are real birds recently spotted near you, pulled from eBird's citizen science database. If you're in London, you might hear a Robin. In California, perhaps a Scrub-Jay. The card adapts to wherever you are in the world.

## For Parents

This is a Make Your Own (MYO) card that updates automatically. Once set up, it requires no maintenance—just pop it in your Yoto player and enjoy a new bird each day. The content is specifically designed for ages 3-8, with:

- **Variety of narrators** from different regions bringing diverse accents and storytelling styles
- **Enhanced audio** with layered nature sounds, ambient backgrounds, and smooth transitions
- **Educational content** that's both fun and factual, adapted from trusted sources
- **Smart location detection** that finds birds actually spotted in your area
- **Daily variety** ensuring your child discovers a different local bird each day

## For Developers

Built in Go, deployed on Google Cloud Run, scheduled with Cloud Scheduler. The system combines multiple APIs to create a seamless experience:

### Tech Stack
- **Language**: Go 1.21+ with modular service architecture
- **Deployment**: Google Cloud Run (serverless container platform)
- **Scheduling**: Cloud Scheduler for daily updates
- **Location Detection**: IP geolocation with device timezone fallback
- **Bird Data**: eBird API for real-time observations, Xeno-canto for high-quality recordings
- **Educational Content**: Wikipedia and iNaturalist integration for child-friendly facts
- **Voice Generation**: ElevenLabs TTS with optimized pacing and natural pauses
- **Audio Processing**: FFmpeg for mixing, layering effects, and smooth transitions
- **Platform Integration**: Yoto API for MYO card updates

### Quick Start

1. Clone and configure:
   ```bash
   cp .env.example .env
   # Add your API keys to .env
   ```

2. Required API Keys:
   - **Yoto**: From your MYO dashboard (client ID)
   - **eBird**: Free at https://ebird.org/api/keygen
   - **ElevenLabs**: For voice generation (required for TTS)

3. Run locally:
   ```bash
   go run cmd/server/main.go
   ```

4. Deploy to Google Cloud:
   ```bash
   # Using the deployment script with environment variables
   ./deploy/deploy_with_env.sh
   
   # Or manually with gcloud
   gcloud run deploy bird-song-explorer \
     --source . \
     --region us-central1 \
     --set-env-vars-from-file=.env
   ```

### Key Endpoints

- `POST /daily-update` - Triggered by Cloud Scheduler for daily updates
- `POST /api/v1/yoto/webhook` - Handles card play events and personalization
- `GET /health` - Service health check
- `POST /api/v1/yoto/oauth/callback` - OAuth callback for Yoto authentication

## Features

### Audio Enhancement
- **Layered soundscapes**: Nature ambience that fades between tracks
- **Dynamic mixing**: Chimes, transitions, and effects for engaging listening
- **Optimized pacing**: Child-friendly speech rates with natural pauses
- **Volume normalization**: Consistent audio levels across all tracks

### Smart Content Generation  
- **Location-aware**: Uses IP geolocation and timezone detection
- **Fallback logic**: Common birds when no local observations exist
- **Rotating narrators**: Variety of voices from different regions
- **Personalized facts**: Location-specific information when available
- **Daily rotation**: Ensures a new bird every day

## Development

### Development Commands
```bash
# Install dependencies
go mod download

# Run locally with hot reload
air

# Build for production
go build -o bird-song-explorer cmd/server/main.go

# Run tests
go test ./...

# Deploy to Google Cloud
./deploy/deploy_with_env.sh
```

## Troubleshooting

### Common Issues

**No birds found for location**
- The system will fall back to common birds automatically
- Check eBird has observations for your region
- Verify location detection is working via logs

**Audio quality issues**
- Ensure FFmpeg is available in your deployment environment
- Check that sound effects files are included in Docker image
- Verify ElevenLabs API quota hasn't been exceeded

## Contributing
Contributions are welcome! Areas for improvement:
- Additional bird fact sources
- More nature sound variations
- Support for different languages
- Enhanced educational content
- Regional bird priorities

Please open an issue first to discuss major changes.

## License

MIT - Feel free to adapt this for your own Yoto adventures! 

---

*Built with love for young explorers and their Yoto players. May every morning bring the wonder of nature to your home.* 🐦
