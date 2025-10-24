# Global Narration Files

This directory contains time-of-day specific introductions and outros that are used across all bird presentations.

## Directory Structure

```
global_narration/
├── morning/
│   ├── intro/     # Morning introduction narrations (5am-11:59am)
│   └── outro/     # Morning outro narrations
├── afternoon/
│   ├── intro/     # Afternoon introduction narrations (12pm-5:59pm)
│   └── outro/     # Afternoon outro narrations
└── evening/
    ├── intro/     # Evening introduction narrations (6pm-4:59am)
    └── outro/     # Evening outro narrations
```

## File Format

Each directory can contain multiple variations of narrations:
- `narration_01.mp3` - Primary version
- `narration_02.mp3` - Alternative version
- `narration_03.mp3` - Another variation

## Time-Based Selection

The system automatically selects the appropriate narration based on the user's local time:
- **Morning** (5:00 AM - 11:59 AM): Bright, energetic greetings
- **Afternoon** (12:00 PM - 5:59 PM): Warm, midday greetings
- **Evening** (6:00 PM - 4:59 AM): Calm, peaceful greetings

## Content Guidelines

### Intro Narrations
- Welcome users to Bird Song Explorer
- Set appropriate mood for time of day
- Duration: 10-20 seconds
- Should flow naturally into bird announcement

### Outro Narrations
- Thank users for listening
- Encourage exploration and observation
- Time-appropriate farewell
- Duration: 10-15 seconds

## Examples

### Morning Intro
"Good morning, little explorers! The birds are singing their dawn chorus just for you. Let's discover today's special feathered friend!"

### Afternoon Intro
"Hello, nature adventurers! The afternoon sun brings us another amazing bird to explore. Listen carefully to its beautiful song!"

### Evening Intro
"Good evening, young naturalists. As the day winds down, let's meet a wonderful bird that might be singing outside your window right now."