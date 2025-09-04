package services

import (
	"fmt"
	"math/rand"
	"time"
)

// OutroManager handles generation of outro content
type OutroManager struct {
	generalJokes      []string
	specificBirdJokes map[string]string
	wisdomQuotes      []string
	funFacts          []string
}

// NewOutroManager creates a new outro manager
func NewOutroManager() *OutroManager {
	return &OutroManager{
		generalJokes:      generalBirdJokes,
		specificBirdJokes: specificJokes,
		wisdomQuotes:      birdWisdom,
		funFacts:          birdFunFacts,
	}
}

// GenerateOutroText generates the appropriate outro text based on day and bird
func (om *OutroManager) GenerateOutroText(birdName string, dayOfWeek time.Weekday) string {
	outroType := om.getOutroType(dayOfWeek)

	var baseOutro string
	switch outroType {
	case "joke":
		baseOutro = om.getJokeOutro(birdName)
	case "wisdom":
		baseOutro = om.getWisdomOutro(birdName)
	case "teaser":
		baseOutro = om.getTeaserOutro(birdName)
	case "challenge":
		baseOutro = om.getChallengeOutro(birdName)
	case "funfact":
		baseOutro = om.getFunFactOutro(birdName)
	default:
		baseOutro = om.getTeaserOutro(birdName)
	}

	// Add seasonal awareness
	seasonalAddition := om.getSeasonalAddition()
	if seasonalAddition != "" {
		baseOutro = baseOutro + " " + seasonalAddition
	}

	return baseOutro
}

// getOutroType determines which type of outro to use based on the day
func (om *OutroManager) getOutroType(dayOfWeek time.Weekday) string {
	switch dayOfWeek {
	case time.Monday, time.Friday:
		return "joke"
	case time.Tuesday, time.Thursday:
		return "teaser"
	case time.Wednesday:
		return "wisdom"
	case time.Saturday:
		return "challenge"
	case time.Sunday:
		return "funfact"
	default:
		return "teaser"
	}
}

// getJokeOutro returns a joke-based outro
func (om *OutroManager) getJokeOutro(birdName string) string {
	var joke string

	// Check for bird-specific joke first
	if specificJoke, exists := om.specificBirdJokes[birdName]; exists {
		joke = specificJoke
	} else {
		// Use general joke
		rand.Seed(time.Now().UnixNano())
		joke = om.generalJokes[rand.Intn(len(om.generalJokes))]
	}

	return fmt.Sprintf("Here's today's giggle before you go! <break time=\"0.5s\" /> %s <break time=\"1.0s\" /> See you tomorrow for another amazing bird adventure, explorers!", joke)
}

// getWisdomOutro returns a wisdom/inspirational outro
func (om *OutroManager) getWisdomOutro(birdName string) string {
	rand.Seed(time.Now().UnixNano())
	wisdom := om.wisdomQuotes[rand.Intn(len(om.wisdomQuotes))]

	return fmt.Sprintf("Remember, little explorers: <break time=\"0.5s\" /> %s <break time=\"1.0s\" /> Think of our %s friend today and remember to spread your wings! <break time=\"0.5s\" /> Until tomorrow!", wisdom, birdName)
}

// getTeaserOutro returns a tomorrow teaser outro
func (om *OutroManager) getTeaserOutro(birdName string) string {
	return fmt.Sprintf("Wow, wasn't the %s amazing? <break time=\"0.5s\" /> Tomorrow we'll meet another incredible feathered friend! <break time=\"0.5s\" /> Will it be big or small? <break time=\"0.3s\" /> Colorful or camouflaged? <break time=\"0.5s\" /> You'll have to come back to find out! <break time=\"0.5s\" /> Keep your ears open for bird songs today, explorers!", birdName)
}

// getChallengeOutro returns an interactive challenge outro
func (om *OutroManager) getChallengeOutro(birdName string) string {
	challenges := []string{
		fmt.Sprintf("Can you make the %s's sound three times today? Try it at breakfast, lunch, and dinner!", birdName),
		fmt.Sprintf("Can you spot a bird outside your window today? See if it moves like our %s friend!", birdName),
		fmt.Sprintf("Can you draw a picture of the %s you heard today? Show someone special your artwork!", birdName),
		fmt.Sprintf("Can you flap your arms like the %s? Count how many flaps you can do!", birdName),
	}

	rand.Seed(time.Now().UnixNano())
	challenge := challenges[rand.Intn(len(challenges))]

	return fmt.Sprintf("Your Bird Explorer Challenge: <break time=\"0.5s\" /> %s <break time=\"1.0s\" /> Tomorrow, we'll learn about a new bird together. <break time=\"0.5s\" /> Happy exploring!", challenge)
}

