package main

import (
	"io"
	"log"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"crypto/sha256"
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
        headers["Transfer-Encoding"] = "chunked" 
		headers["Trailers"] = "X-Content-SHA256, X-Content-Length"
		w.WriteHeaders(headers)

		buf := make([]byte, 1024)
		totalBytesBody := 0
		var respBody []byte

		for { 
			n, err := res.Body.Read(buf)
			respBody = append(respBody, buf[:n]...)
			totalBytesBody += n
			
			if n == 0 {
				w.WriteChunkedBodyDone()
				break
			}

			n, err = w.WriteChunkedBody(buf[:n])


			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("Error: when reading chunk %v\n", err)
				break
			}
		}

		trailers := response.GetDefaultTrailers()

		hash := sha256.Sum256(respBody)
		trailers["X-Content-SHA256"] = fmt.Sprintf("%x", hash)
		trailers["X-Content-Length"] = fmt.Sprintf("%d", totalBytesBody)

		err = w.WriteTrailers(trailers)
		if err != nil {
			log.Println(err)
		}
	} else if req.RequestLine.RequestTarget == "/video" {
		if err := w.WriteStatusLine(200); err != nil {
			// Should respond with an internal server error
			log.Println(err)
		}	

		body, err := os.ReadFile("./assets/vim.mp4")
		if err != nil {
			log.Println(err)
		}

		headers := response.GetDefaultHeaders(len(body))
		headers["content-type"] = "video/mp4"
		w.WriteHeaders(headers)
		w.WriteBody([]byte(body))

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
