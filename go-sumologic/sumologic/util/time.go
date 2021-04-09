package util

import (
	"github.com/relvacode/iso8601"
	"github.com/rs/zerolog/log"
	"time"
)

func FromISO8601String(timestamp string) time.Time {
	t, err := iso8601.ParseString(timestamp)
	if err != nil {
		log.Fatal().Err(err).Str("timestamp", timestamp).
			Msg("Error parsing ISO8601 timestamp")
	}
	return t
}

func UnixMillis(timestamp time.Time) int64 {
	return timestamp.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
