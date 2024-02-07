package wisp 

import (
    "net/http"
)

func Router() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", wispHandler)
    http.ListenAndServe(":8080", mux)
}
