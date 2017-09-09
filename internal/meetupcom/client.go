package meetupcom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

func NewClient(f func(*Client) error) (*Client, error) {
	c := Client{}
	if err := f(&c); err != nil {
		return nil, err
	}
	if c.APIKey == "" {
		return nil, fmt.Errorf("no API Key set")
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}
	if c.Timeout == 0 {
		c.Timeout = time.Second * 2
	}
	return &c, nil
}

func (c *Client) GetEventRSVP(ctx context.Context, urlName, eventID string) (*EventRSVP, error) {
	details, err := c.GetEventDetails(ctx, urlName, eventID)
	if err != nil {
		return nil, err
	}
	rsvp, err := c.GetRSVP(ctx, urlName, eventID)
	if err != nil {
		return nil, err
	}
	var result EventRSVP
	result.EventStatus = details.Status
	result.MaxCount = details.RSVPLimit
	for _, r := range *rsvp {
		switch r.Response {
		case "yes":
			result.YesCount = result.YesCount + 1 + r.Guests
			result.YesGuestCount += r.Guests
			result.YesMembers = append(result.YesMembers, r.Member)
		case "no":
			result.NoCount++
		case "waitlist":
			result.OnWaitlistCount++
		default:
			return nil, fmt.Errorf("unexpected rsvp response %v", r.Response)
		}
	}
	return &result, nil
}

func (c *Client) GetRSVP(ctx context.Context, urlName, eventID string) (*MeetupRSVPResponse, error) {
	resp, err := c.doGetRequest(ctx, fmt.Sprintf("/%s/events/%s/rsvps", url.PathEscape(urlName), url.PathEscape(eventID)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var rsvp MeetupRSVPResponse
	if err := json.NewDecoder(resp.Body).Decode(&rsvp); err != nil {
		return nil, err
	}
	return &rsvp, nil
}

func (c *Client) GetEventDetails(ctx context.Context, urlName, eventID string) (*MeetupEvent, error) {
	resp, err := c.doGetRequest(ctx, fmt.Sprintf("/%s/events/%s", url.PathEscape(urlName), url.PathEscape(eventID)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var evt MeetupEvent
	if err := json.NewDecoder(resp.Body).Decode(&evt); err != nil {
		return nil, err
	}
	return &evt, nil
}

func (c *Client) doGetRequest(ctx context.Context, path string) (*http.Response, error) {
	u := fmt.Sprintf("https://api.meetup.com%s", path)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	timeoutContext, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	return c.HTTPClient.Do(req.WithContext(timeoutContext))
}
