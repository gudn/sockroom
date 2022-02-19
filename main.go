package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"nhooyr.io/websocket"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	s, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		log.Printf("error accepting websocket: %v", err)
		return
	}
	defer s.Close(websocket.StatusTryAgainLater, "bye-bye")
	ctx := r.Context()
	for {
		mt, bytes, err := s.Read(ctx)
		if err != nil {
			log.Printf("error reading: %v", err)
			return
		}
		if mt == websocket.MessageText {
			log.Printf("got text: %s", bytes)
		} else {
			log.Printf("got binary: %v", bytes)
		}
		err = s.Write(ctx, mt, bytes)
		if err != nil {
			log.Printf("error writing: %v", err)
			return
		}
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

	http.HandleFunc("/", handler)

	return http.Serve(l, nil)
}
