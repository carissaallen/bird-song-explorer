package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/callen/bird-song-explorer/pkg/yoto"
)

func main() {
	// Initialize Yoto client
	client := yoto.NewClient(
		os.Getenv("YOTO_CLIENT_ID"),
		os.Getenv("YOTO_ACCESS_TOKEN"),
		os.Getenv("YOTO_REFRESH_TOKEN"),
	)
	
	// Create content manager
	cm := client.NewContentManager()
	
	// Update card with streaming tracks
	cardID := "ipHAS"
	birdName := "American Robin"
	baseURL := "https://bird-song-explorer-362662614716.us-central1.run.app"
	
	fmt.Printf("Updating card %s with streaming tracks for %s...\n", cardID, birdName)
	
	err := cm.UpdateCardWithStreamingTracks(cardID, birdName, baseURL)
	if err != nil {
		log.Fatalf("Failed to update card: %v", err)
	}
	
	fmt.Println("Card updated successfully!")
}