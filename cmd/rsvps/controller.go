package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bluele/gcache"
	"github.com/zerok/rsvps/internal/meetupcom"
	"github.com/zerok/rsvps/internal/whitelist"
)

type controller struct {
	client                        *meetupcom.Client
	cache                         gcache.Cache
	log                           *logrus.Logger
	whitelist                     *whitelist.Whitelist
	upcomingEventsCachingDuration time.Duration
	pastEventsCachingDuration     time.Duration
}

func newController(f func(*controller) error) (*controller, error) {
	c := &controller{}
	if err := f(c); err != nil {
		return nil, err
	}
	if c.log == nil {
		c.log = logrus.New()
	}
	return c, nil
}

func (c *controller) handleInvalidInput(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		c.log.WithError(err).Debug(msg)
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func (c *controller) handleInternalError(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		c.log.WithError(err).Debug(msg)
	}
	http.Error(w, msg, http.StatusInternalServerError)
}

func (c *controller) handleGetRSVPs(w http.ResponseWriter, r *http.Request) {
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.handleInvalidInput(w, "Failed to decode JSON", err)
		return
	}
	defer r.Body.Close()
	ctx := r.Context()
	response := queryResposne{}
	result := make(map[string]*EventRSVP)
	knownProfiles := make(map[int]struct{})
	s := Summary{
		Profiles:  make(map[string]Profile),
		AllYesIDs: make([]string, 0, 0),
	}
	// Make sure that the meetups are in the list of allowed meetups.
	for _, meetup := range req.Meetups {
		select {
		case <-ctx.Done():
			c.handleInternalError(w, "request context cancelled", nil)
			return
		default:
		}
		if !c.whitelist.Contains(meetup) {
			c.log.Warnf("%s not in whitelist", meetup)
			continue
		}
		group, event, err := meetupcom.ParseURL(meetup)
		if err != nil {
			c.log.WithError(err).Debugf("failed to decode %s", meetup)
			continue
		}
		cacheKey := fmt.Sprintf("meetupcom:%s:%s", group, event)
		cached, err := c.cache.Get(cacheKey)
		if err != nil {
			if err == gcache.KeyNotFoundError {
				rsvp, err := c.client.GetEventRSVP(ctx, group, event)
				if err != nil {
					c.handleInternalError(w, "failed to fetch RSVP", err)
					return
				}
				cached = rsvp
				if rsvp.EventStatus == "past" {
					c.cache.SetWithExpire(cacheKey, rsvp, c.upcomingEventsCachingDuration)
				} else {
					c.cache.SetWithExpire(cacheKey, rsvp, c.pastEventsCachingDuration)
				}
			} else {
				c.handleInternalError(w, "failed to fetch RSVP", err)
				return
			}
		}
		rsvp, ok := cached.(*meetupcom.EventRSVP)
		if !ok {
			c.handleInternalError(w, "unexpected type found in cache", nil)
			return
		}
		s.AllYesGuests += rsvp.YesGuestCount
		for _, m := range rsvp.YesMembers {
			memberID := fmt.Sprintf("%d", m.ID)
			if _, found := knownProfiles[m.ID]; found {
				continue
			}
			s.AllYesIDs = append(s.AllYesIDs, memberID)
			s.Profiles[memberID] = Profile{
				Name:     m.Name,
				ID:       memberID,
				ThumbURL: m.Photo.ThumbLink,
			}
			knownProfiles[m.ID] = struct{}{}
		}
		result[meetup] = &EventRSVP{RSVP: rsvp}
	}
	response.Meetups = result
	response.Summary = s
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
