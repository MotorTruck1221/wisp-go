package wisp 

import (
    "net/http"
)

func InternalRouter() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", wisp)
    http.ListenAndServe(":8080", mux)
}