// getFunFactOutro returns a fun fact outro
func (om *OutroManager) getFunFactOutro(birdName string) string {
	rand.Seed(time.Now().UnixNano())
	fact := om.funFacts[rand.Intn(len(om.funFacts))]

	return fmt.Sprintf("Before you go, did you know? <break time=\"0.5s\" /> %s <break time=\"1.0s\" /> Amazing, right? <break time=\"0.5s\" /> Sweet dreams, and tomorrow we'll discover another incredible bird together!", fact)
}

// getSeasonalAddition returns a seasonal message based on the current time of year
func (om *OutroManager) getSeasonalAddition() string {
	now := time.Now()
	month := now.Month()

	// Determine season (Northern Hemisphere)
	var seasonalMessages []string

	switch {
	case month >= 3 && month <= 5:
		seasonalMessages = []string{
			"Spring is here, and birds are building their nests!",
			"Listen for baby birds chirping this spring!",
			"The birds are extra happy because spring has arrived!",
			"Spring brings new bird songs to discover!",
			"Birds are singing their spring melodies!",
		}
	case month >= 6 && month <= 8:
		seasonalMessages = []string{
			"Summer is perfect for bird watching adventures!",
			"Birds wake up early in summer, just like the sun!",
			"Enjoy the long summer days full of bird songs!",
			"Summer birds are teaching their babies to fly!",
			"The warm summer air carries bird songs far and wide!",
		}
	case month >= 9 && month <= 11:
		seasonalMessages = []string{
			"Some birds are getting ready for their autumn journey south!",
			"Watch for birds gathering seeds for the colder days ahead!",
			"The autumn leaves aren't the only things changing - listen for new bird visitors!",
			"Fall is here, and birds are preparing for their big adventures!",
			"Autumn birds are extra busy getting ready for winter!",
		}
	default: // December, January, February
		seasonalMessages = []string{
			"Even in winter, brave birds keep singing their songs!",
			"Winter birds fluff up their feathers like cozy jackets!",
			"Some special birds love the winter cold!",
			"Birds huddle together to stay warm in winter!",
			"Winter is when we can really appreciate our year-round bird friends!",
		}
	}

	// Special holiday messages
	if month == 12 && now.Day() >= 20 && now.Day() <= 31 {
		seasonalMessages = append(seasonalMessages,
			"Happy holidays! Birds celebrate by singing extra sweetly!",
			"'Tis the season for bird songs and joy!",
		)
	}

	// Special messages for migration seasons
	if month == 3 || month == 4 || month == 9 || month == 10 {
		seasonalMessages = append(seasonalMessages,
			"It's migration season - watch for traveling birds!",
			"Birds are on the move during this special migration time!",
		)
	}

	rand.Seed(time.Now().UnixNano())
	return seasonalMessages[rand.Intn(len(seasonalMessages))]
}

// getCurrentSeason returns the current season as a string
func getCurrentSeason() string {
	month := time.Now().Month()
	switch {
	case month >= 3 && month <= 5:
		return "spring"
	case month >= 6 && month <= 8:
		return "summer"
	case month >= 9 && month <= 11:
		return "autumn"
	default:
		return "winter"
	}
}

