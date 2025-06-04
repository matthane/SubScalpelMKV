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

// Language code mapping from ISO 639-1 (2-letter) to ISO 639-2/B (3-letter)
// This includes comprehensive ISO 639 language code support
var LanguageCodeMapping = map[string]string{
	// Major languages
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
	
	// Additional ISO 639-1 codes
	"aa": "aar", // Afar
	"ab": "abk", // Abkhazian
	"ae": "ave", // Avestan
	"ak": "aka", // Akan
	"an": "arg", // Aragonese
	"as": "asm", // Assamese
	"av": "ava", // Avaric
	"ay": "aym", // Aymara
	"az": "aze", // Azerbaijani
	"ba": "bak", // Bashkir
	"bh": "bih", // Bihari languages
	"bi": "bis", // Bislama
	"bm": "bam", // Bambara
	"bo": "tib", // Tibetan
	"br": "bre", // Breton
	"bs": "bos", // Bosnian
	"ce": "che", // Chechen
	"ch": "cha", // Chamorro
	"co": "cos", // Corsican
	"cr": "cre", // Cree
	"cu": "chu", // Church Slavic
	"cv": "chv", // Chuvash
	"dv": "div", // Divehi
	"dz": "dzo", // Dzongkha
	"ee": "ewe", // Ewe
	"eo": "epo", // Esperanto
	"ff": "ful", // Fulah
	"fj": "fij", // Fijian
	"fo": "fao", // Faroese
	"fy": "fry", // Western Frisian
	"gd": "gla", // Scottish Gaelic
	"gn": "grn", // Guarani
	"gv": "glv", // Manx
	"ha": "hau", // Hausa
	"ho": "hmo", // Hiri Motu
	"ht": "hat", // Haitian
	"hz": "her", // Herero
	"ia": "ina", // Interlingua
	"id": "ind", // Indonesian
	"ie": "ile", // Interlingue
	"ig": "ibo", // Igbo
	"ii": "iii", // Sichuan Yi
	"ik": "ipk", // Inupiaq
	"io": "ido", // Ido
	"iu": "iku", // Inuktitut
	"jv": "jav", // Javanese
	"kg": "kon", // Kongo
	"ki": "kik", // Kikuyu
	"kj": "kua", // Kuanyama
	"kk": "kaz", // Kazakh
	"kl": "kal", // Kalaallisut
	"kr": "kau", // Kanuri
	"ks": "kas", // Kashmiri
	"ku": "kur", // Kurdish
	"kv": "kom", // Komi
	"kw": "cor", // Cornish
	"ky": "kir", // Kirghiz
	"la": "lat", // Latin
	"lb": "ltz", // Luxembourgish
	"lg": "lug", // Ganda
	"li": "lim", // Limburgish
	"ln": "lin", // Lingala
	"lu": "lub", // Luba-Katanga
	"mg": "mlg", // Malagasy
	"mh": "mah", // Marshallese
	"mi": "mao", // Maori
	"mn": "mon", // Mongolian
	"mo": "mol", // Moldavian
	"ms": "may", // Malay
	"na": "nau", // Nauru
	"nb": "nob", // Norwegian Bokmål
	"nd": "nde", // North Ndebele
	"ng": "ndo", // Ndonga
	"nn": "nno", // Norwegian Nynorsk
	"nr": "nbl", // South Ndebele
	"nv": "nav", // Navajo
	"ny": "nya", // Chichewa
	"oc": "oci", // Occitan
	"oj": "oji", // Ojibwa
	"om": "orm", // Oromo
	"or": "ori", // Oriya
	"os": "oss", // Ossetian
	"pi": "pli", // Pali
	"ps": "pus", // Pashto
	"qu": "que", // Quechua
	"rm": "roh", // Romansh
	"rn": "run", // Rundi
	"rw": "kin", // Kinyarwanda
	"sa": "san", // Sanskrit
	"sc": "srd", // Sardinian
	"sd": "snd", // Sindhi
	"se": "sme", // Northern Sami
	"sg": "sag", // Sango
	"sh": "srp", // Serbo-Croatian (deprecated, use sr)
	"sm": "smo", // Samoan
	"sn": "sna", // Shona
	"so": "som", // Somali
	"sr": "srp", // Serbian
	"ss": "ssw", // Swati
	"st": "sot", // Southern Sotho
	"su": "sun", // Sundanese
	"tg": "tgk", // Tajik
	"ti": "tir", // Tigrinya
	"tk": "tuk", // Turkmen
	"tl": "tgl", // Tagalog
	"tn": "tsn", // Tswana
	"to": "ton", // Tonga
	"ts": "tso", // Tsonga
	"tt": "tat", // Tatar
	"tw": "twi", // Twi
	"ty": "tah", // Tahitian
	"ug": "uig", // Uighur
	"uz": "uzb", // Uzbek
	"ve": "ven", // Venda
	"vo": "vol", // Volapük
	"wa": "wln", // Walloon
	"wo": "wol", // Wolof
	"xh": "xho", // Xhosa
	"yi": "yid", // Yiddish
	"yo": "yor", // Yoruba
	"za": "zha", // Zhuang
}

