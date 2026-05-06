package main

import (
	squadron "github.com/mlund01/squadron-sdk"
)

var tools = map[string]*squadron.ToolInfo{
	// ─── Containers ─────────────────────────────────────────────
	"list_containers": {
		Name:        "list_containers",
		Description: "List Docker containers. By default returns only running containers.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"all": {
					Type:        squadron.TypeBoolean,
					Description: "Include stopped containers (docker ps -a). Default: false",
				},
				"filter": {
					Type:        squadron.TypeString,
					Description: "Optional docker filter expression, e.g. \"name=web\" or \"status=exited\"",
				},
			},
		},
	},
	"inspect_container": {
		Name:        "inspect_container",
		Description: "Return the full JSON inspect output for a container.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {
					Type:        squadron.TypeString,
					Description: "Container name or ID",
				},
			},
			Required: []string{"container"},
		},
	},
	"start_container": {
		Name:        "start_container",
		Description: "Start a stopped container.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {Type: squadron.TypeString, Description: "Container name or ID"},
			},
			Required: []string{"container"},
		},
	},
	"stop_container": {
		Name:        "stop_container",
		Description: "Stop a running container, optionally with a graceful-shutdown timeout.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {Type: squadron.TypeString, Description: "Container name or ID"},
				"timeout": {
					Type:        squadron.TypeInteger,
					Description: "Seconds to wait for graceful stop before SIGKILL. Default: 10",
				},
			},
			Required: []string{"container"},
		},
	},
	"restart_container": {
		Name:        "restart_container",
		Description: "Restart a container.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {Type: squadron.TypeString, Description: "Container name or ID"},
				"timeout": {
					Type:        squadron.TypeInteger,
					Description: "Seconds to wait for graceful stop before restart. Default: 10",
				},
			},
			Required: []string{"container"},
		},
	},
	"remove_container": {
		Name:        "remove_container",
		Description: "Remove a container. Use force=true to remove a running container.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {Type: squadron.TypeString, Description: "Container name or ID"},
				"force":     {Type: squadron.TypeBoolean, Description: "Force remove even if running"},
				"volumes":   {Type: squadron.TypeBoolean, Description: "Also remove anonymous volumes"},
			},
			Required: []string{"container"},
		},
	},
	"logs": {
		Name:        "logs",
		Description: "Fetch logs from a container.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {Type: squadron.TypeString, Description: "Container name or ID"},
				"tail": {
					Type:        squadron.TypeInteger,
					Description: "Number of trailing lines to return. 0 means all. Default: 200",
				},
				"since": {
					Type:        squadron.TypeString,
					Description: "Only return logs after this timestamp/duration (e.g. \"10m\", \"2025-01-01T00:00:00Z\")",
				},
				"timestamps": {
					Type:        squadron.TypeBoolean,
					Description: "Prefix each line with a timestamp",
				},
			},
			Required: []string{"container"},
		},
	},
	"run_container": {
		Name:        "run_container",
		Description: "Run a new container from an image. Returns the container ID (detached mode) or command output.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"image":   {Type: squadron.TypeString, Description: "Image to run (e.g. \"nginx:alpine\")"},
				"name":    {Type: squadron.TypeString, Description: "Optional container name"},
				"command": {Type: squadron.TypeString, Description: "Optional command to run inside the container (overrides CMD)"},
				"detach": {
					Type:        squadron.TypeBoolean,
					Description: "Run detached (-d). Default: true",
				},
				"remove": {
					Type:        squadron.TypeBoolean,
					Description: "Remove container when it exits (--rm). Cannot be combined with detach=true unless intended.",
				},
				"env": {
					Type:        squadron.TypeArray,
					Description: "Environment variables in KEY=VALUE form",
					Items:       &squadron.Property{Type: squadron.TypeString},
				},
				"ports": {
					Type:        squadron.TypeArray,
					Description: "Port mappings, e.g. [\"8080:80\", \"443:443\"]",
					Items:       &squadron.Property{Type: squadron.TypeString},
				},
				"volumes": {
					Type:        squadron.TypeArray,
					Description: "Volume mappings, e.g. [\"/host/path:/container/path\", \"named-vol:/data\"]",
					Items:       &squadron.Property{Type: squadron.TypeString},
				},
				"network": {Type: squadron.TypeString, Description: "Network to attach the container to"},
			},
			Required: []string{"image"},
		},
	},
	"exec": {
		Name:        "exec",
		Description: "Run a one-shot command inside a running container and return its output.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"container": {Type: squadron.TypeString, Description: "Container name or ID"},
				"command":   {Type: squadron.TypeString, Description: "Shell command to execute"},
				"shell": {
					Type:        squadron.TypeString,
					Description: "Shell to run the command with. Default: \"sh\"",
				},
				"timeout": {
					Type:        squadron.TypeInteger,
					Description: "Timeout in seconds. 0 means no timeout. Default: 60",
				},
			},
			Required: []string{"container", "command"},
		},
	},

	// ─── Images ─────────────────────────────────────────────────
	"list_images": {
		Name:        "list_images",
		Description: "List Docker images on the host.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"filter": {
					Type:        squadron.TypeString,
					Description: "Optional docker filter expression, e.g. \"reference=nginx*\"",
				},
				"dangling_only": {
					Type:        squadron.TypeBoolean,
					Description: "Only show dangling images",
				},
			},
		},
	},
	"pull_image": {
		Name:        "pull_image",
		Description: "Pull an image from a registry.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"image": {Type: squadron.TypeString, Description: "Image reference, e.g. \"nginx:alpine\""},
			},
			Required: []string{"image"},
		},
	},
	"remove_image": {
		Name:        "remove_image",
		Description: "Remove a local image.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"image": {Type: squadron.TypeString, Description: "Image reference or ID"},
				"force": {Type: squadron.TypeBoolean, Description: "Force removal even if in use"},
			},
			Required: []string{"image"},
		},
	},
	"tag_image": {
		Name:        "tag_image",
		Description: "Add a new tag pointing to an existing image.",
		Schema: squadron.Schema{
			Type: squadron.TypeObject,
			Properties: squadron.PropertyMap{
				"source": {Type: squadron.TypeString, Description: "Existing image reference or ID"},
				"target": {Type: squadron.TypeString, Description: "New tag, e.g. \"myrepo/app:v2\""},
			},
			Required: []string{"source", "target"},
		},
	},

	// ─── Networks & Volumes ─────────────────────────────────────
	"list_networks": {
		Name:        "list_networks",
		Description: "List Docker networks.",
		Schema:      squadron.Schema{Type: squadron.TypeObject, Properties: squadron.PropertyMap{}},
	},
	"list_volumes": {
		Name:        "list_volumes",
		Description: "List Docker volumes.",
		Schema:      squadron.Schema{Type: squadron.TypeObject, Properties: squadron.PropertyMap{}},
	},

	// ─── System ─────────────────────────────────────────────────
	"version": {
		Name:        "version",
		Description: "Show docker client and server version information.",
		Schema:      squadron.Schema{Type: squadron.TypeObject, Properties: squadron.PropertyMap{}},
	},
	"info": {
		Name:        "info",
		Description: "Show system-wide Docker daemon information.",
		Schema:      squadron.Schema{Type: squadron.TypeObject, Properties: squadron.PropertyMap{}},
	},
}
