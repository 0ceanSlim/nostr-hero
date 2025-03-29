package routes

import "net/http"

func InitializeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/legacy-registry", LegacyRegistry)
	mux.HandleFunc("/alpha-registry", AlphaRegistry)
	mux.HandleFunc("/discover", Discover)
}
