package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bluele/gcache"
	"github.com/pressly/chi"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"github.com/zerok/rsvps/internal/meetupcom"
	"github.com/zerok/rsvps/internal/whitelist"
)

func main() {
	var verbose bool
	var apiKey string
	var httpAddr string
	var listingURLs []string
	var allowedOrigins []string
	var pastEventsCachingDuration time.Duration
	var upcomingEventsCachingDuration time.Duration
	var whitelistUpdateInterval time.Duration

	log := logrus.New()

	pflag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	pflag.StringVar(&apiKey, "meetup-api-key", "", "API key for meetup.com")
	pflag.StringSliceVar(&listingURLs, "whitelist-url", []string{}, "URL to a meetup whitelist. Can be used multiple times.")
	pflag.DurationVar(&whitelistUpdateInterval, "whitelist-update-interval", time.Minute*10, "Interval for updating the whitelist from the provided URLs")
	pflag.StringVar(&httpAddr, "http-addr", "127.0.0.1:8080", "Address the HTTP server should listen on")
	pflag.StringSliceVar(&allowedOrigins, "allowed-origin", []string{}, "Allowed origin host for XHRs")
	pflag.DurationVar(&upcomingEventsCachingDuration, "cache-duration-upcoming", time.Minute*10, "Cache duration for upcoming events")
	pflag.DurationVar(&pastEventsCachingDuration, "cache-duration-past", time.Minute*120, "Cache duration for past events")
	pflag.Parse()

	if verbose {
		log.Level = logrus.DebugLevel
	}

	cache := gcache.New(1000).LRU().Build()
	ctx := context.Background()

	client, err := meetupcom.NewClient(func(c *meetupcom.Client) error {
		c.APIKey = apiKey
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to create meetup.com client")
	}

	wl, err := whitelist.New(func(w *whitelist.Whitelist) error {
		w.URLs = listingURLs
		w.UpdateInterval = whitelistUpdateInterval
		return nil
	})
	if err != nil {
		log.WithError(err).Fatalf("Failed to setup whitelists")
	}
	wl.StartAutoUpdate(ctx)

	ctrl, err := newController(func(c *controller) error {
		c.client = client
		c.cache = cache
		c.log = log
		c.whitelist = wl
		c.pastEventsCachingDuration = pastEventsCachingDuration
		c.upcomingEventsCachingDuration = upcomingEventsCachingDuration
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to setup controller")
	}

	router := chi.NewRouter()
	router.Post("/query", ctrl.handleGetRSVPs)

	log.Infof("Starting HTTP server on %s", httpAddr)
	srv := http.Server{}
	srv.Handler = cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
	}).Handler(router)
	srv.Addr = httpAddr
	if err := srv.ListenAndServe(); err != nil {
		log.WithError(err).Fatalf("Failed to start HTTP server")
	}
}
