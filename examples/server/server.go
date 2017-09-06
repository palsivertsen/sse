package main

import (
	"log"
	"net/http"
	"time"

	"github.com/palsivertsen/sse"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveScript)
	mux.HandleFunc("/stream", func(w http.ResponseWriter, req *http.Request) {
		stream := sse.NewStream()
		defer stream.Close()
		go func() {
			for {
				select {
				case <-stream.CloseNotify():
					return
				default:
					log.Print("Sending sse message")
					stream.Send(sse.Event{
						Message: "The time is: " + time.Now().String(),
					})
					time.Sleep(time.Second)
				}
			}
		}()
		stream.ServeHTTP(w, req)
	})
	http.ListenAndServe(":8080", mux)
}

func serveScript(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
	<html>
	  <head>
	    <meta charset="utf-8">
	    <title>SSE Example</title>
	  </head>
	  <body>
	    <script>
	      var e = new EventSource("/stream")
	      e.addEventListener("message", (m) => {
	        document.body.innerText = m.data
	      }, false)
	    </script>
	  </body>
	</html>
`))
}
