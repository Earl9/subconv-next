package static

import "embed"

//go:embed index.html app.js style.css favicon.svg
var Assets embed.FS
