
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"bbcbear/sps30-exporter/internal/sensor"
)

var (
	SensorMetrics *prometheus.GaugeVec
	readErrors    prometheus.Counter
)

func Register() {
	SensorMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sps30_value",
			Help: "SPS30 sensor values with type and unit labels",
		},
		[]string{"type", "unit"},
	)
	readErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_read_errors_total",
			Help: "Total number of failed sensor reads",
		},
	)

	prometheus.MustRegister(SensorMetrics)
	prometheus.MustRegister(readErrors)
}

func Unregister() {
	prometheus.Unregister(SensorMetrics)
	prometheus.Unregister(readErrors)
}

func IncReadError() {
	readErrors.Inc()
}

func Update(m sensor.Measurement) {
	SensorMetrics.WithLabelValues("pm1_mass", "µg/m³").Set(float64(m.PM1Mass))
	SensorMetrics.WithLabelValues("pm2_5_mass", "µg/m³").Set(float64(m.PM2_5Mass))
	SensorMetrics.WithLabelValues("pm4_mass", "µg/m³").Set(float64(m.PM4Mass))
	SensorMetrics.WithLabelValues("pm10_mass", "µg/m³").Set(float64(m.PM10Mass))

	SensorMetrics.WithLabelValues("pm0_5_num", "particles/cm³").Set(float64(m.PM0_5Num))
	SensorMetrics.WithLabelValues("pm1_num", "particles/cm³").Set(float64(m.PM1Num))
	SensorMetrics.WithLabelValues("pm2_5_num", "particles/cm³").Set(float64(m.PM2_5Num))
	SensorMetrics.WithLabelValues("pm4_num", "particles/cm³").Set(float64(m.PM4Num))
	SensorMetrics.WithLabelValues("pm10_num", "particles/cm³").Set(float64(m.PM10Num))

	SensorMetrics.WithLabelValues("particle_size", "µm").Set(float64(m.ParticleSize))
}
