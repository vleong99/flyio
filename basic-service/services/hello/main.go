package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {})
	http.ListenAndServe(":8000", nil)
}
