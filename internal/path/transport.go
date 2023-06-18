package path

import (
	"net/http"
	"sync"
	"time"
)

type reqTime struct {
	Duration time.Duration
	URL      string
}

var (
	timingMutex = sync.Mutex{}
	ReqTiming   = make([]reqTime, 0, 1000)
)

type Timer struct {
	transport http.RoundTripper
}

func (t *Timer) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport
	if t.transport != nil {
		transport = t.transport
	}
	now := time.Now()
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	duration := time.Since(now)
	timingMutex.Lock()
	ReqTiming = append(ReqTiming, reqTime{Duration: duration, URL: req.URL.String()})
	timingMutex.Unlock()
	return resp, nil
}
