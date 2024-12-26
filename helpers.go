package stresser

import (
	"strconv"
	"time"

	"github.com/xlab/pace"
	log "github.com/xlab/suplog"
)

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func bytesOrPanic(out []byte, err error) []byte {
	if err != nil {
		panic(err)
	}

	return out
}

func NewPaceReporter(logger log.Logger) pace.ReporterFunc {
	var previous float64
	var stalled time.Time

	return func(label string, timeframe time.Duration, value float64) {
		switch {
		case value == 0 && previous == 0:
			return // don't report anything
		case value == 0 && previous != 0:
			dur := timeframe
			if !stalled.IsZero() {
				dur = time.Since(stalled)
				n := dur / timeframe
				if dur-n*timeframe < 10*time.Millisecond {
					dur = n * timeframe
				}
			} else {
				stalled = time.Now().Add(-dur)
			}
			logger.Warningf("%s: stalled for %v", label, dur)
			return
		default:
			previous = value
			stalled = time.Time{}
		}
		floatFmt := func(f float64) string {
			return strconv.FormatFloat(f, 'f', 3, 64)
		}
		switch timeframe {
		case time.Second:
			logger.Infof("%s: %s/s in %v", label, floatFmt(value), timeframe)
		case time.Minute:
			logger.Infof("%s: %s/m in %v", label, floatFmt(value), timeframe)
		case time.Hour:
			logger.Infof("%s: %s/h in %v", label, floatFmt(value), timeframe)
		case 24 * time.Hour:
			logger.Infof("%s: %s/day in %v", label, floatFmt(value), timeframe)
		default:
			logger.Infof("%s %s in %v (pace: %s/s)", floatFmt(value), label, timeframe,
				floatFmt(value/(float64(timeframe)/float64(time.Second))),
			)
		}
	}
}
