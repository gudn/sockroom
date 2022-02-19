package sockroom

import (
	"context"
	"encoding/json"

	"nhooyr.io/websocket"
)

type Subscriber struct {
	Room string
	C    *websocket.Conn
	context.Context
}

func (s *Subscriber) writeMessage(data, mt json.RawMessage, binary bool) error {
	contents := map[string]interface{}{"data": data, "metadata": mt, "binary": binary}
	encoded, err := json.Marshal(contents)
	if err != nil {
		return err
	}
	return s.C.Write(s.Context, websocket.MessageText, encoded)
}

func (s *Subscriber) WriteBinary(data []byte, mt interface{}) error {
	data, err := json.Marshal(data)
	if err != nil {
		return err
	}
	mte, err := json.Marshal(mt)
	if err != nil {
		return err
	}
	return s.writeMessage(json.RawMessage(data), json.RawMessage(mte), true)
}

func (s *Subscriber) WriteText(data string, mt interface{}) error {
	sdata, err := json.Marshal(data)
	if err != nil {
		return err
	}
	mte, err := json.Marshal(mt)
	if err != nil {
		return err
	}
	return s.writeMessage(json.RawMessage(sdata), json.RawMessage(mte), false)
}
