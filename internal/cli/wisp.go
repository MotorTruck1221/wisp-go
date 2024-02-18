package cli

import (
	"github.com/motortruck1221/wisp-go/internal/wisp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
	startCmd.Flags().StringP("dir", "d", "/", "Directory to serve from")
	startCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to listen on")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the wisp server",
	Long:  "Start the wisp server",
	Run: func(cmd *cobra.Command, args []string) {
		host := cmd.Flag("host").Value.String()
		port := cmd.Flag("port").Value.String()
		dir := cmd.Flag("dir").Value.String()
		if dir[0] != '/' {
			dir = "/" + dir
		}
		if dir[len(dir)-1] != '/' {
			dir = dir + "/"
		}
		wisp.InternalRouter(host, port, dir)
	},
}
