package main

import "github.com/zerok/rsvps/internal/meetupcom"

type queryRequest struct {
	Meetups []string `json:"meetups"`
}

type EventRSVP struct {
	EventURL string               `json:"eventURL"`
	RSVP     *meetupcom.EventRSVP `json:"rsvp"`
}

type EventRSVPs map[string]EventRSVP
