package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "videocall",
	Short:   `Videocall backend server`,
	Long:    `Videocall backend server`,
	Version: "",
}

func Execute() {
	rootCmd.AddCommand(runApplicationCmd)

	if len(os.Args[1:]) == 0 {
		// run server by default
		os.Args = append(os.Args, runApplicationCmd.Use)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		cancel()
		fmt.Println(err)
		os.Exit(1)
	}
	cancel()
}
