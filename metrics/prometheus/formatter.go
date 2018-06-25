package prometheus

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/metrics"
)

const (
	metricGauage  = "gauage"
	metricSummary = "summary"
)

type formatter struct {
	metrics     map[string]*metric
	metricsLock sync.RWMutex
}

func newFormatter() *formatter {
	return &formatter{
		metrics: make(map[string]*metric),
	}
}

func (f *formatter) counter(b *bytes.Buffer, name string, m metrics.Counter) {
	pm := f.normalizeMetric(name, metricGauage)
	b.Write(pm.header())
	b.Write(pm.keyTagValue("value", m.Count()))
}

func (f *formatter) gauge(b *bytes.Buffer, name string, m metrics.Gauge) {
	pm := f.normalizeMetric(name, metricGauage)
	b.Write(pm.header())
	b.Write(pm.keyTagValue("value", m.Value()))
}

func (f *formatter) gaugeFloat64(b *bytes.Buffer, name string, m metrics.GaugeFloat64) {
	pm := f.normalizeMetric(name, metricGauage)
	b.Write(pm.header())
	b.Write(pm.keyTagValue("value", m.Value()))
}

func (f *formatter) histogram(b *bytes.Buffer, name string, m metrics.Histogram) {
	pm := f.normalizeMetric(name, metricSummary)
	ps := m.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
	b.Write(pm.header())
	b.Write(pm.keyTagValue("count", m.Count()))
	b.Write(pm.keyTagValue("max", m.Max()))
	b.Write(pm.keyTagValue("mean", m.Mean()))
	b.Write(pm.keyTagValue("min", m.Min()))
	b.Write(pm.keyTagValue("stddev", m.StdDev()))
	b.Write(pm.keyTagValue("variance", m.Variance()))
	b.Write(pm.keyTagValue("p50", ps[0]))
	b.Write(pm.keyTagValue("p75", ps[1]))
	b.Write(pm.keyTagValue("p95", ps[2]))
	b.Write(pm.keyTagValue("p99", ps[3]))
	b.Write(pm.keyTagValue("p999", ps[4]))
	b.Write(pm.keyTagValue("p9999", ps[5]))
}

func (f *formatter) meter(b *bytes.Buffer, name string, m metrics.Meter) {
	pm := f.normalizeMetric(name, metricGauage)
	b.Write(pm.header())
	b.Write(pm.keyTagValue("count", m.Count()))
	b.Write(pm.keyTagValue("m1", m.Rate1()))
	b.Write(pm.keyTagValue("m5", m.Rate5()))
	b.Write(pm.keyTagValue("m15", m.Rate15()))
	b.Write(pm.keyTagValue("mean", m.RateMean()))
}

func (f *formatter) timer(b *bytes.Buffer, name string, m metrics.Timer) {
	pm := f.normalizeMetric(name, metricSummary)
	ps := m.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
	b.Write(pm.header())
	b.Write(pm.keyTagValue("count", m.Count()))
	b.Write(pm.keyTagValue("max", m.Max()))
	b.Write(pm.keyTagValue("mean", m.Mean()))
	b.Write(pm.keyTagValue("min", m.Min()))
	b.Write(pm.keyTagValue("stddev", m.StdDev()))
	b.Write(pm.keyTagValue("variance", m.Variance()))
	b.Write(pm.keyTagValue("p50", ps[0]))
	b.Write(pm.keyTagValue("p75", ps[1]))
	b.Write(pm.keyTagValue("p95", ps[2]))
	b.Write(pm.keyTagValue("p99", ps[3]))
	b.Write(pm.keyTagValue("p999", ps[4]))
	b.Write(pm.keyTagValue("p9999", ps[5]))
	b.Write(pm.keyTagValue("m1", m.Rate1()))
	b.Write(pm.keyTagValue("m5", m.Rate5()))
	b.Write(pm.keyTagValue("m15", m.Rate15()))
	b.Write(pm.keyTagValue("meanrate", m.RateMean()))
}

func (f *formatter) resettingTimer(b *bytes.Buffer, name string, m metrics.ResettingTimer) {
	if len(m.Values()) <= 0 {
		return
	}
	ps := m.Percentiles([]float64{50, 95, 99})
	val := m.Values()
	pm := f.normalizeMetric(name, metricSummary)
	b.Write(pm.header())
	b.Write(pm.keyTagValue("count", len(val)))
	b.Write(pm.keyTagValue("max", val[len(val)-1]))
	b.Write(pm.keyTagValue("mean", m.Mean()))
	b.Write(pm.keyTagValue("min", val[0]))
	b.Write(pm.keyTagValue("p50", ps[0]))
	b.Write(pm.keyTagValue("p95", ps[1]))
	b.Write(pm.keyTagValue("p99", ps[2]))
}

func (f *formatter) normalizeMetric(mkey, mtype string) *metric {
	f.metricsLock.RLock()
	if m, ok := f.metrics[mkey]; ok {
		f.metricsLock.RUnlock()
		return m
	}
	f.metricsLock.RUnlock()

	m := newMetric(mkey, mtype)

	f.metricsLock.Lock()
	if m, ok := f.metrics[mkey]; ok {
		f.metricsLock.Unlock()
		return m
	}
	f.metrics[mkey] = m
	f.metricsLock.Unlock()

	return m
}

var (
	metricHeaderTemplate = "# HELP %s metric\n# TYPE %s %s\n"
	metricKeyTagTemplate = "%s{mtype=\"%s\",aggr=\"%s\"} %v\n"
)

type metric struct {
	key            string
	mtype          string
	helpTypeHeader []byte
}

func newMetric(mkey, mtype string) *metric {
	m := &metric{
		key:   strings.Replace(mkey, "/", "_", -1),
		mtype: mtype,
	}
	m.helpTypeHeader = []byte(fmt.Sprintf(metricHeaderTemplate, m.key, m.key, m.mtype))
	return m
}

func (m *metric) header() []byte {
	return m.helpTypeHeader
}

func (m *metric) keyTagValue(tag string, value interface{}) []byte {
	return []byte(fmt.Sprintf(metricKeyTagTemplate, m.key, m.mtype, tag, value))
}
