# SubScalpelMKV

`subscalpelmkv` is a cross-platform command-line tool written in Go for extracting subtitles from MKV files quickly and precisely. It is a fork/overhaul of [GMM MKV Subtitles Extract](https://github.com/rhaseven7h/gmmmkvsubsextract) with many new features for increased speed and capability.

## Features

- Extract subtitle tracks from MKV files using MKVToolNix
- **Batch processing**: Process multiple MKV files at once using glob patterns
- Support for multiple subtitle formats:
  - **Text-based**: SRT, ASS, SSA, WebVTT, USF, TXT
  - **Image-based**: SUP (PGS), VOBSUB (IDX/SUB), DVB subtitles, BMP
  - **Other**: KATE, HDMV/TEXTST
- Flexible track selection using language codes, track numbers, or any combination of both
- Output control: custom directories and filename templates
- Interactive mode via drag-and-drop
- Command-line interface for scripting and automation

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

1. **Automatic Detection**: The tool automatically detects when multiple MKV files are provided
2. **File Listing**: Shows all files that will be processed
3. **Unified Selection**: Choose to extract all tracks or make a custom selection that applies to all files
4. **Progress Tracking**: Shows progress for each file (e.g., "Processing file 2/5")
5. **Batch Summary**: Displays total files processed and success/failure counts

**Multi-File Features:**
- **Smart Detection**: Automatically distinguishes between multiple files and single files with spaces in names
- **Same Interface**: Uses the same track selection interface as single-file mode
- **Error Resilience**: Continues processing other files if one fails
- **Consistent Output**: Each file's subtitles are saved using the same naming convention

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

# Batch process with custom output directory
./subscalpelmkv -b "*.mkv" -s eng,spa -o ./subtitles

# Batch process with custom filename template
./subscalpelmkv -b "Season 1/*.mkv" -s eng -f "{basename}-{language}.{extension}"
```

**Batch Processing Features:**
- **Automatic MKV filtering**: Only processes `.mkv` files from the pattern, ignoring other file types
- **Progress tracking**: Shows current file being processed (e.g., "Processing file 2/5")
- **Error resilience**: Continues processing remaining files if one fails
- **Summary reporting**: Displays total files processed and success/failure counts
- **Same options**: All track selection and output options work with batch processing

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

#### Output Control
```sh
# Custom output directory (automatically created if it doesn't exist)
./subscalpelmkv -x "path/to/video.mkv" -o ./subtitles

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
| `--output-dir` | `-o` | Custom output directory (automatically created if it doesn't exist, default: same as input file) |
| `--format` | `-f` | Custom filename template with placeholders |

#### Language Code Support

SubScalpelMKV supports comprehensive ISO language codes:
- **ISO 639-1**: All standard 2-letter language codes (e.g., `en`, `es`, `fr`, `de`, `ja`, `zh`)
- **ISO 639-2/B**: All standard 3-letter bibliographic codes (e.g., `eng`, `spa`, `fre`, `ger`, `jpn`, `chi`)

The tool automatically handles both formats and displays the full language name in track listings for clarity. This ensures compatibility with subtitles in virtually any language found in MKV files.

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
- **Custom Directory**: Use `-o` to specify a different output directory (automatically created if it doesn't exist)

**Directory Examples:**
```sh
# Save to generic directory
./subscalpelmkv -x path/to/video.mkv -o ./extracted-subtitles

# Organize by movie name
./subscalpelmkv -x path/to/video.mkv -o "./subtitles/Movie Name"
```
## License

This project is licensed under the MIT License. See the `LICENSE.md` file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## Acknowledgements

- [GMMK MKV Subtitles Extract](https://github.com/rhaseven7h/gmmmkvsubsextract)
- [MKVToolNix](https://mkvtoolnix.download/)
- [gocmd](https://github.com/devfacet/gocmd)
