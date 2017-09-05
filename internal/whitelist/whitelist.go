package whitelist

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bluele/gcache"
)

type Whitelist struct {
	URLs           []string
	data           map[string]struct{}
	cache          gcache.Cache
	Log            *logrus.Logger
	lock           sync.RWMutex
	HTTPClient     *http.Client
	FetchTimeout   time.Duration
	UpdateInterval time.Duration
}

func (w *Whitelist) Contains(value string) bool {
	w.lock.RLock()
	defer w.lock.RUnlock()
	_, ok := w.data[value]
	return ok
}

func (w *Whitelist) Update(ctx context.Context) error {
	result := make(map[string]struct{})
	for _, u := range w.URLs {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled")
		default:
		}
		entries, err := w.fetchEntries(ctx, u)
		if err != nil {
			return fmt.Errorf("failed to fetch entries from %s: %s", u, err.Error())
		}
		for e := range entries {
			result[e] = struct{}{}
		}
	}
	w.lock.Lock()
	w.data = result
	defer w.lock.Unlock()
	return nil
}

func (w *Whitelist) fetchEntries(ctx context.Context, u string) (map[string]struct{}, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	timeoutCtx, cancel := context.WithTimeout(ctx, w.FetchTimeout)
	defer cancel()
	if err != nil {
		return nil, err
	}
	resp, err := w.HTTPClient.Do(req.WithContext(timeoutCtx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result := make(map[string]struct{})
	r := textproto.NewReader(bufio.NewReader(io.LimitReader(resp.Body, 1*1024*1024)))
	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result[trimmed] = struct{}{}
		}
	}
	return result, nil
}

func (w *Whitelist) StartAutoUpdate(ctx context.Context) {
	go func() {
		if err := w.Update(ctx); err != nil {
			w.Log.WithError(err).Error("Failed to update whitelist")
		}
		ticker := time.Tick(w.UpdateInterval)
		for {
			select {
			case <-ctx.Done():
				break
			case <-ticker:
				if err := w.Update(ctx); err != nil {
					w.Log.WithError(err).Error("Failed to update whitelist")
					continue
				}
			}
		}
	}()
}

func New(f func(*Whitelist) error) (*Whitelist, error) {
	w := &Whitelist{}
	if err := f(w); err != nil {
		return nil, err
	}
	if w.Log == nil {
		w.Log = logrus.New()
	}
	if w.URLs == nil || len(w.URLs) == 0 {
		return nil, fmt.Errorf("no URLs specified for whitelist")
	}
	if w.HTTPClient == nil {
		w.HTTPClient = &http.Client{}
	}
	if w.FetchTimeout == 0 {
		w.FetchTimeout = time.Second * 5
	}
	w.lock = sync.RWMutex{}
	w.data = make(map[string]struct{})
	return w, nil
}
