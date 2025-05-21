package app

import (
    "context"
    "testing"
    "time"

    "bbcbear/sps30-exporter/internal/metrics"
    "bbcbear/sps30-exporter/internal/sensor"
)

type mockSensor struct {
    ReadFunc         func() (sensor.Measurement, error)
    IsMeasuringFunc  func() (bool, error)
}

func (m *mockSensor) Init() error {
    return nil
}

func (m *mockSensor) Stop() error {
    return nil
}

func (m *mockSensor) Clean() error {
    return nil
}

func (m *mockSensor) Read() (sensor.Measurement, error) {
    return m.ReadFunc()
}

func (m *mockSensor) IsMeasuring() (bool, error) {
    return m.IsMeasuringFunc()
}

func TestApp_StartPolling(t *testing.T) {
    // Подменяем UpdateFunc на заглушку, чтобы избежать nil panic
    metrics.Register()
    s := &mockSensor{
        ReadFunc: func() (sensor.Measurement, error) {
            return sensor.Measurement{}, nil
        },
        IsMeasuringFunc: func() (bool, error) {
            return true, nil
        },
    }

    a := &App{
        Sensor:   s,
        Interval: 10 * time.Millisecond,
    }

    ctx, cancel := context.WithCancel(context.Background())
    go a.StartPolling(ctx, cancel)

    time.Sleep(50 * time.Millisecond)
    cancel()
}
