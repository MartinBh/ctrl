package probes

import (
	"context"
	"os/exec"
	"strings"
)

type CommandSpec struct {
	Binary string
	Args   []string
}

type CommandProbe struct {
	displayName string
	candidates  []CommandSpec
}

func NewCommandProbe(displayName string, candidates []CommandSpec) CommandProbe {
	return CommandProbe{
		displayName: displayName,
		candidates:  candidates,
	}
}

func (p CommandProbe) Name() string {
	return p.displayName
}

func (p CommandProbe) Check(ctx context.Context) Status {
	for _, candidate := range p.candidates {
		path, err := exec.LookPath(candidate.Binary)
		if err != nil {
			continue
		}

		cmd := exec.CommandContext(ctx, path, candidate.Args...)
		output, err := cmd.CombinedOutput()
		value := firstLine(strings.TrimSpace(string(output)))

		if ctx.Err() != nil {
			return Status{Name: p.Name(), Value: "timeout", Detail: ctx.Err().Error(), Level: LevelError}
		}
		if err != nil {
			if value == "" {
				value = "command failed"
			}
			return Status{Name: p.Name(), Value: value, Detail: err.Error(), Level: LevelError}
		}
		if value == "" {
			value = "available"
		}

		return Status{Name: p.Name(), Value: value, Detail: candidate.Binary, Level: LevelOK}
	}

	return Status{Name: p.Name(), Value: "not found", Detail: "command not on PATH", Level: LevelWarning}
}

func firstLine(value string) string {
	if value == "" {
		return ""
	}

	lines := strings.Split(value, "\n")
	return strings.TrimSpace(lines[0])
}
