package wisp

import (
	"fmt"
	"net/http"
	"os"
)

func findFolder(dir string) string {
	//helper that returnn the full path of a folder location
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Directory " + dir + " does not exist")
		os.Exit(1)
	}
	return dir
}

func serveStatic(dir string, mux *http.ServeMux) {
    folder := findFolder(dir)
    mux.HandleFunc(dir, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, folder+r.URL.Path)
	})
    fmt.Println("Serving static files from " + folder + " at " + dir)
}

func InternalRouter(host string, port string, wipDir string, staticDir string, dir string) {
	mux := http.NewServeMux()
	mux.HandleFunc(wipDir, wisp)
	fmt.Println("Listening on http://" + host + ":" + port + dir)
    if staticDir != "n/a/" {
        serveStatic(staticDir, mux)
    }
	if host == "0.0.0.0" {
		fmt.Println("Also listening on http://localhost:" + port + dir)
	}
	fmt.Println("Wisp available at " + wipDir)
	http.ListenAndServe(host+":"+port, mux)
}
