package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	squadron "github.com/mlund01/squadron-sdk"
)

type DockerPlugin struct {
	mu     sync.Mutex
	runner *dockerRunner
}

// Configure accepts optional settings:
//
//	host    - DOCKER_HOST value (e.g. "tcp://remote:2375", "ssh://user@host")
//	context - Docker CLI context name to use (--context)
func (p *DockerPlugin) Configure(settings map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runner = &dockerRunner{
		host:    settings["host"],
		context: settings["context"],
	}
	return nil
}

func (p *DockerPlugin) Call(toolName string, payload string) (string, error) {
	p.mu.Lock()
	r := p.runner
	p.mu.Unlock()
	if r == nil {
		r = &dockerRunner{}
	}

	var raw map[string]any
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &raw); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
	}

	switch toolName {
	case "list_containers":
		args := []string{"ps", "--format", "{{json .}}"}
		if getBool(raw, "all") {
			args = append(args, "-a")
		}
		if f := getString(raw, "filter"); f != "" {
			args = append(args, "--filter", f)
		}
		return r.run(30, args...)

	case "inspect_container":
		c := getString(raw, "container")
		if c == "" {
			return "", fmt.Errorf("container is required")
		}
		return r.run(30, "inspect", c)

	case "start_container":
		c := getString(raw, "container")
		if c == "" {
			return "", fmt.Errorf("container is required")
		}
		return r.run(60, "start", c)

	case "stop_container":
		c := getString(raw, "container")
		if c == "" {
			return "", fmt.Errorf("container is required")
		}
		args := []string{"stop"}
		if t := getInt(raw, "timeout"); t > 0 {
			args = append(args, "-t", strconv.Itoa(t))
		}
		args = append(args, c)
		return r.run(0, args...)

	case "restart_container":
		c := getString(raw, "container")
		if c == "" {
			return "", fmt.Errorf("container is required")
		}
		args := []string{"restart"}
		if t := getInt(raw, "timeout"); t > 0 {
			args = append(args, "-t", strconv.Itoa(t))
		}
		args = append(args, c)
		return r.run(0, args...)

	case "remove_container":
		c := getString(raw, "container")
		if c == "" {
			return "", fmt.Errorf("container is required")
		}
		args := []string{"rm"}
		if getBool(raw, "force") {
			args = append(args, "-f")
		}
		if getBool(raw, "volumes") {
			args = append(args, "-v")
		}
		args = append(args, c)
		return r.run(60, args...)

	case "logs":
		c := getString(raw, "container")
		if c == "" {
			return "", fmt.Errorf("container is required")
		}
		tail := getInt(raw, "tail")
		if _, ok := raw["tail"]; !ok {
			tail = 200
		}
		args := []string{"logs"}
		if tail > 0 {
			args = append(args, "--tail", strconv.Itoa(tail))
		}
		if since := getString(raw, "since"); since != "" {
			args = append(args, "--since", since)
		}
		if getBool(raw, "timestamps") {
			args = append(args, "-t")
		}
		args = append(args, c)
		return r.run(60, args...)

	case "run_container":
		image := getString(raw, "image")
		if image == "" {
			return "", fmt.Errorf("image is required")
		}
		// Default detach = true
		detach := true
		if v, ok := raw["detach"]; ok {
			if b, ok := v.(bool); ok {
				detach = b
			}
		}
		args := []string{"run"}
		if detach {
			args = append(args, "-d")
		}
		if getBool(raw, "remove") {
			args = append(args, "--rm")
		}
		if name := getString(raw, "name"); name != "" {
			args = append(args, "--name", name)
		}
		if net := getString(raw, "network"); net != "" {
			args = append(args, "--network", net)
		}
		for _, e := range getStringArray(raw, "env") {
			args = append(args, "-e", e)
		}
		for _, p := range getStringArray(raw, "ports") {
			args = append(args, "-p", p)
		}
		for _, v := range getStringArray(raw, "volumes") {
			args = append(args, "-v", v)
		}
		args = append(args, image)
		if cmd := getString(raw, "command"); cmd != "" {
			// Pass via shell so users can write a normal command line
			args = append(args, "sh", "-c", cmd)
		}
		return r.run(0, args...)

	case "exec":
		c := getString(raw, "container")
		cmd := getString(raw, "command")
		if c == "" || cmd == "" {
			return "", fmt.Errorf("container and command are required")
		}
		shell := getString(raw, "shell")
		if shell == "" {
			shell = "sh"
		}
		timeout := getInt(raw, "timeout")
		if _, ok := raw["timeout"]; !ok {
			timeout = 60
		}
		return r.run(timeout, "exec", c, shell, "-c", cmd)

	case "list_images":
		args := []string{"images", "--format", "{{json .}}"}
		if f := getString(raw, "filter"); f != "" {
			args = append(args, "--filter", f)
		}
		if getBool(raw, "dangling_only") {
			args = append(args, "--filter", "dangling=true")
		}
		return r.run(30, args...)

	case "pull_image":
		image := getString(raw, "image")
		if image == "" {
			return "", fmt.Errorf("image is required")
		}
		return r.run(0, "pull", image)

	case "remove_image":
		image := getString(raw, "image")
		if image == "" {
			return "", fmt.Errorf("image is required")
		}
		args := []string{"rmi"}
		if getBool(raw, "force") {
			args = append(args, "-f")
		}
		args = append(args, image)
		return r.run(60, args...)

	case "tag_image":
		src := getString(raw, "source")
		dst := getString(raw, "target")
		if src == "" || dst == "" {
			return "", fmt.Errorf("source and target are required")
		}
		return r.run(30, "tag", src, dst)

	case "list_networks":
		return r.run(30, "network", "ls", "--format", "{{json .}}")

	case "list_volumes":
		return r.run(30, "volume", "ls", "--format", "{{json .}}")

	case "version":
		return r.run(15, "version", "--format", "{{json .}}")

	case "info":
		return r.run(15, "info", "--format", "{{json .}}")

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (p *DockerPlugin) GetToolInfo(toolName string) (*squadron.ToolInfo, error) {
	info, ok := tools[toolName]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
	return info, nil
}

func (p *DockerPlugin) ListTools() ([]*squadron.ToolInfo, error) {
	result := make([]*squadron.ToolInfo, 0, len(tools))
	for _, info := range tools {
		result = append(result, info)
	}
	return result, nil
}

// ─── payload helpers ──────────────────────────────────────────

func getString(m map[string]any, k string) string {
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]any, k string) bool {
	if v, ok := m[k]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getInt(m map[string]any, k string) int {
	if v, ok := m[k]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}

func getStringArray(m map[string]any, k string) []string {
	v, ok := m[k]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	squadron.Serve(&DockerPlugin{})
}
