package sockroom

type Channels interface {
	Join(s *Subscriber) error
	Unjoin(s *Subscriber) error
	PublishBinary(channel string, data []byte, mt interface{}) error
	PublishText(channel, data string, mt interface{}) error

	Quit()
	Reload()
}
