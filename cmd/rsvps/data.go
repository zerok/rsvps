package main

import "github.com/zerok/rsvps/internal/meetupcom"

type queryRequest struct {
	Meetups []string `json:"meetups"`
}

type queryResposne struct {
	Summary Summary               `json:"summary"`
	Meetups map[string]*EventRSVP `json:"meetups"`
}

type EventRSVP struct {
	EventURL string               `json:"eventURL"`
	RSVP     *meetupcom.EventRSVP `json:"rsvp"`
}

type Summary struct {
	Profiles     map[string]Profile `json:"profiles"`
	AllYesIDs    []string           `json:"allYesIDs"`
	AllYesGuests int                `json:"allYesGuests"`
}

type Profile struct {
	ID       string `json:"id"`
	ThumbURL string `json:"thumbURL"`
	Name     string `json:"name"`
}

type EventRSVPs map[string]EventRSVP
