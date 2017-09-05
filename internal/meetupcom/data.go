package meetupcom

type EventRSVP struct {
	EventTime       int64
	EventUTCOffset  int64
	EventStatus     string
	YesCount        int
	NoCount         int
	MaxCount        int
	OnWaitlistCount int
}

type MeetupEvent struct {
	RSVPLimit int    `json:"rsvp_limit"`
	Status    string `json:"status"`
	Time      int64  `json:"time"`
	UTCOffset int64  `json:"utc_offset"`
}

type MeetupRSVPResponse []MeetupRSVP

type MeetupRSVP struct {
	Response string `json:"response"`
	Guests   int    `json:"guests"`
}
