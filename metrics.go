package sockroom

import (
	"github.com/prometheus/client_golang/prometheus"

	_ "github.com/gudn/sockroom/internal/log"
)

var (
	ActiveSockets prometheus.Gauge
	ActiveRooms prometheus.Gauge
	InBuffer prometheus.Gauge
)

func init() {
	ActiveSockets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_sockets",
		Help: "Count of active websocket connections",
	})
	ActiveRooms = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_rooms",
		Help: "Count of active opened rooms",
	})
	InBuffer = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "messages_in_buffer",
		Help: "Count of not sended messages in buffer",
	})

	prometheus.MustRegister(ActiveRooms, ActiveSockets, InBuffer)
}
