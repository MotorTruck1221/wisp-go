package wisp

import (
	"fmt"
	"net/http"
	"os"
)

func InternalRouter(host string, port string, dir string) {
	mux := http.NewServeMux()
	mux.HandleFunc(dir, wisp)
	fmt.Println("Listening on http://" + host + ":" + port + dir)
	if host == "0.0.0.0" {
		fmt.Println("Also listening on http://localhost:" + port + dir)
	}
	http.ListenAndServe(host+":"+port, mux)
}

func findFolder(dir string) string {
	//helper that returnn the full path of a folder location
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Directory " + dir + " does not exist")
		os.Exit(1)
	}
	return dir
}

func AdvancedInternalRouter(host string, port string, wipDir string, staticDir string, dir string) {
	mux := http.NewServeMux()
	folder := findFolder(staticDir)
	mux.HandleFunc(wipDir, wisp)
	mux.HandleFunc(dir, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, folder+r.URL.Path)
	})
	fmt.Println("Listening on http://" + host + ":" + port + dir)
	if host == "0.0.0.0" {
		fmt.Println("Also listening on http://localhost:" + port + dir)
	}
	fmt.Println("Serving static files from " + folder + " at " + dir)
	fmt.Println("Wisp available at " + wipDir)
	http.ListenAndServe(host+":"+port, mux)
}
