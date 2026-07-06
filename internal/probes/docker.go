package probes

import (
	"context"
	"os/exec"
	"strings"
)

type DockerProbe struct{}

func (DockerProbe) Name() string {
	return "Docker"
}

func (p DockerProbe) Check(ctx context.Context) Status {
	path, err := exec.LookPath("docker")
	if err != nil {
		return Status{Name: p.Name(), Value: "not found", Detail: "docker command not on PATH", Level: LevelWarning}
	}

	cmd := exec.CommandContext(ctx, path, "info", "--format", "{{.ServerVersion}}")
	output, err := cmd.CombinedOutput()
	value := firstLine(strings.TrimSpace(string(output)))

	if ctx.Err() != nil {
		return Status{Name: p.Name(), Value: "timeout", Detail: ctx.Err().Error(), Level: LevelError}
	}
	if err != nil {
		if value == "" {
			value = "daemon offline"
		}
		return Status{Name: p.Name(), Value: value, Detail: "docker info failed", Level: LevelError}
	}
	if value == "" {
		value = "daemon running"
	}

	return Status{Name: p.Name(), Value: "daemon running", Detail: "server " + value, Level: LevelOK}
}
