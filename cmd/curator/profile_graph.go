package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/profilegraph"
	"github.com/relux-works/curator/internal/version"
)

func (c cli) cmdProfileGraph(args []string) int {
	flags := c.newFlagSet("profile graph")
	flags.SetOutput(io.Discard)
	jsonMode := flags.Bool("json", false, "emit curator-profile-graph/1 JSON")
	source := flags.String("source", "", "preview a local source tree offline")
	environment := flags.String("env", "", "select one registered environment")
	positional, err := parseInterspersed(flags, args)
	if errors.Is(err, flag.ErrHelp) {
		_, _ = fmt.Fprintln(c.stdout, "usage: curator profile graph <profile> [--json] [--source <dir>] [--env <id>]")
		return exitOK
	}
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile graph <profile> [--json] [--source <dir>] [--env <id>]")
		return exitUsage
	}
	fail := func(err error) int { _, _ = fmt.Fprintln(c.stderr, "curator:", err); return exitFail }
	if *environment != "" {
		if _, err := envregistry.ByID(envregistry.NormalizeEnvID(*environment)); err != nil {
			return fail(&profilegraph.Error{Class: "graph_invalid_input"})
		}
	}
	reader := profilegraph.NewReader(profilegraph.V1Limits)
	req := profilegraph.Request{Profile: positional[0], Source: *source, Environment: *environment, Version: version.String(), Reader: reader, Home: filepath.Dir(c.config.Path())}
	configPath := c.config.Path()
	projection, err := profilegraph.ReadProjectionWithSystem(reader, configPath, config.SystemCandidatePath(), req.Profile, req.Source != "")
	if err != nil {
		return fail(err)
	}
	req.Projection = projection

	graph, err := profilegraph.Build(req)
	if err != nil {
		return fail(err)
	}
	var output []byte
	if *jsonMode {
		output, err = profilegraph.JSON(graph, reader.Limits.Output)
	} else {
		output, err = profilegraph.Text(graph, reader.Limits.Output)
	}
	if err != nil {
		return fail(err)
	}
	for _, d := range graph.Diagnostics {
		_, _ = fmt.Fprintln(c.stderr, d.Class+": "+d.Message)
	}
	if _, err := c.stdout.Write(output); err != nil {
		return exitFail
	}
	return exitOK
}
