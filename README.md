# SubScalpelMKV

A cross-platform command-line tool for extracting subtitle tracks from MKV files using MKVToolNix.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
  - [Requirements](#requirements)
  - [Download Pre-built Binary](#download-pre-built-binary)
  - [Building from Source](#building-from-source)
- [Usage](#usage)
  - [Interactive Mode](#interactive-mode)
  - [Command Line Mode](#command-line-mode)
  - [Batch Processing](#batch-processing)
  - [Dry Run Mode](#dry-run-mode)
- [Track Selection](#track-selection)
  - [Selection Methods](#selection-methods)
  - [Exclusion Filters](#exclusion-filters)
  - [Language Codes](#language-codes)
- [Output Configuration](#output-configuration)
  - [Output Directory](#output-directory)
  - [Filename Templates](#filename-templates)
- [Configuration Files](#configuration-files)
  - [File Locations](#file-locations)
  - [Configuration Format](#configuration-format)
  - [Using Profiles](#using-profiles)
- [Command Reference](#command-reference)
- [Examples](#examples)
  - [Basic Examples](#basic-examples)
  - [Advanced Examples](#advanced-examples)
- [Supported Formats](#supported-formats)
- [License](#license)
- [Contributing](#contributing)
- [Acknowledgements](#acknowledgements)

## Features

- **Multiple input methods**: Interactive drag-and-drop, command-line interface, batch processing
- **Flexible track selection**: By language code, track number, or subtitle format
- **Track exclusion**: Filter out unwanted tracks from selection
- **Batch processing**: Process multiple files with glob patterns
- **Custom output**: Configurable output directories and filename templates
- **Configuration profiles**: Save common settings for repeated use
- **Dry run mode**: Preview operations before execution
- **Cross-platform**: Works on Windows, macOS, and Linux

## Quick Start

```sh
# Extract all subtitle tracks from a single file
./subscalpelmkv -x movie.mkv

# Extract English subtitles only
./subscalpelmkv -x movie.mkv -s eng

# Extract from multiple files
./subscalpelmkv -b "*.mkv" -s eng,spa

# Show track information
./subscalpelmkv -i video.mkv

# Preview what would be extracted (filename and output path)
./subscalpelmkv -x movie.mkv -s eng --dry-run
```

## Installation

### Requirements

- MKVToolNix (`mkvmerge` and `mkvextract`)

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/matthane/subscalpelmkv/releases):

- **Linux x86_64**: `subscalpelmkv_Linux_x86_64.tar.gz`
- **Linux ARM64**: `subscalpelmkv_Linux_arm64.tar.gz`
- **macOS Intel**: `subscalpelmkv_Darwin_x86_64.tar.gz`
- **macOS Apple Silicon**: `subscalpelmkv_Darwin_arm64.tar.gz`
- **Windows x86_64**: `subscalpelmkv_Windows_x86_64.zip`

After downloading:

1. Extract the archive (tar.gz or zip)
2. The binary is named `subscalpelmkv` (or `subscalpelmkv.exe` on Windows)
3. Make it executable (Linux/macOS): `chmod +x subscalpelmkv`
4. Move it to a directory in your PATH or run it from the current directory
5. Install [MKVToolNix](https://mkvtoolnix.download/) and ensure it's in your PATH

### Building from Source

If you prefer to build from source:

1. Install [Go](https://golang.org/dl/) 1.16 or later
   
2. Clone and build:
   ```sh
   git clone https://github.com/matthane/subscalpelmkv.git
   cd subscalpelmkv
   go build -o subscalpelmkv cmd/subscalpelmkv/main.go
   
   # For Windows
   go build -o subscalpelmkv.exe cmd/subscalpelmkv/main.go
   ```

## Usage

### Interactive Mode

Drag an MKV file, multiple files, or a directory onto the executable:

1. **Single file**: Interactive track selection
2. **Multiple files**: Batch processing with shared settings
3. **Directory**: Recursive processing of all MKV files

The interactive mode guides you through:
- Viewing available subtitle tracks
- Selecting tracks by language, number, or format
- Specifying output preferences
- Applying exclusion filters

### Command Line Mode

```sh
# Basic extraction
./subscalpelmkv -x video.mkv

# With track selection
./subscalpelmkv -x video.mkv -s eng,spa

# With output directory
./subscalpelmkv -x video.mkv -o ./subtitles

# Show track information
./subscalpelmkv -i video.mkv
```

### Batch Processing

Process multiple files using glob patterns:

```sh
# All MKV files in current directory
./subscalpelmkv -b "*.mkv" -s eng

# Files in subdirectory
./subscalpelmkv -b "Season1/*.mkv" -s eng,spa

# With custom output template
./subscalpelmkv -b "*.mkv" -s eng -f "{basename}-{language}.{extension}"
```

### Dry Run Mode

Preview extraction without creating files:

```sh
./subscalpelmkv -x video.mkv -s eng --dry-run
```

Displays:
- Number of tracks to extract
- Track details (number, language, format)
- Output filenames

## Track Selection

### Selection Methods

Select tracks using any combination of:

- **Language codes**: `eng`, `spa`, `fre` (2 or 3 letter ISO codes)
- **Track numbers**: `1`, `3`, `5`
- **Subtitle formats**: `srt`, `ass`, `sup`

```sh
# Language selection
./subscalpelmkv -x video.mkv -s eng,spa

# Track number selection
./subscalpelmkv -x video.mkv -s 1,3,5

# Format selection
./subscalpelmkv -x video.mkv -s srt,ass

# Mixed selection
./subscalpelmkv -x video.mkv -s eng,3,srt
```

### Exclusion Filters

Exclude specific tracks using `-e`:

```sh
# Exclude languages
./subscalpelmkv -x video.mkv -e chi,kor

# Exclude formats
./subscalpelmkv -x video.mkv -e sup,sub

# Select English but exclude image-based formats
./subscalpelmkv -x video.mkv -s eng -e sup,sub
```

### Language Codes

Supports both ISO 639-1 (2-letter) and ISO 639-2/B (3-letter) codes:
- English: `en` or `eng`
- Spanish: `es` or `spa` 
- French: `fr` or `fre`
- German: `de` or `ger`
- Japanese: `ja` or `jpn`
- Chinese: `zh` or `chi`
- And more...

## Output Configuration

### Output Directory

Control where subtitle files are saved:

```sh
# Same directory as input (default)
./subscalpelmkv -x video.mkv

# Auto-create {basename}-subtitles directory
./subscalpelmkv -x video.mkv -o

# Custom directory
./subscalpelmkv -x video.mkv -o ./subtitles
```

### Filename Templates

Customize output filenames with placeholders:

| Placeholder | Description |
|------------|-------------|
| `{basename}` | Original filename without extension |
| `{language}` | Track language code |
| `{trackno}` | Track number (zero-padded) |
| `{trackname}` | Track name (if available) |
| `{forced}` | "forced" for forced tracks |
| `{default}` | "default" for default tracks |
| `{extension}` | File extension |

```sh
# Simple: movie-eng.srt
-f "{basename}-{language}.{extension}"

# With track number: movie.001.eng.srt
-f "{basename}.{trackno}.{language}.{extension}"

# Organized by language: eng/movie.srt
-f "{language}/{basename}.{extension}"
```

## Configuration Files

### File Locations

Configuration files are searched in order:

1. `./subscalpelmkv.yaml` (current directory)
2. **Linux/macOS**: `~/.config/subscalpelmkv/config.yaml`
3. **Windows**: `%APPDATA%\subscalpelmkv\config.yaml`
4. `~/.subscalpelmkv.yaml` (home directory)

### Configuration Format

```yaml
# Default settings
default_languages: [eng, spa]
default_exclusions: [chi, kor]
output_template: "{basename}.{language}.{trackno}.{extension}"
output_dir: "./subtitles"

# Named profiles
profiles:
  anime:
    languages: [jpn, eng]
    exclusions: [chi, kor]
    output_template: "{basename}/{language}.{extension}"
    
  movies:
    languages: [eng]
    exclusions: [sup, sub]
    output_template: "{basename}-{language}.{extension}"
```

### Using Profiles

```sh
# Use default configuration
./subscalpelmkv -x video.mkv --config

# Use named profile
./subscalpelmkv -x video.mkv --profile anime

# Override profile settings
./subscalpelmkv -x video.mkv --profile anime -s eng
```

## Command Reference

| Option | Short | Description |
|--------|-------|-------------|
| `--extract` | `-x` | Extract subtitles from MKV file |
| `--batch` | `-b` | Process multiple files with glob pattern |
| `--select` | `-s` | Select tracks (languages/numbers/formats) |
| `--exclude` | `-e` | Exclude tracks (languages/numbers/formats) |
| `--info` | `-i` | Display track information |
| `--output-dir` | `-o` | Output directory (or auto-create with no args) |
| `--format` | `-f` | Filename template |
| `--dry-run` | `-d` | Preview without extraction |
| `--config` | `-c` | Use default configuration |
| `--profile` | `-p` | Use named profile |
| `--help` | `-h` | Show help |
| `--version` | `-v` | Show version information |

## Examples

### Basic Examples

```sh
# Extract all subtitles
./subscalpelmkv -x movie.mkv

# Extract specific language
./subscalpelmkv -x movie.mkv -s eng

# Extract multiple languages
./subscalpelmkv -x movie.mkv -s eng,spa,fre

# Extract to custom directory
./subscalpelmkv -x movie.mkv -o ./subs
```

### Advanced Examples

```sh
# Batch process with auto-created directories
./subscalpelmkv -b "*.mkv" -s eng -o

# Extract text subtitles only (exclude image-based)
./subscalpelmkv -x movie.mkv -s eng,spa -e sup,sub

# Complex selection with custom naming
./subscalpelmkv -x movie.mkv -s eng,spa,3,5 -e chi -f "{language}/{basename}.{extension}"

# Use configuration profile for anime
./subscalpelmkv -b "Anime/*.mkv" --profile anime

# Dry run to preview batch operation
./subscalpelmkv -b "Season1/*.mkv" -s eng,spa --dry-run
```

## Supported Formats

### Text-based Subtitles
- SubRip (`.srt`)
- Advanced SubStation Alpha (`.ass`)
- SubStation Alpha (`.ssa`)
- WebVTT (`.vtt`)
- Universal Subtitle Format (`.usf`)
- Plain text (`.txt`)

### Image-based Subtitles
- PGS/SUP (`.sup`)
- VOBSUB (`.idx` + `.sub`)
- DVB subtitles (`.sub`)
- Bitmap (`.bmp`)

### Other Formats
- Kate streams (`.kate`)
- HDMV TextST (`.txt`)

## License

This project is licensed under the MIT License. See the `LICENSE.md` file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## Acknowledgements

- [GMMK MKV Subtitles Extract](https://github.com/rhaseven7h/gmmmkvsubsextract)
- [MKVToolNix](https://mkvtoolnix.download/)
- [gocmd](https://github.com/devfacet/gocmd)