#!/bin/bash

# Generate pre-recorded outro files with proper joke timing
# This creates outros for each voice with jokes, wisdom, challenges, etc.

set -e

# Check for API key
if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo "❌ Error: ELEVENLABS_API_KEY not set"
    echo "Run: export ELEVENLABS_API_KEY='your_key'"
    exit 1
fi

VOICE_ID="${1:-ErXwobaYiN019PkySvjV}"
VOICE_NAME="${2:-Antoni}"
OUTPUT_DIR="final_outros"

mkdir -p "$OUTPUT_DIR"

echo "🎤 Generating Outros for $VOICE_NAME"
echo "=================================="

# Function to generate audio with SSML for pauses
generate_outro() {
    local TYPE="$1"
    local INDEX="$2"
    local TEXT="$3"
    local OUTPUT_FILE="${OUTPUT_DIR}/outro_${TYPE}_$(printf "%02d" $INDEX)_${VOICE_NAME}.mp3"
    
    if [ -f "$OUTPUT_FILE" ]; then
        echo "✅ Exists: ${OUTPUT_FILE##*/}"
        return
    fi
    
    echo "🎵 Generating: outro_${TYPE}_$(printf "%02d" $INDEX)_${VOICE_NAME}.mp3"
    
    # Call ElevenLabs with the text
    curl -s --request POST \
        --url "https://api.elevenlabs.io/v1/text-to-speech/$VOICE_ID" \
        --header "xi-api-key: $ELEVENLABS_API_KEY" \
        --header "Content-Type: application/json" \
        --data "{
            \"text\": \"$TEXT\",
            \"model_id\": \"eleven_monolingual_v1\",
            \"voice_settings\": {
                \"stability\": 0.75,
                \"similarity_boost\": 0.75
            }
        }" \
        --output "$OUTPUT_FILE"
    
    if [ -f "$OUTPUT_FILE" ] && [ -s "$OUTPUT_FILE" ]; then
        echo "   ✅ Created ($(ls -lh "$OUTPUT_FILE" | awk '{print $5}'))"
    else
        echo "   ❌ Failed"
        rm -f "$OUTPUT_FILE"
    fi
    
    sleep 0.5  # Rate limiting
}

# Generate joke outros with proper pause timing
echo ""
echo "📝 Generating Joke Outros (with pauses)..."
echo "----------------------------------------"

# Jokes with <break> tags for pauses (using ... as pause indicator)
JOKES_WITH_PAUSES=(
    "Here's today's giggle before you go! Why don't you ever see birds using Facebook? ... ... Because they already have Twitter! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! What do you call a bird that's afraid of heights? ... ... A chicken! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! Why do hummingbirds hum? ... ... Because they don't know the words! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! What's a bird's favorite type of math? ... ... Owl-gebra! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! Why did the pelican get kicked out of the restaurant? ... ... Because he had a very big bill! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! What do you call a very rude bird? ... ... A mockingbird! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! Why don't birds get lost? ... ... Because they always take the fly-way! See you tomorrow for another amazing bird adventure, explorers!"
    "Here's today's giggle before you go! What's a parrot's favorite game? ... ... Hide and speak! See you tomorrow for another amazing bird adventure, explorers!"
)

INDEX=0
for JOKE in "${JOKES_WITH_PAUSES[@]}"; do
    generate_outro "joke" "$INDEX" "$JOKE"
    ((INDEX++))
done

# Generate wisdom outros
echo ""
echo "📝 Generating Wisdom Outros..."
echo "-----------------------------"

WISDOM_OUTROS=(
    "Remember, little explorers: Just like birds, you have your own special song to sing. Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
    "Remember, little explorers: Every bird started as an egg, and look how far they've flown! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
    "Remember, little explorers: Birds teach us that morning is a gift - that's why they sing at dawn! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
    "Remember, little explorers: Even the tiniest hummingbird can fly across an ocean. You can do big things too! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
    "Remember, little explorers: Birds don't worry about falling - they trust their wings. Trust yourself too! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
)

INDEX=0
for WISDOM in "${WISDOM_OUTROS[@]}"; do
    generate_outro "wisdom" "$INDEX" "$WISDOM"
    ((INDEX++))
done

# Generate teaser outros
echo ""
echo "📝 Generating Teaser Outros..."
echo "-----------------------------"

TEASER_OUTROS=(
    "Wow, wasn't that amazing? Tomorrow we'll meet another incredible feathered friend! Will it be big or small? Colorful or camouflaged? You'll have to come back to find out! Keep your ears open for bird songs today, explorers!"
    "What an adventure! Tomorrow's bird might live in your backyard or fly thousands of miles each year! Can you guess which one? Come back tomorrow to discover another amazing bird! Listen for birds singing on your way today!"
    "That was fantastic! Tomorrow's mystery bird has a super special talent. Will it be the fastest? The smartest? The most colorful? Join me tomorrow to find out! Keep watching the sky today, bird detectives!"
)

INDEX=0
for TEASER in "${TEASER_OUTROS[@]}"; do
    generate_outro "teaser" "$INDEX" "$TEASER"
    ((INDEX++))
done

# Generate challenge outros
echo ""
echo "📝 Generating Challenge Outros..."
echo "--------------------------------"

CHALLENGE_OUTROS=(
    "Your weekend challenge: Can you spot three different birds before Monday? Look for different sizes, colors, and listen for different songs! Draw a picture of your favorite one! See you Monday with another amazing bird, explorers!"
    "Weekend bird mission: Find a cozy spot outside and stay as quiet as a mouse for two whole minutes. Count how many bird sounds you hear! Can you hear different songs? Report back on Monday, nature detectives!"
    "Super Saturday challenge: Can you move like three different birds today? Try hopping like a robin, gliding like an eagle, and pecking like a woodpecker! Show someone your best bird moves! See you Monday, little birds!"
)

INDEX=0
for CHALLENGE in "${CHALLENGE_OUTROS[@]}"; do
    generate_outro "challenge" "$INDEX" "$CHALLENGE"
    ((INDEX++))
done

# Generate fun fact outros
echo ""
echo "📝 Generating Fun Fact Outros..."
echo "-------------------------------"

FUN_FACT_OUTROS=(
    "Did you know some birds can sleep while flying? Arctic terns can take tiny naps in the air during their long journeys! What an amazing world of birds we live in! Come back tomorrow for another incredible discovery!"
    "Here's something amazing: A hummingbird's heart beats over 1,200 times per minute! That's 20 times every second! Wow! Nature is full of surprises. Join me tomorrow to learn about another fantastic bird!"
    "Fun fact of the day: Some birds can remember thousands of hiding places for their food! That's like remembering where you put toys in a thousand different toy boxes! See you tomorrow for more bird wonders!"
)

INDEX=0
for FUN_FACT in "${FUN_FACT_OUTROS[@]}"; do
    generate_outro "funfact" "$INDEX" "$FUN_FACT"
    ((INDEX++))
done

echo ""
echo "📊 Summary for $VOICE_NAME:"
echo "  Jokes:      $(ls "$OUTPUT_DIR"/outro_joke_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Wisdom:     $(ls "$OUTPUT_DIR"/outro_wisdom_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Teasers:    $(ls "$OUTPUT_DIR"/outro_teaser_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Challenges: $(ls "$OUTPUT_DIR"/outro_challenge_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Fun Facts:  $(ls "$OUTPUT_DIR"/outro_funfact_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo ""
echo "✅ Done! Run for other voices:"
echo "  ./scripts/generate_outros.sh <VOICE_ID> <VOICE_NAME>"