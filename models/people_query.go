package models

import (
	"net/url"
	"strconv"

	"github.com/n0h4rt/chadango/utils"
)

// PeopleQuery represents a query for searching people.
type PeopleQuery struct {
	AgeFrom    int    // Minimum age
	AgeTo      int    // Maximum age
	Gender     string // Gender (B: both, M: male, F: female, N: unset)
	Username   string // Username
	Radius     int    // Radius
	Latitude   string // Latitude
	Longtitude string // Longitude
	Online     bool   // Online status
	Offset     int    // Offset
	Amount     int    // Amount
}

// GetForm returns the URL-encoded form values for the PeopleQuery.
func (pq *PeopleQuery) GetForm() url.Values {
	pq.AgeFrom = utils.Min(99, utils.Max(0, pq.AgeFrom))
	pq.AgeTo = utils.Min(99, utils.Max(0, pq.AgeTo))

	switch pq.Gender {
	case "B", "M", "F", "N":
	default:
		pq.Gender = "B"
	}

	pq.Radius = utils.Min(9999, utils.Max(0, pq.Radius))

	form := url.Values{
		"ami": {strconv.Itoa(pq.AgeFrom)},
		"ama": {strconv.Itoa(pq.AgeTo)},
		"s":   {pq.Gender},
	}

	if pq.Username != "" {
		form.Set("ss", pq.Username)
	}
	if pq.Radius > 0 {
		form.Set("r", strconv.Itoa(pq.Radius))
	}
	if pq.Latitude != "" && pq.Longtitude != "" {
		form.Set("la", pq.Latitude)
		form.Set("lo", pq.Longtitude)
	}
	if pq.Online {
		form.Set("o", "1")
	}

	form.Set("h5", "1")
	form.Set("f", strconv.Itoa(pq.Offset))
	form.Set("t", strconv.Itoa(pq.Offset+pq.Amount))

	return form
}

// NextOffset updates the offset to retrieve the next set of results.
func (pq *PeopleQuery) NextOffset() {
	pq.Offset += pq.Amount
}

// PeopleResult represents a query result from the [PeopleQuery].
type PeopleResult struct {
	Username string
	IsOnline bool
}
