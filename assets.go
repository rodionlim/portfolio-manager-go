package root

import "embed"

//go:embed config.yaml seed/refdata.yaml
var EmbeddedFiles embed.FS
