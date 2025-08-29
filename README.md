# Bird Song Explorer 🎵

A smart Yoto card that brings local birds to your child's speaker every day. Wake up to the actual birds singing in your neighborhood, learn fun facts, and develop a connection with nature—all through the magic of a Yoto card.

## What It Does

Every morning, this Yoto card updates with a new bird from your area. Your child will hear:
1. A friendly introduction from one of six rotating narrators
2. Today's featured bird announcement  
3. The actual bird's song recorded in nature
4. Fun, kid-friendly facts about the bird
5. A playful outro with jokes or tomorrow's teaser

The birds are real birds recently spotted near you, pulled from eBird's citizen science database. If you're in London, you might hear a Robin. In California, perhaps a Scrub-Jay. The card adapts to wherever you are in the world.

## For Parents

This is a Make Your Own (MYO) card that updates automatically. Once set up, it requires no maintenance—just pop it in your Yoto player and enjoy a new bird each day. The content is specifically designed for ages 3-8, with gentle narration and engaging facts that grow curiosity without overwhelming young minds.

## For Developers

Built in Go, deployed on Google Cloud Run, scheduled with Cloud Scheduler. The system combines multiple APIs to create a seamless experience:

### Tech Stack
- **Location Detection**: IP geolocation with device timezone fallback
- **Bird Data**: eBird API for observations, Xeno-canto for recordings
- **Educational Content**: Wikipedia and iNaturalist for facts
- **Voice Generation**: ElevenLabs with smooth transitions between tracks
- **Platform**: Yoto API for card updates

### Quick Start

1. Clone and configure:
   ```bash
   cp .env.example .env
   # Add your API keys to .env
   ```

2. Required API Keys:
   - **Yoto**: From your MYO dashboard (client ID & secret)
   - **eBird**: Free at https://ebird.org/api/keygen
   - **ElevenLabs**: For voice generation 

3. Run locally:
   ```bash
   go run cmd/server/main.go
   ```

4. Deploy to Google Cloud:
   ```bash
   gcloud run deploy bird-song-explorer \
     --source . \
     --region us-central1
   ```

### Key Endpoints

- `POST /daily-update` - Triggered by Cloud Scheduler at 6 AM
- `POST /api/v1/yoto/webhook` - Handles card play events
- `GET /health` - Service health check

## Configuration

The card automatically:
- Detects location from IP or device timezone
- Selects appropriate regional birds
- Rotates through 6 different voice narrators
- Falls back to common birds if no local observations exist

See `.env.example` for all configuration options.

## Development

```bash
# Run with hot reload
air

# Build for production
go build -o bird-song-explorer cmd/server/main.go

# Run tests
go test ./...
```

## License

MIT - Feel free to adapt this for your own Yoto adventures! 🐦
