package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (app *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Error("can't write")
		}
	}
}
