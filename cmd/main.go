package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gudn/sockroom"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("please provide an address to listen on as the first argument")
	}

	l, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())

	http.HandleFunc("/", sockroom.Handler)

	return http.Serve(l, nil)
}
