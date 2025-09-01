package main

import (
	"fmt"
	"log"
	"os"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/services"
	"github.com/callen/bird-song-explorer/pkg/yoto"
)

func main() {
	// Load config
	cfg := config.Load()

	// Test timezone to location mapping
	tzService := services.NewTimezoneLocationService()
	
	testTimezones := []string{
		"America/New_York",
		"America/Los_Angeles",
		"Europe/London",
		"Australia/Sydney",
		"Asia/Tokyo",
		"America/Chicago",
		"Europe/Paris",
		"GMT-5",
		"UTC+10",
		"PST",
		"EST",
		"",
	}

	fmt.Println("Testing Timezone to Location Mapping:")
	fmt.Println("=====================================")
	for _, tz := range testTimezones {
		location := tzService.GetLocationFromTimezone(tz)
		fmt.Printf("Timezone: %-25s -> %s, %s, %s (%.4f, %.4f)\n", 
			tz, location.City, location.Region, location.Country, 
			location.Latitude, location.Longitude)
	}

	// Test device config API if device ID is provided
	if len(os.Args) > 1 {
		deviceID := os.Args[1]
		fmt.Printf("\nTesting Device Config API for device: %s\n", deviceID)
		fmt.Println("=========================================")
		
		yotoClient := yoto.NewClient(
			cfg.YotoClientID,
			cfg.YotoClientSecret,
			cfg.YotoAPIBaseURL,
		)

		config, err := yotoClient.GetDeviceConfig(deviceID)
		if err != nil {
			log.Printf("Error getting device config: %v", err)
		} else {
			fmt.Printf("Device ID: %s\n", config.Device.DeviceID)
			fmt.Printf("Device Online: %v\n", config.Device.Online)
			fmt.Printf("Device Type: %s\n", config.Device.DeviceType)
			fmt.Printf("Device Family: %s\n", config.Device.DeviceFamily)
			fmt.Printf("Geo Timezone: %s\n", config.Device.Config.GeoTimezone)
			
			// Map timezone to location
			location := tzService.GetLocationFromTimezone(config.Device.Config.GeoTimezone)
			fmt.Printf("Mapped Location: %s, %s, %s (%.4f, %.4f)\n", 
				location.City, location.Region, location.Country,
				location.Latitude, location.Longitude)
			
			// Test bird selection for this location
			birdSelector := services.NewBirdSelector(cfg.EBirdAPIKey, cfg.XenoCantoAPIKey)
			bird, err := birdSelector.SelectBirdOfDay(location)
			if err != nil {
				log.Printf("Error selecting bird: %v", err)
			} else {
				fmt.Printf("Selected Bird: %s\n", bird.CommonName)
				fmt.Printf("Bird Audio URL: %s\n", bird.AudioURL)
			}
		}
	}
}