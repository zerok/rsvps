package meetupcom

type EventRSVP struct {
	EventTime       int64
	EventUTCOffset  int64
	EventStatus     string
	YesCount        int
	YesGuestCount   int
	YesMembers      []Member `json:"-"`
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

type Member struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Photo struct {
		ThumbLink string `json:"thumb_link"`
	} `json:"photo"`
}

type MeetupRSVPResponse []MeetupRSVP

type MeetupRSVP struct {
	Response string `json:"response"`
	Guests   int    `json:"guests"`
	Member   Member `json:"member"`
}
