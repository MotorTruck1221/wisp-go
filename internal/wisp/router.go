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

func serveStatic(dir string, staticDir string, mux *http.ServeMux) {
    folder := findFolder(staticDir)
    mux.HandleFunc(dir, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, folder+r.URL.Path)
	})
    fmt.Println("Serving static files from " + folder + " at " + dir)
}

func InternalRouter(host string, port string, wispDir string, staticDir string, dir string) {
	mux := http.NewServeMux()
	mux.HandleFunc(wispDir, wisp)
	fmt.Println("Listening on http://" + host + ":" + port + dir)
    if staticDir != "n/a/" {
        serveStatic(dir, staticDir, mux)
    } else {
       mux.HandleFunc(dir, func(w http.ResponseWriter, r *http.Request) { 
           w.Write([]byte("Wisp server running on directory " + wispDir))
         })
    }
	if host == "0.0.0.0" {
		fmt.Println("Also listening on http://localhost:" + port + dir)
	}
	fmt.Println("Wisp available at " + wispDir)
	http.ListenAndServe(host+":"+port, mux)
}
