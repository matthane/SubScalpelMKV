package model

import (
	"math/big"
	"strings"
)

// MKVTrackProperties represents the properties of an MKV track
type MKVTrackProperties struct {
	CodecId              string  `json:"codec_id"`
	TrackName            string  `json:"track_name"`
	Encoding             string  `json:"encoding"`
	Language             string  `json:"language"`
	Number               int     `json:"number"`
	Forced               bool    `json:"forced_track"`
	Default              bool    `json:"default_track"`
	Enabled              bool    `json:"enabled_track"`
	TextSubtitles        bool    `json:"text_subtitles"`
	NumberOfIndexEntries int     `json:"num_index_entries"`
	Duration             string  `json:"tag_duration"`
	UId                  big.Int `json:"uid"`
}

// MKVTrack represents a track in an MKV file
type MKVTrack struct {
	Codec      string             `json:"codec"`
	Id         int                `json:"id"`
	Type       string             `json:"type"`
	Properties MKVTrackProperties `json:"properties"`
}

// MKVContainer represents the container information of an MKV file
type MKVContainer struct {
	Type string `json:"type"`
}

// Language code mapping from ISO 639-1 (2-letter) to ISO 639-2 (3-letter)
var LanguageCodeMapping = map[string]string{
	"en": "eng", // English
	"es": "spa", // Spanish
	"fr": "fre", // French
	"de": "ger", // German
	"it": "ita", // Italian
	"pt": "por", // Portuguese
	"ru": "rus", // Russian
	"ja": "jpn", // Japanese
	"ko": "kor", // Korean
	"zh": "chi", // Chinese
	"ar": "ara", // Arabic
	"hi": "hin", // Hindi
	"th": "tha", // Thai
	"vi": "vie", // Vietnamese
	"tr": "tur", // Turkish
	"pl": "pol", // Polish
	"nl": "dut", // Dutch
	"sv": "swe", // Swedish
	"da": "dan", // Danish
	"no": "nor", // Norwegian
	"fi": "fin", // Finnish
	"cs": "cze", // Czech
	"hu": "hun", // Hungarian
	"ro": "rum", // Romanian
	"bg": "bul", // Bulgarian
	"hr": "hrv", // Croatian
	"sk": "slo", // Slovak
	"sl": "slv", // Slovenian
	"et": "est", // Estonian
	"lv": "lav", // Latvian
	"lt": "lit", // Lithuanian
	"el": "gre", // Greek
	"he": "heb", // Hebrew
	"fa": "per", // Persian
	"ur": "urd", // Urdu
	"bn": "ben", // Bengali
	"ta": "tam", // Tamil
	"te": "tel", // Telugu
	"ml": "mal", // Malayalam
	"kn": "kan", // Kannada
	"gu": "guj", // Gujarati
	"pa": "pan", // Punjabi
	"mr": "mar", // Marathi
	"ne": "nep", // Nepali
	"si": "sin", // Sinhala
	"my": "bur", // Burmese
	"km": "khm", // Khmer
	"lo": "lao", // Lao
	"ka": "geo", // Georgian
	"am": "amh", // Amharic
	"sw": "swa", // Swahili
	"zu": "zul", // Zulu
	"af": "afr", // Afrikaans
	"is": "ice", // Icelandic
	"ga": "gle", // Irish
	"cy": "wel", // Welsh
	"eu": "baq", // Basque
	"ca": "cat", // Catalan
	"gl": "glg", // Galician
	"mt": "mlt", // Maltese
	"sq": "alb", // Albanian
	"mk": "mac", // Macedonian
	"be": "bel", // Belarusian
	"uk": "ukr", // Ukrainian
}

// MatchesLanguageFilter checks if a track language matches the specified filter
// Supports both 2-letter (ISO 639-1) and 3-letter (ISO 639-2) language codes
func MatchesLanguageFilter(trackLanguage, filterLanguage string) bool {
	if filterLanguage == "" {
		return true // No filter specified, match all
	}

	// Direct match
	if strings.EqualFold(trackLanguage, filterLanguage) {
		return true
	}

	// Check if filter is 2-letter code and track uses 3-letter code
	if len(filterLanguage) == 2 {
		if mappedCode, exists := LanguageCodeMapping[strings.ToLower(filterLanguage)]; exists {
			return strings.EqualFold(trackLanguage, mappedCode)
		}
	}

	// Check if filter is 3-letter code and track uses 2-letter code
	if len(filterLanguage) == 3 {
		for twoLetter, threeLetter := range LanguageCodeMapping {
			if strings.EqualFold(filterLanguage, threeLetter) {
				return strings.EqualFold(trackLanguage, twoLetter)
			}
		}
	}

	return false
}

// MKVInfo represents the complete information about an MKV file
type MKVInfo struct {
	Tracks    []MKVTrack   `json:"tracks"`
	Container MKVContainer `json:"container"`
}

// TrackSelection represents the user's track selection criteria
type TrackSelection struct {
	LanguageCodes []string
	TrackNumbers  []int
}

// OutputConfig represents output configuration options
type OutputConfig struct {
	OutputDir string // Custom output directory
	Template  string // Filename template with placeholders
	CreateDir bool   // Whether to create output directory if it doesn't exist
}

// DefaultOutputTemplate is the default filename template
const DefaultOutputTemplate = "{basename}.{language}.{trackno}.{trackname}.{forced}.{default}.{extension}"

// SubtitleExtensionByCodec maps codec IDs to file extensions
var SubtitleExtensionByCodec = map[string]string{
	// Text-based subtitle formats
	"S_TEXT/UTF8":   "srt",
	"S_TEXT/ASS":    "ass",
	"S_TEXT/SSA":    "ssa",
	"S_TEXT/WEBVTT": "vtt",
	"S_TEXT/USF":    "usf",
	"S_ASS":         "ass",
	"S_SSA":         "ssa",

	// Image-based subtitle formats
	"S_HDMV/PGS":  "sup",
	"S_VOBSUB":    "sub",
	"S_DVBSUB":    "sub",
	"S_IMAGE/BMP": "bmp",

	// Legacy and other formats
	"S_KATE":        "kate",
	"S_TEXT/PLAIN":  "txt",
	"S_HDMV/TEXTST": "sup",
}
