package webui

import "embed"

// Assets contains the production frontend bundle.
//
//go:embed dist/*
var Assets embed.FS
