#!/bin/bash

# Test Yoto Authentication
# This script tests if your Yoto tokens are working

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}     Testing Yoto Authentication     ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    exit 1
fi

# Load environment variables
source .env

# Check if tokens exist
if [ -z "$YOTO_ACCESS_TOKEN" ]; then
    echo -e "${RED}Error: YOTO_ACCESS_TOKEN not found in .env${NC}"
    echo "Run ./scripts/get_yoto_tokens.sh to get tokens"
    exit 1
fi

echo -e "${YELLOW}Testing access token...${NC}"
echo ""

# Test the token by getting user info or cards
RESPONSE=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer ${YOTO_ACCESS_TOKEN}" \
    "https://api.yotoplay.com/user/cards")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Authentication successful!${NC}"
    echo ""
    echo "Response preview:"
    echo "$BODY" | head -c 200
    echo "..."
    echo ""
    
    # Try to get card count
    CARD_COUNT=$(echo "$BODY" | grep -o '"cardId"' | wc -l | tr -d ' ')
    if [ "$CARD_COUNT" -gt 0 ]; then
        echo -e "${BLUE}Found $CARD_COUNT card(s) in your account${NC}"
    fi
    
elif [ "$HTTP_CODE" = "401" ]; then
    echo -e "${RED}✗ Authentication failed - Token may be expired${NC}"
    echo ""
    echo "Try refreshing your token:"
    echo "./scripts/refresh_yoto_tokens.sh"
    exit 1
else
    echo -e "${RED}✗ Unexpected response: HTTP $HTTP_CODE${NC}"
    echo "Response body:"
    echo "$BODY"
    exit 1
fi

# If we have a card ID, test getting its details
if [ ! -z "$YOTO_CARD_ID" ]; then
    echo ""
    echo -e "${YELLOW}Testing card access (${YOTO_CARD_ID})...${NC}"
    
    CARD_RESPONSE=$(curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer ${YOTO_ACCESS_TOKEN}" \
        "https://api.yotoplay.com/content/${YOTO_CARD_ID}")
    
    CARD_HTTP_CODE=$(echo "$CARD_RESPONSE" | tail -n 1)
    
    if [ "$CARD_HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Can access card ${YOTO_CARD_ID}${NC}"
    else
        echo -e "${YELLOW}⚠ Cannot access card ${YOTO_CARD_ID} (HTTP $CARD_HTTP_CODE)${NC}"
        echo "This card ID might not exist in your account"
    fi
fi

echo ""
echo -e "${GREEN}Authentication test complete!${NC}"