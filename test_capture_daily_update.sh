#!/bin/bash

# Test script to capture raw JSON request for daily-update content endpoint

# Set up the base URL (adjust if needed)
BASE_URL="http://localhost:8080"

# Create a timestamp for the output file
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="daily_update_request_${TIMESTAMP}.txt"

echo "Capturing daily-update request to: $OUTPUT_FILE"
echo "======================================" > "$OUTPUT_FILE"
echo "Daily Update Content Request" >> "$OUTPUT_FILE"
echo "Timestamp: $(date)" >> "$OUTPUT_FILE"
echo "======================================" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Make the request with curl and capture the raw request
echo "Request Details:" >> "$OUTPUT_FILE"
echo "---------------" >> "$OUTPUT_FILE"
echo "URL: ${BASE_URL}/content?type=daily-update" >> "$OUTPUT_FILE"
echo "Method: GET" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Use curl with verbose output to capture the full request
echo "Full Request (with headers):" >> "$OUTPUT_FILE"
echo "----------------------------" >> "$OUTPUT_FILE"
curl -v -X GET "${BASE_URL}/content?type=daily-update" 2>&1 | tee -a "$OUTPUT_FILE"

echo "" >> "$OUTPUT_FILE"
echo "======================================" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Also capture just the response body in JSON format
echo "Response Body (JSON):" >> "$OUTPUT_FILE"
echo "--------------------" >> "$OUTPUT_FILE"
curl -s -X GET "${BASE_URL}/content?type=daily-update" | jq '.' >> "$OUTPUT_FILE" 2>/dev/null || \
curl -s -X GET "${BASE_URL}/content?type=daily-update" >> "$OUTPUT_FILE"

echo ""
echo "Request captured to: $OUTPUT_FILE"
echo "Use 'cat $OUTPUT_FILE' to view the contents"