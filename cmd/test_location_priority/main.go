package main

import (
	"fmt"
	"log"

	"github.com/callen/bird-song-explorer/internal/config"
	"github.com/callen/bird-song-explorer/internal/services"
)

func main() {
	_ = config.Load() // Load env vars
	
	locationService := services.NewLocationService()
	tzService := services.NewTimezoneLocationService()
	
	// Test scenarios
	scenarios := []struct {
		name         string
		ip           string
		timezone     string
		description  string
	}{
		{
			name:        "Valid IP, Valid Timezone",
			ip:          "8.8.8.8", // Google DNS - should resolve to Mountain View, CA
			timezone:    "America/New_York",
			description: "Should use IP location (Mountain View)",
		},
		{
			name:        "Invalid IP, Valid Timezone",
			ip:          "127.0.0.1",
			timezone:    "Europe/London",
			description: "Should use timezone location (London)",
		},
		{
			name:        "Invalid IP, Invalid Timezone",
			ip:          "127.0.0.1",
			timezone:    "Invalid/Timezone",
			description: "Should use default location (Bend, OR) with WARNING",
		},
		{
			name:        "Valid IP, No Timezone",
			ip:          "1.1.1.1", // Cloudflare DNS
			timezone:    "",
			description: "Should use IP location",
		},
	}
	
	fmt.Println("Testing Location Priority Logic:")
	fmt.Println("=================================")
	fmt.Println()
	
	for _, scenario := range scenarios {
		fmt.Printf("Scenario: %s\n", scenario.name)
		fmt.Printf("Description: %s\n", scenario.description)
		fmt.Printf("Input IP: %s, Timezone: %s\n", scenario.ip, scenario.timezone)
		
		// Simulate the logic from webhook handler
		location, err := locationService.GetLocationFromIP(scenario.ip)
		var deviceTimezone = scenario.timezone
		
		if err == nil && location.City != "Bend" {
			fmt.Printf("✓ Using IP-based location: %s, %s\n", location.City, location.Country)
		} else {
			// Try timezone fallback
			if deviceTimezone != "" {
				tzLocation := tzService.GetLocationFromTimezone(deviceTimezone)
				if tzLocation.City != "Bend" || deviceTimezone == "America/Denver" {
					location = tzLocation
					fmt.Printf("✓ Using timezone-based location: %s, %s\n", location.City, location.Country)
				}
			}
			
			// Check if still default
			if location.City == "Bend" && deviceTimezone != "America/Denver" {
				log.Printf("⚠️  WARNING: Using default location (Bend, OR) - IP: %s, Timezone: %s", 
					scenario.ip, deviceTimezone)
			}
		}
		
		fmt.Printf("Final Location: %s, %s\n", location.City, location.Country)
		fmt.Println()
	}
}