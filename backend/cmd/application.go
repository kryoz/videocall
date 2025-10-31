package cmd

import (
	"github.com/spf13/cobra"

	"videocall/app"
	"videocall/app/config"
)

var runApplicationCmd = &cobra.Command{
	Use:   "application",
	Short: "run application",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewFromEnv()
		if err != nil {
			panic(err)
		}

		if err := app.Run(cmd.Context(), cfg); err != nil {
			panic(err)
		}

	},
}
