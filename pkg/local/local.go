package local

import (
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/gudn/sockroom"
	_ "github.com/gudn/sockroom/internal/config"
	_ "github.com/gudn/sockroom/internal/log"
)

var logger *zap.Logger

func init() {
	logger = zap.L()
}

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
	mutex          *sync.RWMutex
	subsribers     map[string]map[*sockroom.Subscriber]struct{}
	quit           chan struct{}
	binaryMessages chan localBinaryMessage
	textMessages   chan localTextMessage
	nWorkers       uint
	bufSize        uint
}

func (l *LocalChannels) Join(s *sockroom.Subscriber) error {
	l.mutex.Lock()
	room := l.subsribers[s.Room]
	if room == nil {
		room = make(map[*sockroom.Subscriber]struct{})
		l.subsribers[s.Room] = room
		sockroom.ActiveRooms.Inc()
	}
	room[s] = struct{}{}
	l.mutex.Unlock()
	logger.Info("joined to room", zap.String("room", s.Room))
	sockroom.ActiveSockets.Inc()
	return nil
}

func (l *LocalChannels) Unjoin(s *sockroom.Subscriber) error {
	l.mutex.Lock()
	delete(l.subsribers[s.Room], s)
	if len(l.subsribers[s.Room]) == 0 {
		delete(l.subsribers, s.Room)
		sockroom.ActiveRooms.Dec()
	}
	l.mutex.Unlock()
	logger.Info("unjoined from room", zap.String("room", s.Room))
	sockroom.ActiveSockets.Dec()
	return nil
}

func (l *LocalChannels) PublishBinary(channel string, data []byte, mt interface{}) error {
	l.mutex.RLock()
	for k := range l.subsribers[channel] {
		l.binaryMessages <- localBinaryMessage{k, data, mt}
	}
	l.mutex.RUnlock()
	sockroom.InBuffer.Inc()
	return nil
}

func (l *LocalChannels) PublishText(channel, data string, mt interface{}) error {
	l.mutex.RLock()
	for k := range l.subsribers[channel] {
		l.textMessages <- localTextMessage{k, data, mt}
	}
	l.mutex.RUnlock()
	sockroom.InBuffer.Inc()
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
		sockroom.InBuffer.Dec()
	}
}

func (l *LocalChannels) setNWorkersLocked(n uint) {
	logger.Info("change nworkers", zap.Uint("from", l.nWorkers), zap.Uint("to", n))
	if n > l.nWorkers {
		delta := n - l.nWorkers
		var i uint
		for i = 0; i < delta; i++ {
			go localWorker(l.quit, l.binaryMessages, l.textMessages)
		}
	} else {
		delta := l.nWorkers - n
		var i uint
		for i = 0; i < delta; i++ {
			l.quit <- struct{}{}
		}
	}
	l.nWorkers = n
}

func (l *LocalChannels) SetNWorkers(n uint) {
	l.mutex.Lock()
	l.setNWorkersLocked(n)
	l.mutex.Unlock()
}

func (l *LocalChannels) Quit() {
	l.SetNWorkers(0)
}

func newWith(nWorkers, bufSize uint, mutex *sync.RWMutex) *LocalChannels {
	quit := make(chan struct{})
	binary := make(chan localBinaryMessage, bufSize)
	text := make(chan localTextMessage, bufSize)
	l := &LocalChannels{
		quit:           quit,
		binaryMessages: binary,
		textMessages:   text,
		subsribers:     make(map[string]map[*sockroom.Subscriber]struct{}),
		nWorkers:       nWorkers,
		bufSize:        bufSize,
		mutex:          mutex,
	}
	var i uint
	for i = 0; i < nWorkers; i++ {
		go localWorker(quit, binary, text)
	}
	return l
}

func readConfig() (uint, uint) {
	nWorkers := viper.GetUint("local.nworkers")
	bufSize := viper.GetUint("local.bufsize")
	if nWorkers == 0 {
		nWorkers = 0
	}
	if bufSize == 0 {
		bufSize = 16
	}
	return nWorkers, bufSize
}

func (l *LocalChannels) Reload() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	nWorkers, bufSize := readConfig()
	if bufSize != l.bufSize {
		oldQuit := l.quit
		u := l.nWorkers
		logger.Info(
			"recreate local channels",
			zap.Uint("nworkers", nWorkers),
			zap.Uint("bufSize", bufSize),
		)
		newL := newWith(nWorkers, bufSize, l.mutex)
		newL.subsribers = l.subsribers
		*l = *newL
		var i uint
		for i = 0; i < u; i++ {
			oldQuit <- struct{}{}
		}
	} else if nWorkers != l.nWorkers {
		l.setNWorkersLocked(nWorkers)
	}
}

func New() *LocalChannels {
	nWorkers, bufSize := readConfig()
	logger.Info(
		"create local channels",
		zap.Uint("nworkers", nWorkers),
		zap.Uint("bufSize", bufSize),
	)
	return newWith(nWorkers, bufSize, &sync.RWMutex{})
}
