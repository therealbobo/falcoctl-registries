package main

import (
	"fmt"
	"oauth/server"

	// "fmt"
	// "io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/v2/myrepo/tags/list", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(server.FormatRequest(r))
	})
	err := http.ListenAndServe(":443", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
