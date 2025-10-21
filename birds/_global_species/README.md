# Global Species Directory

This directory contains the master recordings for all bird species. Each bird has its own directory here, and regional directories use symbolic links to reference these master files.

## Directory Structure

Each bird species has its own directory with the following structure:

```
_global_species/
├── western_meadowlark/
│   ├── metadata.json           # Bird metadata including regions, habitats, etc.
│   ├── songs/
│   │   ├── song_01.mp3         # Primary song
│   │   ├── song_02.mp3         # Alternate songs
│   │   ├── call_01.mp3         # Call sounds
│   │   └── alarm_01.mp3        # Alarm calls
│   ├── narration/
│   │   ├── introduction.mp3    # Human-narrated bird introduction
│   │   ├── description.mp3     # Detailed bird description
│   │   └── fun_facts.mp3       # Fun facts about the bird
│   └── icons/
│       └── icon.png            # Bird icon (optional, usually empty)
```

## Benefits

1. **No Duplication**: Each recording exists only once
2. **Easy Updates**: Update in one place, affects all regions
3. **Consistent Metadata**: Single source of truth for bird information
4. **Efficient Storage**: Saves disk space by avoiding copies
5. **Version Control Friendly**: Changes tracked in one location