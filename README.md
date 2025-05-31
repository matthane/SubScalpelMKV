# 🗡️ SubScalpelMKV

`subscalpelmkv` is a cross-platform command-line tool written in Go for extracting subtitles from MKV files quickly and precisely. It is a fork/overhaul of [GMM MKV Subtitles Extract](https://github.com/rhaseven7h/gmmmkvsubsextract) with many new features for increased speed and capability.

## Features

- Extract subtitle tracks from MKV files using MKVToolNix
- Support for SRT, ASS, and SUP (PGS) subtitle formats
- Track export selection using language codes, track numbers, or any combination of both
- Automatic file naming based on track properties (language, number, name, forced status)
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

Drag an MKV file onto the executable for interactive mode:

1. The tool analyzes the file and displays available subtitle tracks

2. Choose to extract all tracks or make a custom selection

3. For custom selection, enter:
    - Language codes: `eng`, `spa`, `fre`
    - Track numbers: `3`, `5`, `7`
    - Mixed selection: `eng,3,spa,7`

### Command Line Mode

#### Basic Usage
```sh
# Extract all subtitle tracks
./subscalpelmkv -x "path/to/video.mkv"
```

#### Language Filtering
```sh
# Single language
./subscalpelmkv -x "path/to/video.mkv" -l eng

# Multiple languages
./subscalpelmkv -x "path/to/video.mkv" -l eng,spa,fre
```

#### Track Selection
```sh
# Specific track numbers
./subscalpelmkv -x "path/to/video.mkv" -t 3,5,7

# Mixed language and track selection
./subscalpelmkv -x "path/to/video.mkv" -s eng,3,spa,7
```

#### Info Flag Usage
```sh
# Show information about available subtitle tracks
./subscalpelmkv -i "path/to/video.mkv"
```

#### Command Line Options

| Option | Short | Description |
|--------|-------|-------------|
| `--extract` | `-x` | Path to MKV file (required) |
| `--language` | `-l` | Language codes (comma-separated) |
| `--tracks` | `-t` | Track numbers (comma-separated) |
| `--selection` | `-s` | Mixed language codes and track numbers |
| `--info` | `-i` | Show information about available subtitle tracks |
| `--help` | `-h` | Show help message |

#### Supported Language Codes
- **2-letter (ISO 639-1)**: `en`, `es`, `fr`, `de`, `it`, `pt`, `ru`, `ja`, `ko`, `zh`, `ar`, `hi`, `th`, `vi`, `tr`, `pl`, `nl`, `sv`, `da`, `no`, `fi`, `cs`, `hu`, `ro`, `bg`, `hr`, `sk`, `sl`, `et`, `lv`, `lt`, `el`
- **3-letter (ISO 639-2)**: `eng`, `spa`, `fre`, `ger`, `ita`, `por`, `rus`, `jpn`, `kor`, `chi`, `ara`, `hin`, `tha`, `vie`, `tur`, `pol`, `dut`, `swe`, `dan`, `nor`, `fin`, `cze`, `hun`, `rum`, `bul`, `hrv`, `slo`, `slv`, `est`, `lav`, `lit`, `gre`

The tool will extract subtitle tracks from `example.mkv` and save them with appropriate file names based on track properties. When using language filtering, only tracks matching the specified language code will be extracted.

### Output File Naming

Output files are named using the pattern:
```
<original_filename>.<language>.<track_number>[.forced/.default].<extension>
```

Examples:
- `movie.eng.001.srt` - English SRT subtitle, track 1
- `movie.spa.002.ass` - Spanish ASS subtitle, track 2
- `movie.eng.003.forced.sup` - English forced SUP subtitle, track 3
- `movie.fre.004.default.sup` - French default SUP subtitle, track 4
## License

This project is licensed under the MIT License. See the `LICENSE.md` file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## Acknowledgements

- [GMMK MKV Subtitles Extract](https://github.com/rhaseven7h/gmmmkvsubsextract)
- [MKVToolNix](https://mkvtoolnix.download/)
- [gocmd](https://github.com/devfacet/gocmd)
