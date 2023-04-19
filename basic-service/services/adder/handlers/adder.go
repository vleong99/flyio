package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func Adder(w http.ResponseWriter, r *http.Request) {
	a := r.URL.Query().Get("a")
	b := r.URL.Query().Get("b")
	ai, _ := strconv.Atoi(a)
	bi, _ := strconv.Atoi(b)
	fmt.Fprintf(w, "%d + %d = %d", ai, bi, ai+bi)
}
