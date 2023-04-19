package main

import (
	"net/http"

	"github.com/vleong99/basic-service/services/adder"
)

func main() {
	http.HandleFunc("/", handlers.Adder)
	http.ListenAndServe(":8000", nil)
}
