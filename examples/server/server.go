package main

import (
	"log"
	"net/http"
	"time"

	"github.com/palsivertsen/sse"
)

func main() {
	stream := sse.NewStream()
	defer stream.Close()
	go func() {
		for {
			log.Print("Sending sse message")
			stream.Send(sse.Event{
				Message: "The time is: " + time.Now().String(),
			})
			time.Sleep(time.Second)
		}
	}()
	http.ListenAndServe(":8080", stream)
}
