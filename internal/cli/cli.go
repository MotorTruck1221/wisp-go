package cli

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "wisp-go",
	Short: "Wisp server written in Go",
	Long:  "Wisp server written in Go",
}

func Init() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