// LanguageNames maps language codes (both 2 and 3 letter) to full language names
var LanguageNames = map[string]string{
	// 2-letter codes - Major languages
	"en": "English",
	"es": "Spanish",
	"fr": "French",
	"de": "German",
	"it": "Italian",
	"pt": "Portuguese",
	"ru": "Russian",
	"ja": "Japanese",
	"ko": "Korean",
	"zh": "Chinese",
	"ar": "Arabic",
	"hi": "Hindi",
	"th": "Thai",
	"vi": "Vietnamese",
	"tr": "Turkish",
	"pl": "Polish",
	"nl": "Dutch",
	"sv": "Swedish",
	"da": "Danish",
	"no": "Norwegian",
	"fi": "Finnish",
	"cs": "Czech",
	"hu": "Hungarian",
	"ro": "Romanian",
	"bg": "Bulgarian",
	"hr": "Croatian",
	"sk": "Slovak",
	"sl": "Slovenian",
	"et": "Estonian",
	"lv": "Latvian",
	"lt": "Lithuanian",
	"el": "Greek",
	"he": "Hebrew",
	"fa": "Persian",
	"ur": "Urdu",
	"bn": "Bengali",
	"ta": "Tamil",
	"te": "Telugu",
	"ml": "Malayalam",
	"kn": "Kannada",
	"gu": "Gujarati",
	"pa": "Punjabi",
	"mr": "Marathi",
	"ne": "Nepali",
	"si": "Sinhala",
	"my": "Burmese",
	"km": "Khmer",
	"lo": "Lao",
	"ka": "Georgian",
	"am": "Amharic",
	"sw": "Swahili",
	"zu": "Zulu",
	"af": "Afrikaans",
	"is": "Icelandic",
	"ga": "Irish",
	"cy": "Welsh",
	"eu": "Basque",
	"ca": "Catalan",
	"gl": "Galician",
	"mt": "Maltese",
	"sq": "Albanian",
	"mk": "Macedonian",
	"be": "Belarusian",
	"uk": "Ukrainian",
	
	// Additional 2-letter codes
	"aa": "Afar",
	"ab": "Abkhazian",
	"ae": "Avestan",
	"ak": "Akan",
	"an": "Aragonese",
	"as": "Assamese",
	"av": "Avaric",
	"ay": "Aymara",
	"az": "Azerbaijani",
	"ba": "Bashkir",
	"bh": "Bihari languages",
	"bi": "Bislama",
	"bm": "Bambara",
	"bo": "Tibetan",
	"br": "Breton",
	"bs": "Bosnian",
	"ce": "Chechen",
	"ch": "Chamorro",
	"co": "Corsican",
	"cr": "Cree",
	"cu": "Church Slavic",
	"cv": "Chuvash",
	"dv": "Divehi",
	"dz": "Dzongkha",
	"ee": "Ewe",
	"eo": "Esperanto",
	"ff": "Fulah",
	"fj": "Fijian",
	"fo": "Faroese",
	"fy": "Western Frisian",
	"gd": "Scottish Gaelic",
	"gn": "Guarani",
	"gv": "Manx",
	"ha": "Hausa",
	"ho": "Hiri Motu",
	"ht": "Haitian",
	"hz": "Herero",
	"ia": "Interlingua",
	"id": "Indonesian",
	"ie": "Interlingue",
	"ig": "Igbo",
	"ii": "Sichuan Yi",
	"ik": "Inupiaq",
	"io": "Ido",
	"iu": "Inuktitut",
	"jv": "Javanese",
	"kg": "Kongo",
	"ki": "Kikuyu",
	"kj": "Kuanyama",
	"kk": "Kazakh",
	"kl": "Kalaallisut",
	"kr": "Kanuri",
	"ks": "Kashmiri",
	"ku": "Kurdish",
	"kv": "Komi",
	"kw": "Cornish",
	"ky": "Kirghiz",
	"la": "Latin",
	"lb": "Luxembourgish",
	"lg": "Ganda",
	"li": "Limburgish",
	"ln": "Lingala",
	"lu": "Luba-Katanga",
	"mg": "Malagasy",
	"mh": "Marshallese",
	"mi": "Maori",
	"mn": "Mongolian",
	"mo": "Moldavian",
	"ms": "Malay",
	"na": "Nauru",
	"nb": "Norwegian Bokmål",
	"nd": "North Ndebele",
	"ng": "Ndonga",
	"nn": "Norwegian Nynorsk",
	"nr": "South Ndebele",
	"nv": "Navajo",
	"ny": "Chichewa",
	"oc": "Occitan",
	"oj": "Ojibwa",
	"om": "Oromo",
	"or": "Oriya",
	"os": "Ossetian",
	"pi": "Pali",
	"ps": "Pashto",
	"qu": "Quechua",
	"rm": "Romansh",
	"rn": "Rundi",
	"rw": "Kinyarwanda",
	"sa": "Sanskrit",
	"sc": "Sardinian",
	"sd": "Sindhi",
	"se": "Northern Sami",
	"sg": "Sango",
	"sh": "Serbo-Croatian",
	"sm": "Samoan",
	"sn": "Shona",
	"so": "Somali",
	"sr": "Serbian",
	"ss": "Swati",
	"st": "Southern Sotho",
	"su": "Sundanese",
	"tg": "Tajik",
	"ti": "Tigrinya",
	"tk": "Turkmen",
	"tl": "Tagalog",
	"tn": "Tswana",
	"to": "Tonga",
	"ts": "Tsonga",
	"tt": "Tatar",
	"tw": "Twi",
	"ty": "Tahitian",
	"ug": "Uighur",
	"uz": "Uzbek",
	"ve": "Venda",
	"vo": "Volapük",
	"wa": "Walloon",
	"wo": "Wolof",
	"xh": "Xhosa",
	"yi": "Yiddish",
	"yo": "Yoruba",
	"za": "Zhuang",
	// 3-letter codes - Major languages
	"eng": "English",
	"spa": "Spanish",
	"fre": "French",
	"ger": "German",
	"ita": "Italian",
	"por": "Portuguese",
	"rus": "Russian",
	"jpn": "Japanese",
	"kor": "Korean",
	"chi": "Chinese",
	"ara": "Arabic",
	"hin": "Hindi",
	"tha": "Thai",
	"vie": "Vietnamese",
	"tur": "Turkish",
	"pol": "Polish",
	"dut": "Dutch",
	"swe": "Swedish",
	"dan": "Danish",
	"nor": "Norwegian",
	"fin": "Finnish",
	"cze": "Czech",
	"hun": "Hungarian",
	"rum": "Romanian",
	"bul": "Bulgarian",
	"hrv": "Croatian",
	"slo": "Slovak",
	"slv": "Slovenian",
	"est": "Estonian",
	"lav": "Latvian",
	"lit": "Lithuanian",
	"gre": "Greek",
	"heb": "Hebrew",
	"per": "Persian",
	"urd": "Urdu",
	"ben": "Bengali",
	"tam": "Tamil",
	"tel": "Telugu",
	"mal": "Malayalam",
	"kan": "Kannada",
	"guj": "Gujarati",
	"pan": "Punjabi",
	"mar": "Marathi",
	"nep": "Nepali",
	"sin": "Sinhala",
	"bur": "Burmese",
	"khm": "Khmer",
	"lao": "Lao",
	"geo": "Georgian",
	"amh": "Amharic",
	"swa": "Swahili",
	"zul": "Zulu",
	"afr": "Afrikaans",
	"ice": "Icelandic",
	"gle": "Irish",
	"wel": "Welsh",
	"baq": "Basque",
	"cat": "Catalan",
	"glg": "Galician",
	"mlt": "Maltese",
	"alb": "Albanian",
	"mac": "Macedonian",
	"bel": "Belarusian",
	"ukr": "Ukrainian",
	
	// Additional 3-letter codes
	"aar": "Afar",
	"abk": "Abkhazian",
	"ave": "Avestan",
	"aka": "Akan",
	"arg": "Aragonese",
	"asm": "Assamese",
	"ava": "Avaric",
	"aym": "Aymara",
	"aze": "Azerbaijani",
	"bak": "Bashkir",
	"bih": "Bihari languages",
	"bis": "Bislama",
	"bam": "Bambara",
	"tib": "Tibetan",
	"bre": "Breton",
	"bos": "Bosnian",
	"che": "Chechen",
	"cha": "Chamorro",
	"cos": "Corsican",
	"cre": "Cree",
	"chu": "Church Slavic",
	"chv": "Chuvash",
	"div": "Divehi",
	"dzo": "Dzongkha",
	"ewe": "Ewe",
	"epo": "Esperanto",
	"ful": "Fulah",
	"fij": "Fijian",
	"fao": "Faroese",
	"fry": "Western Frisian",
	"gla": "Scottish Gaelic",
	"grn": "Guarani",
	"glv": "Manx",
	"hau": "Hausa",
	"hmo": "Hiri Motu",
	"hat": "Haitian",
	"her": "Herero",
	"ina": "Interlingua",
	"ind": "Indonesian",
	"ile": "Interlingue",
	"ibo": "Igbo",
	"iii": "Sichuan Yi",
	"ipk": "Inupiaq",
	"ido": "Ido",
	"iku": "Inuktitut",
	"jav": "Javanese",
	"kon": "Kongo",
	"kik": "Kikuyu",
	"kua": "Kuanyama",
	"kaz": "Kazakh",
	"kal": "Kalaallisut",
	"kau": "Kanuri",
	"kas": "Kashmiri",
	"kur": "Kurdish",
	"kom": "Komi",
	"cor": "Cornish",
	"kir": "Kirghiz",
	"lat": "Latin",
	"ltz": "Luxembourgish",
	"lug": "Ganda",
	"lim": "Limburgish",
	"lin": "Lingala",
	"lub": "Luba-Katanga",
	"mlg": "Malagasy",
	"mah": "Marshallese",
	"mao": "Maori",
	"mon": "Mongolian",
	"mol": "Moldavian",
	"may": "Malay",
	"nau": "Nauru",
	"nob": "Norwegian Bokmål",
	"nde": "North Ndebele",
	"ndo": "Ndonga",
	"nno": "Norwegian Nynorsk",
	"nbl": "South Ndebele",
	"nav": "Navajo",
	"nya": "Chichewa",
	"oci": "Occitan",
	"oji": "Ojibwa",
	"orm": "Oromo",
	"ori": "Oriya",
	"oss": "Ossetian",
	"pli": "Pali",
	"pus": "Pashto",
	"que": "Quechua",
	"roh": "Romansh",
	"run": "Rundi",
	"kin": "Kinyarwanda",
	"san": "Sanskrit",
	"srd": "Sardinian",
	"snd": "Sindhi",
	"sme": "Northern Sami",
	"sag": "Sango",
	"srp": "Serbian",
	"smo": "Samoan",
	"sna": "Shona",
	"som": "Somali",
	"ssw": "Swati",
	"sot": "Southern Sotho",
	"sun": "Sundanese",
	"tgk": "Tajik",
	"tir": "Tigrinya",
	"tuk": "Turkmen",
	"tgl": "Tagalog",
	"tsn": "Tswana",
	"ton": "Tonga",
	"tso": "Tsonga",
	"tat": "Tatar",
	"twi": "Twi",
	"tah": "Tahitian",
	"uig": "Uighur",
	"uzb": "Uzbek",
	"ven": "Venda",
	"vol": "Volapük",
	"wln": "Walloon",
	"wol": "Wolof",
	"xho": "Xhosa",
	"yid": "Yiddish",
	"yor": "Yoruba",
	"zha": "Zhuang",
}

