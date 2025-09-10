package main

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"github.com/callen/bird-song-explorer/pkg/yoto"
)

func main() {
	// Force an update to the card with streaming configuration
	fmt.Println("Forcing update of Yoto card ipHAS with streaming configuration...")
	
	// Initialize Yoto client
	client := yoto.NewClient(
		"qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU",
		"", // No client secret needed
		"https://api.yotoplay.com",
	)
	
	// Set the tokens
	accessToken := os.Getenv("YOTO_ACCESS_TOKEN")
	refreshToken := os.Getenv("YOTO_REFRESH_TOKEN")
	
	fmt.Printf("Access token present: %v (length: %d)\n", accessToken != "", len(accessToken))
	fmt.Printf("Refresh token present: %v (length: %d)\n", refreshToken != "", len(refreshToken))
	
	client.SetTokens(accessToken, refreshToken, 86400)
	
	// Create content manager
	cm := client.NewContentManager()
	
	// Update card with streaming tracks for a simple bird
	cardID := "ipHAS"
	birdName := "Northern Cardinal" // Changed to force cache refresh
	baseURL := "https://bird-song-explorer-362662614716.us-central1.run.app"
	
	fmt.Printf("Updating card %s with streaming tracks for %s...\n", cardID, birdName)
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	
	err := cm.UpdateCardWithStreamingTracks(cardID, birdName, baseURL)
	if err != nil {
		log.Fatalf("Failed to update card: %v", err)
	}
	
	fmt.Println("✅ Card updated successfully with streaming configuration!")
	fmt.Println("The card should now:")
	fmt.Println("  - Have title: 'Bird Song Explorer' (not 'Bird Song Explorer - American Robin')")
	fmt.Println("  - Use streaming URLs for dynamic content")
	fmt.Println("  - Show American Robin chapters for now")
}