// Bird jokes collection
var generalBirdJokes = []string{
	"Why don't you ever see birds using Facebook? Because they already have Twitter!",
	"What do you call a bird that's afraid of heights? A chicken!",
	"Why do hummingbirds hum? Because they don't know the words!",
	"What's a bird's favorite type of math? Owl-gebra!",
	"Why did the pelican get kicked out of the restaurant? Because he had a very big bill!",
	"What do you call a very rude bird? A mockingbird!",
	"Why don't birds get lost? Because they always take the fly-way!",
	"What's a parrot's favorite game? Hide and speak!",
	"Why did the bird go to school? To improve its tweet-ing skills!",
	"What do you call a bird in winter? Brrr-d!",
	"What's a bird's favorite snack? Chocolate chirp cookies!",
	"Why are birds so good at dodgeball? They're excellent at duck-ing!",
	"What do you give a sick bird? Tweet-ment!",
	"Why did the bird join the band? Because it had perfect tweet-ing!",
	"What's a bird's favorite subject? Owl-gebra and fly-namics!",
	"Why don't seagulls fly over bays? Because then they'd be bagels!",
	"What do you call a funny chicken? A comedi-hen!",
	"Why did the baby bird get in trouble? It was caught peeping!",
	"What do you call two birds in love? Tweet-hearts!",
	"Why are birds always happy? They wake up on the bright side of the perch!",
}

// Bird-specific jokes
var specificJokes = map[string]string{
	"Owl":               "What do you call an owl magician? Hoo-dini!",
	"Great Horned Owl":  "What do you call an owl magician? Hoo-dini!",
	"Barred Owl":        "Knock knock! Who's there? Owl. Owl who? Owl tell you another joke tomorrow!",
	"Woodpecker":        "Why don't woodpeckers ever forget? They always remember to knock!",
	"Duck":              "What time do ducks wake up? At the quack of dawn!",
	"Mallard":           "What time do ducks wake up? At the quack of dawn!",
	"Crow":              "What's a crow's favorite drink? Caw-fee!",
	"American Crow":     "What's a crow's favorite drink? Caw-fee!",
	"Turkey":            "Why did the turkey cross the road? To prove it wasn't chicken!",
	"Wild Turkey":       "Why did the turkey cross the road? To prove it wasn't chicken!",
	"Eagle":             "Why don't eagles like fast food? Because they can't catch it!",
	"Bald Eagle":        "Why don't eagles like fast food? Because they can't catch it!",
	"Robin":             "What's a robin's favorite chocolate? Nestle!",
	"American Robin":    "What's a robin's favorite chocolate? Nestle!",
	"Blue Jay":          "Why are Blue Jays so good at baseball? They're always stealing!",
	"Cardinal":          "Why don't cardinals ever get lost? They always know which way is north!",
	"Northern Cardinal": "Why don't cardinals ever get lost? They always know which way is north!",
	"Penguin":           "What's a penguin's favorite relative? Aunt Arctica!",
	"Flamingo":          "Why don't flamingos like online shopping? They prefer to stand in stores!",
	"Peacock":           "What do you call a peacock that won't stop talking? A show off!",
}

// Bird wisdom quotes
var birdWisdom = []string{
	"Birds sing after every storm. There's always a song waiting after tough times!",
	"Like birds, we all have our own special song to sing.",
	"Every bird has wings, but each flies in their own special way.",
	"Birds teach us that it's good to sing, even on cloudy days.",
	"Just like birds build nests one twig at a time, we can do big things with small steps.",
	"Birds don't worry about tomorrow, they just enjoy today's adventure!",
	"Even the smallest bird can soar high in the sky.",
	"Birds remind us that everyone has their own beautiful voice.",
	"Like birds flying together, we're stronger when we help each other.",
	"Every bird was once in an egg - great things start small!",
}

// Fun bird facts
var birdFunFacts = []string{
	"Some birds can fly backwards! Hummingbirds are the helicopters of the bird world!",
	"Penguins propose with pebbles! They give a special stone to someone they love!",
	"Owls can't move their eyes, so they turn their heads almost all the way around!",
	"A group of flamingos is called a flamboyance! How fancy!",
	"Some birds can sleep while flying! They take power naps in the clouds!",
	"Woodpeckers can peck 20 times per second! That's faster than you can blink!",
	"Ravens can learn to talk better than some parrots!",
	"Birds are the only animals with feathers - that makes them extra special!",
	"Baby birds are called chicks, just like chickens! But not all birds are chickens!",
	"The ostrich egg is so big, you could make breakfast for 12 people with just one!",
	"Peacocks are actually boys - the girls are called peahens!",
	"Some birds can see colors we can't even imagine!",
	"Cardinals get their red color from the food they eat!",
	"Birds have hollow bones that make them light enough to fly!",
	"The Arctic Tern flies from the North Pole to the South Pole every year!",
}
