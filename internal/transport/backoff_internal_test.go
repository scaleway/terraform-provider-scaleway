package transport

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		header  string
		wantOK  bool
		want    time.Duration
		wantMin time.Duration // for date-based values, a lower bound
	}{
		{name: "empty", header: "", wantOK: false},
		{name: "seconds", header: "30", wantOK: true, want: 30 * time.Second},
		{name: "zero seconds", header: "0", wantOK: true, want: 0},
		{name: "negative rejected", header: "-5", wantOK: false},
		{name: "malformed", header: "soon", wantOK: false},
		{
			name:    "http date in the future",
			header:  time.Now().Add(2 * time.Minute).UTC().Format(http.TimeFormat),
			wantOK:  true,
			wantMin: time.Minute, // roughly two minutes, allow slack
		},
		{
			name:   "http date in the past",
			header: time.Now().Add(-2 * time.Minute).UTC().Format(http.TimeFormat),
			wantOK: true,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseRetryAfter(tt.header)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}

			if !ok {
				return
			}

			if tt.wantMin > 0 {
				if got < tt.wantMin {
					t.Fatalf("got %v, want >= %v", got, tt.wantMin)
				}

				return
			}

			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimitBackoff(t *testing.T) {
	t.Parallel()

	minWait := 2 * time.Second
	maxWait := 2 * time.Minute

	t.Run("honors Retry-After over computed backoff", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{Header: http.Header{"Retry-After": {"30"}}}

		got := rateLimitBackoff(minWait, maxWait, 0, resp)
		if got != 30*time.Second {
			t.Fatalf("got %v, want 30s", got)
		}
	})

	t.Run("jittered exponential when no timing header present", func(t *testing.T) {
		t.Parallel()

		// Scaleway's gateway 429 carries no Retry-After: must fall back to backoff.
		resp := &http.Response{Header: http.Header{"X-Envoy-Ratelimited": {"true"}}}

		got := rateLimitBackoff(minWait, maxWait, 0, resp)
		if got < minWait || got > 2*minWait {
			t.Fatalf("got %v, want within [%v, %v]", got, minWait, 2*minWait)
		}
	})

	t.Run("jittered exponential stays within bounds and varies", func(t *testing.T) {
		t.Parallel()

		seen := map[time.Duration]struct{}{}

		for range 50 {
			got := rateLimitBackoff(minWait, maxWait, 4, nil)
			// attempt 4: exponential ceiling is min*2^4 = 32s, capped at maxWait.
			if got < minWait || got > 32*time.Second {
				t.Fatalf("got %v, want within [%v, 32s]", got, minWait)
			}

			seen[got] = struct{}{}
		}

		if len(seen) < 2 {
			t.Fatalf("expected jitter to produce varied delays, got %d distinct values", len(seen))
		}
	})

	t.Run("caps at maxWait", func(t *testing.T) {
		t.Parallel()

		got := rateLimitBackoff(minWait, maxWait, 20, nil)
		if got > maxWait {
			t.Fatalf("got %v, want <= %v", got, maxWait)
		}
	})

	t.Run("zero min does not panic", func(t *testing.T) {
		t.Parallel()

		got := rateLimitBackoff(0, 0, 3, nil)
		if got != 0 {
			t.Fatalf("got %v, want 0", got)
		}
	})
}
