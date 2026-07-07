package probes

import (
	"context"
	"sync"
	"time"
)

type Level string

const (
	LevelOK      Level = "ok"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
	LevelMuted   Level = "muted"
)

type Status struct {
	Name      string
	Value     string
	Detail    string
	Level     Level
	CheckedAt time.Time
}

type Probe interface {
	Name() string
	Check(context.Context) Status
}

func Default() []Probe {
	return []Probe{
		DockerProbe{},
		NewCommandProbe("Node", []CommandSpec{{Binary: "node", Args: []string{"--version"}}}),
		NewCommandProbe("Python", []CommandSpec{
			{Binary: "python3", Args: []string{"--version"}},
			{Binary: "python", Args: []string{"--version"}},
			{Binary: "py", Args: []string{"-3", "--version"}},
		}),
		NewCommandProbe("Go", []CommandSpec{{Binary: "go", Args: []string{"version"}}}),
		NewCommandProbe("Forge", []CommandSpec{{Binary: "forge", Args: []string{"--version"}}}),
	}
}

func CheckAll(ctx context.Context, checks []Probe, timeout time.Duration) []Status {
	statuses := make([]Status, len(checks))

	var wg sync.WaitGroup
	for index, check := range checks {
		wg.Add(1)
		go func(index int, check Probe) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			status := check.Check(checkCtx)
			if status.CheckedAt.IsZero() {
				status.CheckedAt = time.Now()
			}
			statuses[index] = status
		}(index, check)
	}

	wg.Wait()
	return statuses
}
