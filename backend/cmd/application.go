package cmd

import (
	"videocall/internal"
	"videocall/internal/infrastructure/config"

	"github.com/spf13/cobra"
)

var runApplicationCmd = &cobra.Command{
	Use:   "application",
	Short: "run application",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewFromEnv()
		if err != nil {
			panic(err)
		}

		if err := internal.Run(cmd.Context(), cfg); err != nil {
			panic(err)
		}

	},
}
