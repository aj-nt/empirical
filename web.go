package empirical

import "embed"

// WebFiles embeds the web frontend directory for the HTTP server.
//go:embed web/*
var WebFiles embed.FS
