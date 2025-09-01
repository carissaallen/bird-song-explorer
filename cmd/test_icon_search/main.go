package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func searchYotoiconsTest(query string) {
	fmt.Printf("\n=== Testing search for: %s ===\n", query)
	
	searchURL := fmt.Sprintf("https://www.yotoicons.com/icons?tag=%s", url.QueryEscape(query))
	fmt.Printf("Search URL: %s\n", searchURL)
	
	resp, err := http.Get(searchURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP Status: %d\n", resp.StatusCode)
		return
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return
	}
	
	html := string(body)
	
	// Check what we get back
	if strings.Contains(html, "No icons found") || strings.Contains(html, "no results") {
		fmt.Printf("Result: No icons found\n")
	} else {
		// Parse HTML to find icon images
		iconRegex := regexp.MustCompile(`<img[^>]+src=["']/static/uploads/(\d+)\.png["'][^>]*>`)
		matches := iconRegex.FindAllStringSubmatch(html, -1)
		
		fmt.Printf("Found %d icon(s)\n", len(matches))
		
		// Check if "bird" appears in the page
		lowerHTML := strings.ToLower(html)
		if strings.Contains(lowerHTML, "bird") {
			fmt.Printf("Page contains 'bird' keyword\n")
		}
		
		// Check if the specific bird type appears
		lowerQuery := strings.ToLower(query)
		if strings.Contains(lowerHTML, lowerQuery) {
			fmt.Printf("Page contains '%s' keyword\n", lowerQuery)
		}
		
		// Look for title or description containing our search terms
		titleRegex := regexp.MustCompile(`<title>([^<]+)</title>`)
		titleMatch := titleRegex.FindStringSubmatch(html)
		if len(titleMatch) > 1 {
			fmt.Printf("Page title: %s\n", titleMatch[1])
		}
		
		// Check h1/h2 headers
		headerRegex := regexp.MustCompile(`<h[12][^>]*>([^<]+)</h[12]>`)
		headers := headerRegex.FindAllStringSubmatch(html, 5)
		for _, header := range headers {
			if len(header) > 1 {
				fmt.Printf("Header: %s\n", header[1])
			}
		}
	}
}

func main() {
	// Test with American Robin
	searchYotoiconsTest("American Robin")
	searchYotoiconsTest("Robin")
	searchYotoiconsTest("bird")
	
	// Test the proposed logic: search for "Robin" and check if results contain "bird"
	fmt.Printf("\n=== PROPOSED LOGIC TEST ===\n")
	fmt.Printf("Searching for 'Robin' and checking if result contains 'bird'...\n")
	
	searchURL := fmt.Sprintf("https://www.yotoicons.com/icons?tag=%s", url.QueryEscape("Robin"))
	resp, err := http.Get(searchURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	lowerHTML := strings.ToLower(html)
	
	hasRobin := strings.Contains(lowerHTML, "robin")
	hasBird := strings.Contains(lowerHTML, "bird")
	hasIcons := regexp.MustCompile(`<img[^>]+src=["']/static/uploads/(\d+)\.png["'][^>]*>`).MatchString(html)
	
	fmt.Printf("Has 'robin': %v\n", hasRobin)
	fmt.Printf("Has 'bird': %v\n", hasBird)
	fmt.Printf("Has icon images: %v\n", hasIcons)
	
	if (hasRobin || hasBird) && hasIcons {
		fmt.Printf("✓ Would successfully find a bird icon for American Robin!\n")
	} else {
		fmt.Printf("✗ Would NOT find a bird icon for American Robin\n")
	}
}