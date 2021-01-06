package cmd

import (
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cmd/serve"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cmd/version"
)

type Options struct {
	Serve   serve.Command   `command:"serve" alias:"s" description:"Start web-server"`
	Version version.Command `command:"version" alias:"v" description:"Display application version"`
}
