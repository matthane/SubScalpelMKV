# SubScalpelMKV

`subscalpelmkv` is a cross-platform command-line tool written in Go for extracting subtitle tracks from MKV files. It uses MKVToolNix for the extraction operations and supports both interactive and command-line modes.

## Features

- Extract subtitle tracks from MKV files using MKVToolNix
- Batch processing with glob patterns for multiple files
- Subtitle format support:
  - Text-based: SRT, ASS, SSA, WebVTT, USF, TXT
  - Image-based: SUP (PGS), VOBSUB (IDX/SUB), DVB subtitles, BMP
  - Other: KATE, HDMV/TEXTST
- Track selection by language codes, track numbers, or format types
- Customizable output directories and filename templates
- Dry run mode to preview extraction without creating files
- Configuration file support with named profiles
- Interactive drag-and-drop mode
- Command-line interface

## Requirements

- Go 1.16 or later
- `mkvmerge` and `mkvextract` tools from the MKVToolNix package
- `gocmd` library

## Installation

1. Install Go from [golang.org](https://golang.org/dl/)

2. Install MKVToolNix from [mkvtoolnix.download](https://mkvtoolnix.download/)
    - Add to PATH system environment variables

3. Clone the repository and navigate to the project directory:

    ```sh
    git clone https://github.com/matthane/subscalpelmkv.git
    cd subscalpelmkv
    ```
4. Build the project:

    ```sh
    go build -o subscalpelmkv cmd/subscalpelmkv/main.go

    # For Windows
    go build -o subscalpelmkv.exe cmd/subscalpelmkv/main.go
    ```

## Usage

### Drag-and-Drop Mode

#### Single File Drag-and-Drop
Drag an MKV file onto the executable for interactive mode:

1. The tool analyzes the file and displays available subtitle tracks
2. Choose to extract all tracks or make a custom selection
3. For custom selection, enter:
    - Language codes: `eng,spa,fre`
    - Track numbers: `3,5,7`
    - Subtitle formats: `srt,ass,sup`
    - Mixed selection: `eng,3,srt,sup`

#### Multi-File Drag-and-Drop
Drag multiple MKV files onto the executable for batch processing:

1. Detects when multiple MKV files are provided
2. Shows all files that will be processed
3. Apply the same track selection to all files
4. Shows progress for each file
5. Displays processing summary

Multi-file processing continues with remaining files if one fails.

#### Directory Drag-and-Drop
Drag a folder onto the executable to process all MKV files within it:

1. Recursively scans the directory for all MKV files
2. Shows all found MKV files that will be processed
3. Apply the same track selection to all files
4. Shows progress for each file
5. Displays processing summary

You can also drag multiple folders and files together - the tool will collect all MKV files from both individual files and directories.

### Command Line Mode

#### Basic Usage
```sh
# Extract all subtitle tracks
./subscalpelmkv -x "path/to/video.mkv"
```

#### Batch Processing
Process multiple MKV files at once using glob patterns:

```sh
# Process all MKV files in current directory
./subscalpelmkv -b "*.mkv" -s eng

# Process all episodes in a season folder
./subscalpelmkv -b "Season 1/*.mkv" -s eng,spa

# Process files in a specific path
./subscalpelmkv -b "/path/to/movies/*.mkv" -s eng

# Batch process with auto-created {basename}-subtitles directories
./subscalpelmkv -b "*.mkv" -s eng,spa -o

# Batch process with custom output directory
./subscalpelmkv -b "*.mkv" -s eng,spa -o ./subtitles

# Batch process with custom filename template
./subscalpelmkv -b "Season 1/*.mkv" -s eng -f "{basename}-{language}.{extension}"
```

Batch processing filters to `.mkv` files only and continues processing remaining files if one fails.

#### Subtitle Track Selection With Additive Filtering
```sh
# Single language
./subscalpelmkv -x "path/to/video.mkv" -s eng

# Multiple languages
./subscalpelmkv -x "path/to/video.mkv" -s eng,spa,fre

# Specific track numbers
./subscalpelmkv -x "path/to/video.mkv" -s 1,3,5

# Subtitle format filtering
./subscalpelmkv -x "path/to/video.mkv" -s srt,ass

# Extract only image-based subtitles
./subscalpelmkv -x "path/to/video.mkv" -s sup

# Mixed selection: languages, track numbers, and formats
./subscalpelmkv -x "path/to/video.mkv" -s eng,3,srt,sup
```

#### Dry Run Mode
Preview what subtitles would be extracted without performing the actual extraction:

```sh
# Preview extraction for specific language
./subscalpelmkv -x "path/to/video.mkv" -s eng --dry-run

# Preview batch processing
./subscalpelmkv -b "*.mkv" -s eng,spa --dry-run

# Preview with custom output template
./subscalpelmkv -x "path/to/video.mkv" -s eng -f "{basename}-{language}.{extension}" --dry-run
```

Dry run mode displays:
- Number of tracks that would be extracted
- Track details (number, language, name, format, attributes)
- Output filename that would be created

#### Output Control
```sh
# Custom output directory (automatically created if it doesn't exist)
./subscalpelmkv -x "path/to/video.mkv" -o ./subtitles

# Auto-create {basename}-subtitles directory in input file's directory
./subscalpelmkv -x "path/to/video.mkv" -o

# Custom filename template
./subscalpelmkv -x "path/to/video.mkv" -f "{basename}-{language}.{extension}"

# Combined: custom directory, template, and track selection
./subscalpelmkv -x "path/to/video.mkv" -s eng,spa -o ./subs -f "{language}-{trackno}.{extension}"
```

#### Info Flag Usage
```sh
# Show information about available subtitle tracks
./subscalpelmkv -i "path/to/video.mkv"
```

#### Command Line Options

##### Selection Options
| Option | Short | Description |
|--------|-------|-------------|
| `--extract` | `-x` | Path to MKV file (required) |
| `--batch` | `-b` | Extract subtitles from multiple MKV files using glob pattern (e.g., '*.mkv', 'Season 1/*.mkv') |
| `--select` | `-s` | Language codes, track numbers, subtitle formats, or any combination (comma-separated) |
| `--info` | `-i` | Show information about available subtitle tracks |
| `--help` | `-h` | Show help message |

##### Output Options
| Option | Short | Description |
|--------|-------|-------------|
| `--output-dir` | `-o` | Custom output directory, or when used without arguments, auto-create `{basename}-subtitles` directory in input file's location (automatically created if it doesn't exist, default: same as input file) |
| `--format` | `-f` | Custom filename template with placeholders |
| `--dry-run` | `-d` | Preview what would be extracted without performing extraction |
| `--config` | `-c` | Use default configuration profile |
| `--profile` | `-p` | Use named configuration profile |

#### Language Code Support

Supports ISO 639-1 (2-letter) and ISO 639-2/B (3-letter) language codes:
- 2-letter: `en`, `es`, `fr`, `de`, `ja`, `zh`, and more
- 3-letter: `eng`, `spa`, `fre`, `ger`, `jpn`, `chi`, and more

### Configuration Files

SubScalpelMKV supports configuration files to set default options and create named profiles for different use cases. This eliminates the need to specify the same options repeatedly.

#### Configuration File Locations

Configuration files are searched in this order (first found is used):

1. **Current directory**: `./subscalpelmkv.yaml`
2. **OS-specific config directory**:
   - **Linux/macOS**: `~/.config/subscalpelmkv/config.yaml`
   - **Windows**: `%APPDATA%\subscalpelmkv\config.yaml`
3. **Home directory**: `~/.subscalpelmkv.yaml`

#### Configuration Format

Configuration files use YAML format with the following structure:

```yaml
# Default settings applied when using --config
default_languages: [eng, spa]
output_template: "{basename}.{language}.{trackno}.{extension}"
output_dir: "./subtitles"

# Named profiles for different use cases
profiles:
  anime:
    languages: [jpn, eng]
    output_template: "{basename}/{language}.{extension}"
    output_dir: "./anime_subs"
    
  movies:
    languages: [eng]
    output_template: "{basename}-{language}.{extension}"
    output_dir: "./movie_subs"
    
  multilang:
    languages: [eng, spa, fre, ger]
    output_template: "{basename}/{language}/{trackname}.{extension}"
    output_dir: "./multi_subs"
```

#### Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `default_languages` | Array | Default language codes to extract |
| `output_template` | String | Default filename template |
| `output_dir` | String | Default output directory |
| `profiles` | Object | Named configuration profiles |

#### Using Configuration

```sh
# Use default configuration settings
./subscalpelmkv -x video.mkv --config

# Use a named profile
./subscalpelmkv -x video.mkv --profile anime

# CLI flags override config settings
./subscalpelmkv -x video.mkv --profile anime -s eng

# Batch processing with configuration
./subscalpelmkv -b "*.mkv" --profile movies

# Dry run with configuration
./subscalpelmkv -x video.mkv --profile anime --dry-run
```

#### Configuration Priority

Settings are applied in this order (later values override earlier ones):

1. Built-in defaults
2. Configuration file defaults
3. Named profile settings (if `--profile` is used)
4. Command-line flags

This means CLI flags always take precedence over configuration settings.

### Output File Naming

#### Default Naming Pattern
By default, output files are named using the pattern:
```
<original_filename>.<language>.<track_number>[.track_name][.forced][.default].<extension>
```

Examples:
- `movie.eng.001.srt` - English SRT subtitle, track 1
- `movie.spa.002.ass` - Spanish ASS subtitle, track 2
- `movie.eng.003.forced.sup` - English forced SUP subtitle, track 3
- `movie.fre.004.default.vtt` - French default WebVTT subtitle, track 4
- `movie.ger.005.sub` - German VOBSUB subtitle, track 5 (creates .idx and .sub files)

#### Custom Filename Templates
You can customize the output filename format using the `-f` flag with placeholders:

**Available Placeholders:**
- `{basename}` - Original filename without extension
- `{language}` - Track language code (e.g., "eng", "spa")
- `{trackno}` - Track number (3-digit padded, e.g., "001", "042")
- `{trackname}` - Track name (if available)
- `{forced}` - "forced" if track is forced, empty otherwise
- `{default}` - "default" if track is default, empty otherwise
- `{extension}` - Subtitle file extension (srt, ass, ssa, vtt, usf, sup, sub, bmp, kate, txt)

**Template Examples:**
```sh
# Simple format: movie-eng.srt
-f "{basename}-{language}.{extension}"

# Include track numbers: eng-001.srt
-f "{language}-{trackno}.{extension}"

# Detailed format: movie.english.track001.srt
-f "{basename}.{language}.track{trackno}.{extension}"

# Include forced/default flags: movie.forced.srt
-f "{basename}.{forced}.{extension}"
```

#### Output Directory Control
- **Default**: Files are saved in the same directory as the input MKV file
- **Auto-create Directory**: Use `-o` without arguments to create a `{basename}-subtitles` directory in the same location as the input file
- **Custom Directory**: Use `-o <directory>` to specify a different output directory (automatically created if it doesn't exist)

**Directory Examples:**
```sh
# Default behavior - save in same directory as input file
./subscalpelmkv -x path/to/video.mkv
# Result: subtitles saved to path/to/

# Auto-create {basename}-subtitles directory in input file's location
./subscalpelmkv -x path/to/video.mkv -o
# Result: subtitles saved to path/to/video-subtitles/

# Save to specific directory
./subscalpelmkv -x path/to/video.mkv -o ./extracted-subtitles
# Result: subtitles saved to ./extracted-subtitles/

# Organize by movie name
./subscalpelmkv -x path/to/video.mkv -o "./subtitles/Movie Name"
# Result: subtitles saved to ./subtitles/Movie Name/

# Batch processing with auto-created {basename}-subtitles directories per file
./subscalpelmkv -b "Season1/*.mkv" -o -s eng
# Result: each file gets its own directory (e.g., Season1/episode01-subtitles/)
```
## License

This project is licensed under the MIT License. See the `LICENSE.md` file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## Acknowledgements

- [GMMK MKV Subtitles Extract](https://github.com/rhaseven7h/gmmmkvsubsextract)
- [MKVToolNix](https://mkvtoolnix.download/)
- [gocmd](https://github.com/devfacet/gocmd)
