package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/martinbhatta/ctrl/internal/app"
	"github.com/martinbhatta/ctrl/internal/store"
)

func Execute(version string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root := NewRootCommand(version)
	root.SetContext(ctx)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func NewRootCommand(version string) *cobra.Command {
	var refreshEvery time.Duration

	cmd := &cobra.Command{
		Use:           "ctrl",
		Short:         "A personal retro terminal command center.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			todoPath, err := store.DefaultTodosPath()
			if err != nil {
				return err
			}

			dashboard := app.New(app.Options{
				Version:      version,
				TodoPath:     todoPath,
				RefreshEvery: refreshEvery,
			})

			return dashboard.Run(cmd.Context())
		},
	}

	cmd.Flags().DurationVar(&refreshEvery, "refresh", 5*time.Minute, "environment refresh interval")
	cmd.AddCommand(versionCommand(version))

	return cmd
}

func versionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ctrl version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}
