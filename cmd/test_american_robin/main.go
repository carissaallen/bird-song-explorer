package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func testSearch(query string) bool {
	searchURL := fmt.Sprintf("https://www.yotoicons.com/icons?tag=%s", url.QueryEscape(query))
	
	resp, err := http.Get(searchURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	lowerHTML := strings.ToLower(html)
	
	// Check for icons
	iconRegex := regexp.MustCompile(`<img[^>]+src=["']/static/uploads/(\d+)\.png["'][^>]*>`)
	matches := iconRegex.FindAllStringSubmatch(html, -1)
	
	// Check for search term and bird keyword
	hasSearchTerm := strings.Contains(lowerHTML, strings.ToLower(query))
	hasBirdKeyword := strings.Contains(lowerHTML, "bird")
	hasIcons := len(matches) > 0
	
	fmt.Printf("  Query: '%s'\n", query)
	fmt.Printf("    Has icons: %v (%d found)\n", hasIcons, len(matches))
	fmt.Printf("    Has search term '%s': %v\n", query, hasSearchTerm)
	fmt.Printf("    Has 'bird' keyword: %v\n", hasBirdKeyword)
	
	// Success if we have icons AND (search term + bird keyword)
	success := hasIcons && (hasSearchTerm && hasBirdKeyword)
	fmt.Printf("    ✓ Would accept this result: %v\n", success)
	
	return success
}

func main() {
	fmt.Println("=== TESTING AMERICAN ROBIN SEARCH ===\n")
	
	fmt.Println("Step 1: Try full name 'American Robin'")
	fullNameSuccess := testSearch("American Robin")
	
	fmt.Println("\nStep 2: Try variation 'Robin'")
	variationSuccess := testSearch("Robin")
	
	fmt.Println("\n=== FINAL RESULT ===")
	if fullNameSuccess {
		fmt.Println("✅ SUCCESS: 'American Robin' search would find an icon directly!")
	} else if variationSuccess {
		fmt.Println("✅ SUCCESS: 'American Robin' would find an icon via 'Robin' variation!")
	} else {
		fmt.Println("❌ FAILURE: Would not find an icon for 'American Robin'")
	}
	
	fmt.Println("\nThe updated logic ensures that:")
	fmt.Println("- When searching for 'Robin', we verify both 'robin' AND 'bird' appear on the page")
	fmt.Println("- This confirms we're getting a bird icon, not something unrelated")
	fmt.Println("- The search is now less strict but still validates we're getting bird-related content")
}