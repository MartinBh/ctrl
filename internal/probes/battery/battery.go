package battery

import (
	"context"
	"errors"
	"fmt"
	"math"

	batterylib "github.com/distatus/battery"

	"github.com/martinbhatta/ctrl/internal/probes"
)

type Status struct {
	Present bool
	Percent float64
	State   string
	Detail  string
	Level   probes.Level
	Err     error
}

func Check(ctx context.Context) Status {
	if err := ctx.Err(); err != nil {
		return errorStatus(err)
	}

	batteries, err := batterylib.GetAll()
	if errors.Is(err, batterylib.ErrNotFound) {
		return noBatteryStatus()
	}
	if err != nil {
		return errorStatus(err)
	}

	return Summarize(batteries)
}

func Summarize(batteries []*batterylib.Battery) Status {
	if len(batteries) == 0 {
		return noBatteryStatus()
	}

	var current float64
	var capacity float64
	stateCounts := map[batterylib.AgnosticState]int{}

	for _, battery := range batteries {
		if battery == nil {
			continue
		}

		current += math.Max(0, battery.Current)
		capacity += bestCapacity(*battery)
		stateCounts[battery.State.Raw]++
	}

	if len(stateCounts) == 0 {
		return noBatteryStatus()
	}

	state := dominantState(stateCounts)
	percent := percent(current, capacity)

	return Status{
		Present: true,
		Percent: percent,
		State:   displayState(state),
		Detail:  detailForCount(len(batteries)),
		Level:   levelForState(state, percent),
	}
}

func bestCapacity(battery batterylib.Battery) float64 {
	if battery.Full > 0 {
		return battery.Full
	}
	if battery.Design > 0 {
		return battery.Design
	}
	return 0
}

func percent(current float64, capacity float64) float64 {
	if capacity <= 0 {
		return 0
	}

	return math.Max(0, math.Min(100, (current/capacity)*100))
}

func dominantState(stateCounts map[batterylib.AgnosticState]int) batterylib.AgnosticState {
	preferred := []batterylib.AgnosticState{
		batterylib.Charging,
		batterylib.Discharging,
		batterylib.Full,
		batterylib.Empty,
		batterylib.Idle,
		batterylib.Unknown,
		batterylib.Undefined,
	}

	for _, state := range preferred {
		if stateCounts[state] > 0 {
			return state
		}
	}

	return batterylib.Unknown
}

func displayState(state batterylib.AgnosticState) string {
	switch state {
	case batterylib.Charging:
		return "charging"
	case batterylib.Discharging:
		return "discharging"
	case batterylib.Full:
		return "full"
	case batterylib.Empty:
		return "empty"
	case batterylib.Idle:
		return "idle"
	default:
		return "unknown"
	}
}

func levelForState(state batterylib.AgnosticState, pct float64) probes.Level {
	switch state {
	case batterylib.Discharging, batterylib.Empty:
		if pct < 15 {
			return probes.LevelError
		}
		if pct <= 30 {
			return probes.LevelWarning
		}
		return probes.LevelOK
	case batterylib.Unknown, batterylib.Undefined:
		return probes.LevelWarning
	default:
		return probes.LevelOK
	}
}

func detailForCount(count int) string {
	if count == 1 {
		return "1 battery detected"
	}

	return fmt.Sprintf("%d batteries detected", count)
}

func noBatteryStatus() Status {
	return Status{
		Present: false,
		State:   "no battery detected",
		Detail:  "desktop or external power only",
		Level:   probes.LevelMuted,
	}
}

func errorStatus(err error) Status {
	return Status{
		Present: false,
		State:   "unavailable",
		Detail:  err.Error(),
		Level:   probes.LevelError,
		Err:     err,
	}
}
