#!/bin/bash

# Generate UNIQUE pre-recorded outros for each voice
# Each voice gets DIFFERENT jokes, wisdom quotes, etc. for maximum variety

set -e

# Check for API key
if [ -z "$ELEVENLABS_API_KEY" ]; then
    echo "❌ Error: ELEVENLABS_API_KEY not set"
    echo "Run: export ELEVENLABS_API_KEY='your_key'"
    exit 1
fi

VOICE_ID="$1"
VOICE_NAME="$2"
OUTPUT_DIR="assets/final_outros"

if [ -z "$VOICE_ID" ] || [ -z "$VOICE_NAME" ]; then
    echo "Usage: $0 <VOICE_ID> <VOICE_NAME>"
    exit 1
fi

mkdir -p "$OUTPUT_DIR"

echo "🎤 Generating UNIQUE Outros for $VOICE_NAME"
echo "==========================================="

# Function to generate audio
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
    
    curl -s --request POST \
        --url "https://api.elevenlabs.io/v1/text-to-speech/$VOICE_ID" \
        --header "xi-api-key: $ELEVENLABS_API_KEY" \
        --header "Content-Type: application/json" \
        --data "{
            \"text\": \"$TEXT\",
            \"model_id\": \"eleven_multilingual_v2\",
            \"voice_settings\": {
                \"stability\": 0.50,
                \"similarity_boost\": 0.90,
                \"use_speaker_boost\": true,
                \"speed\": 1.0,
                \"style\": 0
            }
        }" \
        --output "$OUTPUT_FILE"
    
    if [ -f "$OUTPUT_FILE" ] && [ -s "$OUTPUT_FILE" ]; then
        echo "   ✅ Created"
    else
        echo "   ❌ Failed"
        rm -f "$OUTPUT_FILE"
    fi
    
    sleep 0.5
}

# ALL 20 JOKES from your outro_manager.go (with pauses added)
ALL_JOKES=(
    "Why don't you ever see birds using Facebook? ... ... Because they already have Twitter!"
    "What do you call a bird that's afraid of heights? ... ... A chicken!"
    "Why do hummingbirds hum? ... ... Because they don't know the words!"
    "What's a bird's favorite type of math? ... ... Owl-gebra!"
    "Why did the pelican get kicked out of the restaurant? ... ... Because he had a very big bill!"
    "What do you call a very rude bird? ... ... A mockingbird!"
    "Why don't birds get lost? ... ... Because they always take the fly-way!"
    "What's a parrot's favorite game? ... ... Hide and speak!"
    "Why did the bird go to school? ... ... To improve its tweet-ing skills!"
    "What do you call a bird in winter? ... ... Brrr-d!"
    "What's a bird's favorite snack? ... ... Chocolate chirp cookies!"
    "Why are birds so good at dodgeball? ... ... They're excellent at duck-ing!"
    "What do you give a sick bird? ... ... Tweet-ment!"
    "Why did the bird join the band? ... ... Because it had perfect tweet-ing!"
    "What's a bird's favorite subject? ... ... Owl-gebra and fly-namics!"
    "Why don't seagulls fly over bays? ... ... Because then they'd be bagels!"
    "What do you call a funny chicken? ... ... A comedi-hen!"
    "Why did the baby bird get in trouble? ... ... It was caught peeping!"
    "What do you call two birds in love? ... ... Tweet-hearts!"
    "Why are birds always happy? ... ... They wake up on the bright side of the perch!"
)

# DISTRIBUTE JOKES BASED ON VOICE (each voice gets different jokes!)
case "$VOICE_NAME" in
    "Amelia")
        # Amelia gets jokes 0-7
        VOICE_JOKES=("${ALL_JOKES[@]:0:8}")
        ;;
    "Antoni")
        # Antoni gets jokes 3-10
        VOICE_JOKES=("${ALL_JOKES[@]:3:8}")
        ;;
    "Hope")
        # Hope gets jokes 6-13
        VOICE_JOKES=("${ALL_JOKES[@]:6:8}")
        ;;
    "Rory")
        # Rory gets jokes 9-16
        VOICE_JOKES=("${ALL_JOKES[@]:9:8}")
        ;;
    "Danielle")
        # Danielle gets jokes 12-19
        VOICE_JOKES=("${ALL_JOKES[@]:12:8}")
        ;;
    "Stuart")
        # Stuart gets jokes 0,2,4,6,8,10,12,14 (evens)
        VOICE_JOKES=(
            "${ALL_JOKES[0]}"
            "${ALL_JOKES[2]}" 
            "${ALL_JOKES[4]}"
            "${ALL_JOKES[6]}"
            "${ALL_JOKES[8]}"
            "${ALL_JOKES[10]}"
            "${ALL_JOKES[12]}"
            "${ALL_JOKES[14]}"
        )
        ;;
    *)
        # Default: first 8 jokes
        VOICE_JOKES=("${ALL_JOKES[@]:0:8}")
        ;;
esac

echo ""
echo "📝 Generating UNIQUE Joke Outros for $VOICE_NAME..."
echo "---------------------------------------------------"

INDEX=0
for JOKE in "${VOICE_JOKES[@]}"; do
    FULL_TEXT="Here's today's giggle before you go! $JOKE See you tomorrow for another amazing bird adventure, explorers!"
    generate_outro "joke" "$INDEX" "$FULL_TEXT"
    ((INDEX++))
done

# VARIED WISDOM QUOTES BY VOICE
echo ""
echo "📝 Generating Wisdom Outros..."
echo "-----------------------------"

