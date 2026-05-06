package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// dockerRunner runs the docker CLI with optional DOCKER_HOST / DOCKER_CONTEXT.
type dockerRunner struct {
	host    string
	context string
}

func (r *dockerRunner) run(timeoutSec int, args ...string) (string, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	if timeoutSec > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	full := []string{}
	if r.context != "" {
		full = append(full, "--context", r.context)
	}
	full = append(full, args...)

	cmd := exec.CommandContext(ctx, "docker", full...)
	if r.host != "" {
		cmd.Env = append(cmd.Environ(), "DOCKER_HOST="+r.host)
	}
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("docker command timed out after %d seconds", timeoutSec)
	}
	if err != nil {
		return "", fmt.Errorf("docker %v failed: %s", args, string(out))
	}
	return string(out), nil
}
