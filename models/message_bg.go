package models

import (
	"net/url"
	"strconv"

	"github.com/n0h4rt/chadango/utils"
)

// MessageBackground represents the message background information of a user.
type MessageBackground struct {
	Align        string  `xml:"align,attr"`  // Background image alignment
	Alpha        float64 `xml:"bgalp,attr"`  // Background color transparency
	Color        string  `xml:"bgc,attr"`    // Background color
	HasRecording int64   `xml:"hasrec,attr"` // Media recording timestamp (ms)
	ImageAlpha   float64 `xml:"ialp,attr"`   // Background image transparency
	IsVid        bool    `xml:"isvid,attr"`  // Is the Media a video?
	Tile         bool    `xml:"tile,attr"`   // Tile image?
	UseImage     bool    `xml:"useimg,attr"` // Use image?

	Username string
}

// GetForm returns the URL-encoded form values for the `MessageBackground`.
func (mb *MessageBackground) GetForm() url.Values {
	switch mb.Align {
	case "tr", "br", "tl", "bl":
	default:
		mb.Align = "tl"
	}
	mb.Alpha = utils.Min(100, utils.Max(0, mb.Alpha))
	if mb.Color == "" {
		mb.Color = "ffffff"
	}
	mb.ImageAlpha = utils.Min(100, utils.Max(0, mb.ImageAlpha))

	form := url.Values{
		"align":  {mb.Align},
		"bgalp":  {strconv.FormatFloat(mb.Alpha, 'f', 1, 64)},
		"bgc":    {mb.Color},
		"hasrec": {strconv.FormatInt(mb.HasRecording, 10)},
		"ialp":   {strconv.FormatFloat(mb.ImageAlpha, 'f', 1, 64)},
		"isvid":  {utils.BoolZeroOrOne(mb.IsVid)},
		"tile":   {utils.BoolZeroOrOne(mb.Tile)},
		"useimg": {utils.BoolZeroOrOne(mb.UseImage)},
	}

	return form
}

// GetImageURL returns an url of the message background image.
//
// It formated as https://ust.chatango.com/profileimg/u/s/username/msgbg.jpg
func (mb *MessageBackground) GetImageURL() string {
	return utils.UsernameToURL(API_MSG_BG_IMG, mb.Username)
}
