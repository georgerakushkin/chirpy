package main

import "net/http"

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// this handler is used to check if the server is ready to accept requests. It takes in a response writer and pointer to the request.
// It then sets the Header to Content-Type to text/plain and the status code to 200 OK.
// It then writes the status text for 200 OK to the response body.
