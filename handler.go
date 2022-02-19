package sockroom

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"nhooyr.io/websocket"
)

type Handler struct {
	chans Channels
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		log.Printf("error accepting websocket: %v", err)
		return
	}
	ctx := r.Context()
	s := &Subscriber{r.URL.Path, c, ctx}
	if err = h.chans.Join(s); err != nil {
		c.Close(websocket.StatusInternalError, "failed join")
		return
	}
	defer c.Close(websocket.StatusGoingAway, "bye-bye")
	defer h.chans.Unjoin(s)
	mt := make(map[string]string)
	for k, v := range r.Header {
		stripped := strings.TrimPrefix(k, "X-Sockroom-")
		if stripped != k && len(v) > 0 {
			mt[k] = v[0]
		}
	}
	for {
		type_, data, err := c.Read(ctx)
		if errors.Is(err, io.EOF) {
			return
		}
		if type_ == websocket.MessageBinary {
			err = h.chans.PublishBinary(s.Room, data, mt)
		} else {
			err = h.chans.PublishText(s.Room, string(data), mt)
		}
		if err != nil {
			err = s.WriteText("", map[string]interface{}{"error": "failed publish message"})
			if err != nil {
				return
			}
		}
	}
}

func New(chans Channels) Handler {
	return Handler{chans}
}
