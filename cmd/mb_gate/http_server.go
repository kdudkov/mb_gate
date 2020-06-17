package main

import (
	"net/http"
)

func (app *App) setRoute() {
	http.HandleFunc("/", app.handleIndex())
}

func (app *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			app.Logger.Error("can't write")
		}
	}
}
