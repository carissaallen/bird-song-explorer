# Bird Audio Files Organization

This directory contains bird audio files organized by geographic regions with a centralized storage system to prevent duplication.

## Storage Architecture

### Centralized Storage with Regional Links
To avoid duplicating recordings for birds that live across multiple regions, we use a centralized storage approach:

1. **Master Storage**: All bird recordings are stored in `_global_species/`
2. **Regional Access**: Regional directories contain symbolic links to the master files
3. **No Duplication**: Each recording exists only once on disk

### Directory Structure

#### Global Species Directory
```
_global_species/
└── [bird_name]/
    ├── metadata.json     # Complete bird information and regions
    ├── songs/            # Bird song recordings (songs, calls, alarms)
    ├── narration/        # Bird-specific narrated descriptions
    └── icons/            # Bird icons (also in assets/icons) 
```

#### Global Narration Directory
```
global_narration/
├── morning/
│   ├── intro/        # Morning introduction narrations
│   └── outro/        # Morning outro narrations
├── afternoon/
│   ├── intro/        # Afternoon introduction narrations
│   └── outro/        # Afternoon outro narrations
└── evening/
    ├── intro/        # Evening introduction narrations
    └── outro/        # Evening outro narrations
```

#### Regional Directories

### North America
- `north_america/us/` - United States
- `north_america/canada/` - Canada
- `north_america/mexico/` - Mexico
- `north_america/caribbean/` - Caribbean islands
- `north_america/central_america/` - Central American countries

### Europe
- `europe/uk/` - United Kingdom
- `europe/western_europe/` - France, Germany, Netherlands, Belgium, etc.
- `europe/northern_europe/` - Scandinavia, Baltic states
- `europe/eastern_europe/` - Poland, Ukraine, Russia (European part), etc.
- `europe/southern_europe/` - Spain, Portugal, Italy, Greece, etc.
- `europe/central_europe/` - Austria, Switzerland, Czech Republic, etc.

### Africa
- `africa/north_africa/` - Egypt, Libya, Tunisia, Algeria, Morocco
- `africa/west_africa/` - Nigeria, Ghana, Senegal, etc.
- `africa/east_africa/` - Kenya, Ethiopia, Tanzania, etc.
- `africa/central_africa/` - DRC, Cameroon, Central African Republic, etc.
- `africa/southern_africa/` - South Africa, Zimbabwe, Botswana, etc.
- `africa/sub_saharan/` - Sub-Saharan region (overlap with other African regions)
- `africa/madagascar/` - Madagascar and nearby islands

### Asia
- `asia/east_asia/` - China, Japan, Korea, Mongolia
- `asia/southeast_asia/` - Thailand, Vietnam, Indonesia, Philippines, etc.
- `asia/south_asia/` - India, Pakistan, Bangladesh, Sri Lanka, Nepal
- `asia/central_asia/` - Kazakhstan, Uzbekistan, Kyrgyzstan, etc.
- `asia/middle_east/` - Saudi Arabia, Iran, Iraq, Israel, Jordan, etc.

### Oceania
- `oceania/australia/` - Australia
- `oceania/new_zealand/` - New Zealand
- `oceania/pacific_islands/` - Fiji, Samoa, Hawaii, etc.
- `oceania/papua_new_guinea/` - Papua New Guinea and nearby islands

### South America
- `south_america/amazon_basin/` - Amazon rainforest region
- `south_america/andes_region/` - Andean mountain regions
- `south_america/southern_cone/` - Argentina, Chile, Uruguay
- `south_america/atlantic_forest/` - Brazil's Atlantic coastal forests
- `south_america/guianas/` - Guyana, Suriname, French Guiana
- `south_america/pantanal/` - Pantanal wetland region

### Polar Regions
- `polar_regions/arctic/` - Arctic region birds
- `polar_regions/antarctica/` - Antarctic region birds

## Usage

### Adding a New Bird Species

Use the provided script to add a new bird:

```bash
./scripts/add_bird.sh "Common Name" "Scientific name" "region1,region2"

# Example:
./scripts/add_bird.sh "American Robin" "Turdus migratorius" "north_america/us,north_america/canada"
```

This will:
1. Create the bird directory in `_global_species/`
2. Generate a metadata.json template
3. Create symbolic links in all specified regional directories

### Manual Process

1. Create bird directory: `birds/_global_species/[bird_name]/`
2. Add subdirectories: `songs/`, `narration/`, `icons/`
3. Create `metadata.json` with bird information
4. Add symbolic links in regional directories:
   ```bash
   cd birds/[region]/[subregion]
   ln -s ../../_global_species/[bird_name] [bird_name]
   ```

## Example: Western Meadowlark

The Western Meadowlark lives across multiple North American regions:

```
_global_species/western_meadowlark/     # Master files
├── metadata.json
├── songs/
├── narration/
└── icons/

north_america/us/western_meadowlark     # → symlink to ../../_global_species/western_meadowlark
north_america/canada/western_meadowlark # → symlink to ../../_global_species/western_meadowlark
north_america/mexico/western_meadowlark # → symlink to ../../_global_species/western_meadowlark
```
