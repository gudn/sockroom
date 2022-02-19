package sockroom

import (
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

func Handler(w http.ResponseWriter, r *http.Request) {
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
