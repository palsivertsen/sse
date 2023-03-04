package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/palsivertsen/sse"
)

//go:embed html/*
var html embed.FS

func main() {
	if err := run(); err != nil {
		log.Printf("run: %s", err.Error())
	}
}

func run() error {
	fsys, err := fs.Sub(html, "html")
	if err != nil {
		return fmt.Errorf("embed html dir: %w", err)
	}

	mux := http.NewServeMux()
	// Serve files from "html" dir
	mux.Handle("/", http.FileServer(http.FS(fsys)))
	// Push counter through SSE
	mux.Handle("/sse/counter", sse.NewHTTPHandler(sse.HandlerFunc(func(w *sse.ResponseWriter, r *http.Request) {
		// Resume counter if this is a reconnect
		counter, _ := strconv.Atoi(r.Header.Get("Last-Event-ID"))

		for {
			counter++
			e := sse.Event{
				Name: "counter",
				ID:   strconv.Itoa(counter),
				Data: strings.NewReader(strconv.Itoa(counter)),
			}
			if err := w.PushEvent(&e); err != nil {
				log.Printf("Error writing event %d: %s", counter, err)
				return
			}
			time.Sleep(time.Second)
		}
	})))

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	log.Printf("Visit %s to view example", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
