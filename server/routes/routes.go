package routes

import "net/http"

// RegisterStatic registers static file serving routes
func RegisterStatic(mux *http.ServeMux) {
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("www/res/"))))
	mux.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("www/dist/"))))
	mux.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("game-data/"))))
}

// RegisterPages registers all page routes
func RegisterPages(mux *http.ServeMux) {
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/legacy-registry", LegacyRegistry)
	mux.HandleFunc("/alpha-registry", AlphaRegistry)
	mux.HandleFunc("/discover", Discover)
	mux.HandleFunc("/load-save", LoadSave)
	mux.HandleFunc("/saves", SavesHandler)
	mux.HandleFunc("/new-game", NewGameHandler)
	mux.HandleFunc("/game", GameHandler)
	mux.HandleFunc("/settings", SettingsHandler)
}