// GetLanguageName returns the full language name for a given language code
func GetLanguageName(code string) string {
	if name, exists := LanguageNames[strings.ToLower(code)]; exists {
		return name
	}
	return code // Return the code itself if no name is found
}

// MatchesLanguageFilter checks if a track language matches the specified filter
// Supports both 2-letter (ISO 639-1) and 3-letter (ISO 639-2) language codes
func MatchesLanguageFilter(trackLanguage, filterLanguage string) bool {
	if filterLanguage == "" {
		return true // No filter specified, match all
	}

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
	FormatFilters []string // Subtitle format filters (e.g., "srt", "ass", "sup")
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

// GetSubtitleFormatFromCodec returns the subtitle format (extension) for a given codec
func GetSubtitleFormatFromCodec(codecId string) string {
	if ext, exists := SubtitleExtensionByCodec[codecId]; exists {
		return ext
	}
	return "srt" // fallback
}

// MatchesFormatFilter checks if a track format matches the specified filter
func MatchesFormatFilter(codecId, formatFilter string) bool {
	if formatFilter == "" {
		return true // No filter specified, match all
	}

	trackFormat := GetSubtitleFormatFromCodec(codecId)
	return strings.EqualFold(trackFormat, formatFilter)
}

// ExtractionJob represents a single subtitle extraction task
type ExtractionJob struct {
	Track         MKVTrack
	OriginalTrack MKVTrack
	OutFileName   string
	MksFileName   string
}

// ExtractionResult represents the result of an extraction operation
type ExtractionResult struct {
	Job   ExtractionJob
	Error error
}

// BatchFileInfo represents information about a file in batch processing
type BatchFileInfo struct {
	FileName       string
	FilePath       string
	SubtitleCount  int
	LanguageCodes  []string
	SubtitleFormats []string
	HasError       bool
	ErrorMessage   string
}
