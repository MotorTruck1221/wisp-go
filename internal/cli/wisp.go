package cli

import (
	"github.com/motortruck1221/wisp-go/internal/wisp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
	startCmd.Flags().StringP("dir", "d", "/", "Directory to listen on")
	startCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to listen on")
	startCmd.Flags().StringP("static", "s", "n/a", "Directory to serve static files from")
	startCmd.Flags().StringP("wisp", "w", "wisp", "Directory to serve the wisp websocket from")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the wisp server",
	Long:  "Start the wisp server",
	Run: func(cmd *cobra.Command, args []string) {
		host := cmd.Flag("host").Value.String()
		port := cmd.Flag("port").Value.String()
		dir := cmd.Flag("dir").Value.String()
		static := cmd.Flag("static").Value.String()
		wispDir := cmd.Flag("wisp").Value.String()
		if dir[0] != '/' {
			dir = "/" + dir
		}
		if dir[len(dir)-1] != '/' {
			dir = dir + "/"
		}
		if static[len(static)-1] != '/' {
			static = static + "/"
		}
		if wispDir[0] != '/' {
			wispDir = "/" + wispDir
		}
		if wispDir[len(wispDir)-1] != '/' {
			wispDir = wispDir + "/"
		}
		wisp.InternalRouter(host, port, wispDir, static, dir)
	},
}
