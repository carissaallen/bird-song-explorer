#!/bin/bash

# Script to add a new bird species to the system
# Usage: ./add_bird.sh "common_name" "scientific_name" "region1,region2,region3"

set -e

if [ $# -lt 3 ]; then
    echo "Usage: $0 \"common_name\" \"scientific_name\" \"region1,region2,region3\""
    echo "Example: $0 \"American Robin\" \"Turdus migratorius\" \"north_america/us,north_america/canada\""
    exit 1
fi

COMMON_NAME="$1"
SCIENTIFIC_NAME="$2"
REGIONS="$3"

# Convert common name to directory name (lowercase, replace spaces with underscores)
DIR_NAME=$(echo "$COMMON_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')

BASE_DIR="birds/_global_species/$DIR_NAME"

# Check if bird already exists
if [ -d "$BASE_DIR" ]; then
    echo "Error: Bird '$COMMON_NAME' already exists at $BASE_DIR"
    exit 1
fi

echo "Creating bird directory structure for: $COMMON_NAME"

# Create directory structure
mkdir -p "$BASE_DIR/songs"
mkdir -p "$BASE_DIR/narration"
mkdir -p "$BASE_DIR/icons"

# Create metadata.json
cat > "$BASE_DIR/metadata.json" << EOF
{
  "common_name": "$COMMON_NAME",
  "scientific_name": "$SCIENTIFIC_NAME",
  "family": "",
  "regions": [
$(IFS=','; regions_array=($REGIONS); \
  for i in "${!regions_array[@]}"; do \
    if [ $i -eq $((${#regions_array[@]} - 1)) ]; then \
      echo "    \"${regions_array[$i]}\""; \
    else \
      echo "    \"${regions_array[$i]}\","; \
    fi; \
  done)
  ],
  "primary_habitat": "",
  "habitats": [],
  "conservation_status": "",
  "size": {
    "length_cm": "",
    "wingspan_cm": "",
    "weight_g": ""
  },
  "diet": [],
  "breeding_season": "",
  "migration_pattern": "",
  "distinctive_features": [],
  "fun_facts": [],
  "song_locations": {},
  "recorded_date": "$(date +%Y)",
  "narrator": "pending_human_recording"
}
EOF

echo "Created bird directory at: $BASE_DIR"
echo "Created metadata.json template"

# Create symbolic links in regional directories
IFS=',' read -ra REGION_ARRAY <<< "$REGIONS"
for region in "${REGION_ARRAY[@]}"; do
    region_path="birds/$region"
    if [ ! -d "$region_path" ]; then
        echo "Warning: Region directory '$region_path' does not exist. Skipping symlink."
        continue
    fi

    # Calculate relative path from region to global species
    depth=$(echo "$region" | tr -cd '/' | wc -c)
    relative_path="../.."
    for ((i=0; i<depth; i++)); do
        relative_path="../$relative_path"
    done
    relative_path="$relative_path/_global_species/$DIR_NAME"

    # Create symlink
    ln -s "$relative_path" "$region_path/$DIR_NAME"
    echo "Created symlink in: $region_path/$DIR_NAME"
done

echo ""
echo "Bird '$COMMON_NAME' has been successfully added!"
echo "Next steps:"
echo "1. Edit $BASE_DIR/metadata.json to add complete information"
echo "2. Add song recordings to $BASE_DIR/songs/"
echo "3. Add narration files to $BASE_DIR/narration/"
echo "4. Add icon images to $BASE_DIR/icons/ (if needed)"