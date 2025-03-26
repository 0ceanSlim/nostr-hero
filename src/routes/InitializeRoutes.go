package routes

import "net/http"

func InitializeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/registry", Registry)
}
