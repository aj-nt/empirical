package empirical

import "embed"

// WebFiles embeds the web frontend build output for the HTTP server.
//go:embed web/dist web/dist/assets
var WebFiles embed.FS
