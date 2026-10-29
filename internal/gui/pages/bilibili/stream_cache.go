package bilibili

import (
	"net/url"
	"strconv"
	"time"
)

// Honor signed stream deadlines, leaving a short request margin. Without an
// expiry field, use a bounded fallback rather than caching URLs indefinitely.
const streamCacheTTL = 10 * time.Minute

func streamValidity(resolvedAt time.Time, streams ...string) time.Time {
	if resolvedAt.IsZero() {
		return time.Time{}
	}
	var until time.Time
	for _, raw := range streams {
		if raw == "" {
			continue
		}
		parsed, err := url.Parse(raw)
		if err != nil {
			return time.Time{}
		}
		var signedDeadline time.Time
		for _, key := range []string{"deadline", "expires", "expire", "wsTime"} {
			value := parsed.Query().Get(key)
			if value == "" {
				continue
			}
			base := 10
			if key == "wsTime" {
				base = 16
			}
			stamp, err := strconv.ParseInt(value, base, 64)
			if err != nil || stamp <= 0 {
				continue
			}
			deadline := time.Unix(stamp, 0).Add(-30 * time.Second)
			if signedDeadline.IsZero() || deadline.Before(signedDeadline) {
				signedDeadline = deadline
			}
		}
		if signedDeadline.IsZero() {
			signedDeadline = resolvedAt.Add(streamCacheTTL)
		}
		if until.IsZero() || signedDeadline.Before(until) {
			until = signedDeadline
		}
	}
	return until
}
