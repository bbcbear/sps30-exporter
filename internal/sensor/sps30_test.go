package sensor

import (
    "testing"
	"math"
)

type mockBus struct {
    TxFunc func(w, r []byte) error
}

func (m *mockBus) Tx(w, r []byte) error {
    return m.TxFunc(w, r)
}

func TestSPS30Sensor_Read_CRCError(t *testing.T) {
    bus := &mockBus{
        TxFunc: func(w, r []byte) error {
            for i := range r {
                r[i] = 0xFF // Нарушаем CRC
            }
            return nil
        },
    }

    sensor := New(bus)
    _, err := sensor.Read()
    if err == nil {
        t.Error("expected CRC error, got nil")
    }
}

func TestSPS30Sensor_Read_Success(t *testing.T) {
    float32ToBE := func(f float32) []byte {
        bits := math.Float32bits(f)
        return []byte{
            byte(bits >> 24),
            byte(bits >> 16),
            calcCRC([]byte{byte(bits >> 24), byte(bits >> 16)}),
            byte(bits >> 8),
            byte(bits),
            calcCRC([]byte{byte(bits >> 8), byte(bits)}),
        }
    }

    // 10 метрик по 6 байт каждая
    data := append([]byte{}, float32ToBE(1.0)...) // PM1
    for i := 0; i < 9; i++ {
        data = append(data, float32ToBE(0.0)...)
    }

    bus := &mockBus{
        TxFunc: func(w, r []byte) error {
            copy(r, data)
            return nil
        },
    }

    sensor := New(bus)
    m, err := sensor.Read()
    if err != nil {
        t.Fatalf("expected success, got error: %v", err)
    }

    if m.PM1Mass != 1.0 {
        t.Errorf("expected PM1Mass = 1.0, got %f", m.PM1Mass)
    }
}


func TestSPS30Sensor_Init(t *testing.T) {
    var received []byte

    bus := &mockBus{
        TxFunc: func(w, r []byte) error {
            received = append(received, w...)
            return nil
        },
    }

    sensor := New(bus)
    err := sensor.Init()
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if len(received) != 5 {
        t.Fatalf("expected 5 bytes sent, got %d: %v", len(received), received)
    }

    // Проверка команды (без CRC)
    if received[0] != 0x00 || received[1] != 0x10 {
        t.Errorf("unexpected command bytes: %v", received[:2])
    }

    // Проверка аргумента и CRC
    if received[2] != 0x03 || received[3] != 0x00 {
        t.Errorf("unexpected argument bytes: %v", received[2:4])
    }

    expectedCRC := calcCRC([]byte{0x03, 0x00})
    if received[4] != expectedCRC {
        t.Errorf("unexpected CRC: got 0x%X, want 0x%X", received[4], expectedCRC)
    }
}
