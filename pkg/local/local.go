package local

import (
	"sync"

	"github.com/gudn/sockroom"
)

type localBinaryMessage struct {
	s    *sockroom.Subscriber
	data []byte
	mt   interface{}
}

type localTextMessage struct {
	s    *sockroom.Subscriber
	data string
	mt   interface{}
}

type LocalChannels struct {
	mutex          sync.RWMutex
	subsribers     map[string]map[*sockroom.Subscriber]struct{}
	quit           chan<- struct{}
	binaryMessages chan<- localBinaryMessage
	textMessages   chan<- localTextMessage
	nWorkers uint
}

func (l *LocalChannels) Join(s *sockroom.Subscriber) error {
	l.mutex.Lock()
	room := l.subsribers[s.Room]
	if room == nil {
		room = make(map[*sockroom.Subscriber]struct{})
		l.subsribers[s.Room] = room
	}
	room[s] = struct{}{}
	l.mutex.Unlock()
	return nil
}

func (l *LocalChannels) Unjoin(s *sockroom.Subscriber) error {
	l.mutex.Lock()
	delete(l.subsribers[s.Room], s)
	if len(l.subsribers[s.Room]) == 0 {
		delete(l.subsribers, s.Room)
	}
	l.mutex.Unlock()
	return nil
}

func (l *LocalChannels) PublishBinary(channel string, data []byte, mt interface{}) error {
	l.mutex.RLock()
	for k := range l.subsribers[channel] {
		l.binaryMessages <- localBinaryMessage{k, data, mt}
	}
	l.mutex.RUnlock()
	return nil
}

func (l *LocalChannels) PublishText(channel, data string, mt interface{}) error {
	l.mutex.RLock()
	for k := range l.subsribers[channel] {
		l.textMessages <- localTextMessage{k, data, mt}
	}
	l.mutex.RUnlock()
	return nil
}

func localWorker(quit <-chan struct{}, binary <-chan localBinaryMessage, text <-chan localTextMessage) {
	for {
		select {
		case b := <-binary:
			b.s.WriteBinary(b.data, b.mt)
		case t := <-text:
			t.s.WriteText(t.data, t.mt)
		case <-quit:
			return
		}
	}
}


func New(nWorkers uint) *LocalChannels {
	quit := make(chan struct{})
	binary := make(chan localBinaryMessage)
	text := make(chan localTextMessage)
	l := &LocalChannels{
		quit: quit,
		binaryMessages: binary,
		textMessages: text,
		subsribers: make(map[string]map[*sockroom.Subscriber]struct{}),
		nWorkers: nWorkers,
	}
	var i uint
	for i = 0; i < nWorkers; i++ {
		go localWorker(quit, binary, text)
	}
	return l
}

func (l *LocalChannels) Quit() {
	l.mutex.Lock()
	var i uint
	for i = 0; i < l.nWorkers; i++ {
		l.quit <- struct{}{}
	}
	l.nWorkers = 0
	l.mutex.Unlock()
}
