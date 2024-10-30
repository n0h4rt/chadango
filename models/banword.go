package models

import (
	"net/url"
	"strings"
)

// BanWord represents a banned word.
type BanWord struct {
	WholeWords string `json:"wholeWords"` // Whole words to be banned.
	Words      string `json:"words"`      // Words to be banned.
}

// SetWhole sets the whole banned words.
func (bw *BanWord) SetWhole(exact []string) {
	bw.WholeWords = url.QueryEscape(strings.Join(exact, ","))
}

// GetWhole returns the whole banned words as a slice of strings.
func (bw *BanWord) GetWhole() []string {
	unquoted, _ := url.QueryUnescape(bw.WholeWords)
	return strings.Split(unquoted, ",")
}

// SetPartial sets the partial banned words.
func (bw *BanWord) SetPartial(partial []string) {
	bw.Words = url.QueryEscape(strings.Join(partial, ","))
}

// GetPartial returns the partial banned words as a slice of strings.
func (bw *BanWord) GetPartial() []string {
	unquoted, _ := url.QueryUnescape(bw.Words)
	return strings.Split(unquoted, ",")
}
