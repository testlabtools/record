package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type setup struct {
	env   map[string]string
	debug bool
	log   *slog.Logger
}

func setupCommand(cmd *cobra.Command, args []string) setup {
	env := getEnv()
	if val := cmd.Context().Value("env"); val != nil {
		env = val.(map[string]string)
	}

	debug := cmd.Flag("debug").Value.String() == "true"
	if !debug {
		debug = env["TESTLAB_DEBUG"] != ""
	}

	if debug {
		setLogLevel(slog.LevelDebug)
	}

	l := slog.Default()

	var flags []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, fmt.Sprintf("%s=%s", flag.Name, flag.Value))
	})

	l.Info(fmt.Sprintf("start %s command", cmd.Use),
		"args", args,
		"flags", flags,
		"version", version,
		"commit", commit,
		"built", date,
	)

	return setup{
		env:   env,
		debug: debug,
		log:   l,
	}
}
