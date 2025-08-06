package models

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/zeebo/blake3"
)

type PageDataModel struct {
	Metadata Metadata `json:"metadata"`
	Content  []byte   `json:"content"`
}

type Metadata struct {
	URL        string    `json:"url"`
	Host       string    `json:"host"`
	HTMLHash   string    `json:"html_hash"`
	Latency    Latency   `json:"latency_ms"`
	Timestamp  time.Time `json:"time"`
	ContentLen int       `json:"content_length"`
}
type Latency time.Duration

func (l Latency) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Duration(l).Milliseconds(), 10)), nil
}

func (l *Latency) UnmarshalJSON(data []byte) error {
	val, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid Latency: %v", err)
	}
	if val < 0 {
		return fmt.Errorf("Latency cannot be negative")
	}
	*l = Latency(time.Duration(val) * time.Millisecond)
	return nil
}

func NewPageDataModel(url, host string, content []byte, latency time.Duration) (*PageDataModel, error) {
	htmlHahs := blake3.Sum256(content)

	hashString := hex.EncodeToString(htmlHahs[:])
	metadata := Metadata{
		URL:        url,
		Host:       host,
		HTMLHash:   hashString,
		Latency:    Latency(latency),
		Timestamp:  time.Now(),
		ContentLen: len(content),
	}

	return &PageDataModel{
		Metadata: metadata,
		Content:  content,
	}, nil
}

func (m *PageDataModel) IsValid() bool {
	return m.Content != nil
}
