package main

import (
    "fmt"
    "strconv"
    "strings"
)

func cleanLocationName(s string) string {
    // Remove common year patterns (4-digit numbers)
    words := strings.Fields(s)
    var cleaned []string
    
    for _, word := range words {
        // Skip if it's a 4-digit year (1900-2099)
        if len(word) == 4 {
            if year, err := strconv.Atoi(word); err == nil && year >= 1900 && year <= 2099 {
                fmt.Printf("Skipping year: %s\n", word)
                continue
            }
        }
        
        // Skip if it contains date patterns like "2023-01-15" or "01/15/2023"
        if strings.Contains(word, "-") || strings.Contains(word, "/") {
            hasOnlyNumbersAndSeparators := true
            for _, r := range word {
                if !((r >= '0' && r <= '9') || r == '-' || r == '/') {
                    hasOnlyNumbersAndSeparators = false
                    break
                }
            }
            if hasOnlyNumbersAndSeparators {
                fmt.Printf("Skipping date pattern: %s\n", word)
                continue
            }
        }
        
        cleaned = append(cleaned, word)
    }
    
    return strings.Join(cleaned, " ")
}

func main() {
    testCases := []string{
        "Bend 2023",
        "Oregon",
        "Bend",
        "California 2024",
        "New York",
    }
    
    for _, test := range testCases {
        result := cleanLocationName(test)
        fmt.Printf("Input: %-20s -> Output: \"%s\"\n", fmt.Sprintf("\"%s\"", test), result)
    }
}