package prometheus

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

// Run prometheus http server, returns metrics for any addr path
func Run(r metrics.Registry, addr string) {
	s := http.Server{
		Addr:         addr,
		Handler:      handler(r),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Info("Starting prometheus http server", "addr", addr)
	if err := s.ListenAndServe(); err != nil {
		log.Warn("Unable to start prometheus metrics server", "addr", addr, "err", err)
	}
}

func handler(reg metrics.Registry) http.Handler {
	f := newFormatter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := getBuf()
		defer giveBuf(buf)

		reg.Each(func(name string, i interface{}) {
			switch m := i.(type) {
			case metrics.Counter:
				ms := m.Snapshot()
				f.counter(buf, name, ms)
			case metrics.Gauge:
				ms := m.Snapshot()
				f.gauge(buf, name, ms)
			case metrics.GaugeFloat64:
				ms := m.Snapshot()
				f.gaugeFloat64(buf, name, ms)
			case metrics.Histogram:
				ms := m.Snapshot()
				f.histogram(buf, name, ms)
			case metrics.Meter:
				ms := m.Snapshot()
				f.meter(buf, name, ms)
			case metrics.Timer:
				ms := m.Snapshot()
				f.timer(buf, name, ms)
			case metrics.ResettingTimer:
				ms := m.Snapshot()
				f.resettingTimer(buf, name, ms)
			}
		})

		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Content-Length", fmt.Sprint(buf.Len()))
		w.Write(buf.Bytes())
	})
}

var bufPool sync.Pool

func getBuf() *bytes.Buffer {
	buf := bufPool.Get()
	if buf == nil {
		return &bytes.Buffer{}
	}
	return buf.(*bytes.Buffer)
}

func giveBuf(buf *bytes.Buffer) {
	buf.Reset()
	bufPool.Put(buf)
}
