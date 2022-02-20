package sockroom

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"nhooyr.io/websocket"

	_ "github.com/gudn/sockroom/internal/log"
)

var logger *zap.Logger

func init() {
	logger = zap.L()
}

type Handler struct {
	chans Channels
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		logger.Warn("error accepting connnection", zap.Error(err))
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
		if err != nil {
			s.WriteText("", map[string]interface{}{"error": "failed read message"})
			logger.Warn("failed read", zap.Error(err))
			return
		}
		if type_ == websocket.MessageBinary {
			err = h.chans.PublishBinary(s.Room, data, mt)
		} else {
			err = h.chans.PublishText(s.Room, string(data), mt)
		}
		if err != nil {
			err = s.WriteText("", map[string]interface{}{"error": "failed publish message"})
			logger.Warn("failed publish message", zap.String("room", s.Room), zap.Error(err))
			if err != nil {
				logger.Warn("failed send error", zap.Error(err))
				return
			}
		}
	}
}

func New(chans Channels) Handler {
	return Handler{chans}
}
