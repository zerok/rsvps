package meetupcom

import (
	"fmt"
	"net/url"
	"strings"
)

func ParseURL(u string) (string, string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", "", err
	}
	if parsedURL.Host != "www.meetup.com" {
		return "", "", fmt.Errorf("host must be www.meetup.com")
	}
	pathSegments := strings.Split(parsedURL.Path, "/")
	if len(pathSegments) < 4 || pathSegments[2] != "events" {
		return "", "", fmt.Errorf("not an event URL")
	}
	group, err := url.PathUnescape(pathSegments[1])
	if err != nil {
		return "", "", fmt.Errorf("failed to unescape group name: %s", err.Error())
	}
	eventID, err := url.PathUnescape(pathSegments[3])
	if err != nil {
		return "", "", fmt.Errorf("failed to unescape event ID: %s", err.Error())
	}
	return group, eventID, nil
}