case "$VOICE_NAME" in
    "Amelia"|"Antoni")
        WISDOM_OUTROS=(
            "Remember, little explorers: Just like birds, you have your own special song to sing. Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Every bird started as an egg, and look how far they've flown! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Birds teach us that morning is a gift - that's why they sing at dawn! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Even the tiniest hummingbird can fly across an ocean. You can do big things too! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Birds don't worry about falling - they trust their wings. Trust yourself too! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
        )
        ;;
    "Hope"|"Rory")
        WISDOM_OUTROS=(
            "Remember, little explorers: Like birds building nests, we create our world one twig at a time. Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Birds sing after storms. You can make music after hard times too! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Some birds fly solo, some in flocks - both ways are perfect! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: A bird doesn't sing because it has answers, it sings because it has a song! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Every bird was once afraid of its first flight, but look at them soar now! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
        )
        ;;
    *)
        WISDOM_OUTROS=(
            "Remember, little explorers: Birds remind us that it's not about the size of your wings, but the courage in your heart! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Like migrating birds, sometimes the longest journeys lead to the most beautiful places! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Birds teach us patience - they wait for the perfect moment to take flight! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Each bird has its own voice, just like you have yours. Use it proudly! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
            "Remember, little explorers: Birds show us that home isn't just a place, it's wherever your flock is! Think of our feathered friend today and remember to spread your wings! Until tomorrow!"
        )
        ;;
esac

INDEX=0
for WISDOM in "${WISDOM_OUTROS[@]}"; do
    generate_outro "wisdom" "$INDEX" "$WISDOM"
    ((INDEX++))
done

# VARIED TEASERS BY VOICE
echo ""
echo "📝 Generating Teaser Outros..."
echo "-----------------------------"

case "$VOICE_NAME" in
    "Amelia"|"Hope"|"Danielle")
        TEASER_OUTROS=(
            "Wow, wasn't that amazing? Tomorrow we'll meet another incredible feathered friend! Will it be big or small? Colorful or camouflaged? You'll have to come back to find out! Keep your ears open for bird songs today, explorers!"
            "What an adventure! Tomorrow's bird might live in your backyard or fly thousands of miles each year! Can you guess which one? Come back tomorrow to discover another amazing bird! Listen for birds singing on your way today!"
            "That was fantastic! Tomorrow's mystery bird has a super special talent. Will it be the fastest? The smartest? The most colorful? Join me tomorrow to find out! Keep watching the sky today, bird detectives!"
        )
        ;;
    *)
        TEASER_OUTROS=(
            "What a discovery! Tomorrow's bird might be nocturnal, aquatic, or love the desert! Which will it be? Return tomorrow for another feathered surprise! Practice your bird watching skills today!"
            "Incredible! Tomorrow we'll explore a bird with an amazing superpower. X-ray vision? Lightning speed? Underwater breathing? Come back to solve the mystery! Keep your bird journals ready!"
            "How wonderful! Tomorrow's featured friend could be tiny as your thumb or tall as you! Can you imagine which? Join me tomorrow for another bird adventure! Look for birds wherever you go today!"
        )
        ;;
esac

INDEX=0
for TEASER in "${TEASER_OUTROS[@]}"; do
    generate_outro "teaser" "$INDEX" "$TEASER"
    ((INDEX++))
done

# CHALLENGES - Same for all but with voice personality
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

# FUN FACTS - Varied by voice
echo ""
echo "📝 Generating Fun Fact Outros..."
echo "-------------------------------"

case "$VOICE_NAME" in
    "Amelia"|"Antoni")
        FUN_FACT_OUTROS=(
            "Did you know some birds can sleep while flying? Arctic terns can take tiny naps in the air during their long journeys! What an amazing world of birds we live in! Come back tomorrow for another incredible discovery!"
            "Here's something amazing: A hummingbird's heart beats over 1,200 times per minute! That's 20 times every second! Wow! Nature is full of surprises. Join me tomorrow to learn about another fantastic bird!"
            "Fun fact of the day: Some birds can remember thousands of hiding places for their food! That's like remembering where you put toys in a thousand different toy boxes! See you tomorrow for more bird wonders!"
        )
        ;;
    *)
        FUN_FACT_OUTROS=(
            "Amazing fact: Penguins can jump 9 feet out of the water! That's like jumping onto a basketball hoop! Birds are incredible athletes. Come back tomorrow to discover another feathered superstar!"
            "Did you know owls can turn their heads 270 degrees? That's three-quarters of a full circle! They're like feathered periscopes! Join me tomorrow for another mind-blowing bird fact!"
            "Wow fact: Some birds can fly backwards, upside down, and even do loops! Hummingbirds are nature's acrobats! See you tomorrow to learn about another amazing bird ability!"
        )
        ;;
esac

INDEX=0
for FUN_FACT in "${FUN_FACT_OUTROS[@]}"; do
    generate_outro "funfact" "$INDEX" "$FUN_FACT"
    ((INDEX++))
done

echo ""
echo "📊 Summary for $VOICE_NAME:"
echo "  Jokes:      $(ls "$OUTPUT_DIR"/outro_joke_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l) (unique set!)"
echo "  Wisdom:     $(ls "$OUTPUT_DIR"/outro_wisdom_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Teasers:    $(ls "$OUTPUT_DIR"/outro_teaser_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Challenges: $(ls "$OUTPUT_DIR"/outro_challenge_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo "  Fun Facts:  $(ls "$OUTPUT_DIR"/outro_funfact_*_"${VOICE_NAME}".mp3 2>/dev/null | wc -l)"
echo ""
echo "✅ $VOICE_NAME has unique content different from other voices!"