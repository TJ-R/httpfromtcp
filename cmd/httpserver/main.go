package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"github.com/TJ-R/httpfromtcp/internal/request"
	"github.com/TJ-R/httpfromtcp/internal/response"
	"github.com/TJ-R/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerFunc(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		w.WriteStatusLine(400)
		body := 
`
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`

		headers := response.GetDefaultHeaders(len(body))
		headers["content-type"] = "text/html"
		w.WriteHeaders(headers)
		w.WriteBody([]byte(body))
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		w.WriteStatusLine(500)
		body := 
`
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`

		headers := response.GetDefaultHeaders(len(body))
		headers["content-type"] = "text/html"
		w.WriteHeaders(headers)
		w.WriteBody([]byte(body))
	} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {

		path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
		url := "https://httpbin.org/" + path
		res, err := http.Get(url)	
		if err != nil { 
			log.Println(err)
		} 

		w.WriteStatusLine(response.StatusOk)

		headers := response.GetDefaultHeaders(0)
		delete(headers, "content-length")
        headers["transfer-encoding"] = "chunked" 
		w.WriteHeaders(headers)

		buf := make([]byte, 32)

		for { 
			n, err := res.Body.Read(buf)
			log.Println(n)
			
			if n == 0 {
				w.WriteChunkedBodyDone()
				break
			}

			_, err = w.WriteChunkedBody(buf[:n])

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("Error: when reading chunk %v\n", err)
				break
			}
		}
							

	} else {
		w.WriteStatusLine(200)
		body := 
`
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`

		headers := response.GetDefaultHeaders(len(body))
		headers["content-type"] = "text/html"
		w.WriteHeaders(headers)
		w.WriteBody([]byte(body))
	}
}
