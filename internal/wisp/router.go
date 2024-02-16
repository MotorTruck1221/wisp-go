package wisp 

import (
    "net/http"
    "fmt"
)

func InternalRouter(host string, port string, dir string) {
    mux := http.NewServeMux()
    mux.HandleFunc(dir, wisp)
    fmt.Println("Listening on http://" + host + ":" + port + dir)
    if host == "0.0.0.0" { fmt.Println("Also listening on http://localhost:" + port + dir) }
    http.ListenAndServe(host + ":" + port, mux)
}
