package handlers

import (
	"net/http"
	"os"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"bbcbear/sps30-exporter/internal/sensor"
)

func Init(sensorRef sensor.Sensor, isHealthy *atomic.Bool) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("ENABLE_CLEAN_ENDPOINT") != "true" {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("POST only"))
			return
		}

		if err := sensorRef.Clean(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to start cleaning"))
			slog.Error("Sensor cleaning failed", "error", err)
			return
		}

		slog.Info("Sensor cleaning started via /clean", "remote", r.RemoteAddr, "time", time.Now())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cleaning started"))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !isHealthy.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("sensor error"))
			slog.Warn("Health check failed", "remote", r.RemoteAddr)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return mux
}
