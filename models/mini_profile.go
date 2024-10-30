package models

import (
	"encoding/xml"
	"net/url"
	"time"

	"github.com/n0h4rt/chadango/utils"
)

// MiniProfile represents a mini profile of a user.
type MiniProfile struct {
	Username string
	XMLName  xml.Name     `xml:"mod"`  // Tag name
	Body     QueryEscaped `xml:"body"` // Mini profile info
	Gender   string       `xml:"s"`    // Gender (M, F)
	Birth    BirthDate    `xml:"b"`    // Date of birth (yyyy-mm-dd)
	Location Location     `xml:"l"`    // Location
	Premium  PremiumDate  `xml:"d"`    // Premium expiration
}

func (m MiniProfile) PhotoLargeURL() string {
	return utils.UsernameToURL(API_PHOTO_FULL_IMG, m.Username)
}

// QueryEscaped represents a query-escaped string.
type QueryEscaped string

// UnmarshalXML unmarshals the XML data into the QueryEscaped value.
func (c *QueryEscaped) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var rawText string
	if err := d.DecodeElement(&rawText, &start); err != nil {
		return err
	}

	parsedText, _ := url.QueryUnescape(rawText)

	*c = QueryEscaped(parsedText)
	return nil
}

// BirthDate represents a birth date of a user.
type BirthDate time.Time

// UnmarshalXML unmarshals the XML data into the BirthDate value.
func (c *BirthDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var rawDate string
	if err := d.DecodeElement(&rawDate, &start); err != nil {
		return err
	}

	parsedDate, _ := time.Parse("2006-01-02", rawDate)

	*c = BirthDate(parsedDate)
	return nil
}

// Location represents the location information of a user.
type Location struct {
	Country   string  `xml:"c,attr"`    // Country name or US postal code
	G         string  `xml:"g,attr"`    // Reserved
	Latitude  float64 `xml:"lat,attr"`  // Latitude
	Longitude float64 `xml:"lon,attr"`  // Longitude
	Text      string  `xml:",chardata"` // String text of the location
}

// PremiumDate represents a premium date of a user.
type PremiumDate time.Time

// UnmarshalXML unmarshals the XML data into the PremiumDate value.
func (c *PremiumDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var rawDate string
	if err := d.DecodeElement(&rawDate, &start); err != nil {
		return err
	}

	parsedTimestamp, _ := utils.ParseTime(rawDate)

	*c = PremiumDate(parsedTimestamp)
	return nil
}
