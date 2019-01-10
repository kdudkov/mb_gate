package main

import (
	"fmt"
	"net/http"
)

func (app *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := Job{Ch: make(chan bool)}
		app.Jobs <- &job
		<-job.Ch
		close(job.Ch)

		fmt.Println(job.Payload)
	}
}
