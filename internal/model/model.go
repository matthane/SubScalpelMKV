package model

import "math/big"

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
