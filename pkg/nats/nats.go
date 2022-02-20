package nats

import (
	"encoding/json"
	"strings"

	cnats "github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/gudn/sockroom/internal/config"
	_ "github.com/gudn/sockroom/internal/log"
	"github.com/gudn/sockroom/pkg/local"
)

var logger *zap.Logger

func init() {
	logger = zap.L()
}

type NatsChannels struct {
	conn *cnats.Conn
	quit chan<- struct{}
	sub  *cnats.Subscription
	*local.LocalChannels
}

type natsMessage struct {
	Data   []byte
	Mt     interface{}
	Binary bool
}

func (n *NatsChannels) PublishText(channel string, data string, mt interface{}) error {
	msg := natsMessage{[]byte(data), mt, false}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	logger.Debug(
		"send message",
		zap.String("data", data),
		zap.Any("mt", mt),
		zap.ByteString("bytes", bytes),
	)
	return n.conn.Publish("sr."+channel, bytes)
}

func (n *NatsChannels) PublishBinary(channel string, data []byte, mt interface{}) error {
	msg := natsMessage{data, mt, true}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	logger.Debug(
		"send message",
		zap.ByteString("data", data),
		zap.Any("mt", mt),
		zap.ByteString("bytes", bytes),
	)
	return n.conn.Publish("sr."+channel, bytes)
}

func (n *NatsChannels) Quit() {
	n.sub.Unsubscribe()
	n.quit <- struct{}{}
	n.LocalChannels.Quit()
}

func natsUrl() string {
	if viper.IsSet("nats.url") {
		return viper.GetString("nats.url")
	}
	return cnats.DefaultURL
}

func natsWorker(msgs chan *cnats.Msg, quit <-chan struct{}, l *local.LocalChannels) {
	for {
		select {
		case msg := <-msgs:
			channel := strings.TrimPrefix(msg.Subject, "sr.")
			var data natsMessage
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				logger.Error("error decoding nats message", zap.Error(err))
			} else {
				if data.Binary {
					l.PublishBinary(channel, data.Data, data.Mt)
				} else {
					l.PublishText(channel, string(data.Data), data.Mt)
				}
			}
		case <-quit:
			close(msgs)
			return
		}
	}
}

func New() (*NatsChannels, error) {
	conn, err := cnats.Connect(natsUrl())
	if err != nil {
		return nil, err
	}
	quit := make(chan struct{})
	msgs := make(chan *cnats.Msg)
	subs, err := conn.ChanSubscribe("sr.*", msgs)

	l := local.New()
	go natsWorker(msgs, quit, l)

	return &NatsChannels{conn, quit, subs, l}, nil
}
