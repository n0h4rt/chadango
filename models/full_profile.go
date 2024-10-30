package models

import "encoding/xml"

// FullProfile represents a full profile of a user.
type FullProfile struct {
	XMLName xml.Name     `xml:"mod"`  // Tag name
	Body    QueryEscaped `xml:"body"` // Full profile info
	T       string       `xml:"t"`    // Reserved
}
