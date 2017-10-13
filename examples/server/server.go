package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/palsivertsen/sse"
)

func main() {
	var (
		addr = flag.String("addr", ":8080", "Server listen address")
	)
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveScript)
	mux.HandleFunc("/stream", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Client %s connected", req.RemoteAddr)
		stream := sse.NewStream()
		defer stream.Close()
		go func() {
			for {
				select {
				case <-stream.CloseNotify():
					log.Printf("Client %s disconnected", req.RemoteAddr)
					return
				default:
					log.Print("Sending sse message to ", req.RemoteAddr)
					stream.Send(sse.Event{
						Message: "The time is: " + time.Now().String(),
					})
					time.Sleep(time.Second)
				}
			}
		}()
		stream.ServeHTTP(w, req)
	})
	log.Printf("A live demo is shortly available at %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
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
