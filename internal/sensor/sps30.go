package sensor

import (
    "encoding/binary"
    "fmt"
    "math"
    "time"
	"log/slog"
)

type Bus interface {
    Tx(w, r []byte) error
}

type Sensor interface {
    Init() error
    Stop() error
    Clean() error
    Read() (Measurement, error)
    IsMeasuring() (bool, error)
}

type SPS30Sensor struct {
    bus Bus
}

type Measurement struct {
    PM1Mass      float32
    PM2_5Mass    float32
    PM4Mass      float32
    PM10Mass     float32
    PM0_5Num     float32
    PM1Num       float32
    PM2_5Num     float32
    PM4Num       float32
    PM10Num      float32
    ParticleSize float32
}

const (
    cmdReadMeasurement   = 0x0300
    cmdStartMeasurement  = 0x0010
    cmdStopMeasurement   = 0x0104
    cmdFanCleaning       = 0x5607
    cmdStatusMeasurement = 0x0202
)

func New(bus Bus) *SPS30Sensor {
    return &SPS30Sensor{bus: bus}
}

func (s *SPS30Sensor) Init() error {
	slog.Info("Sending start measurement command to SPS30")
	cmd := []byte{0x03, 0x00}
    return s.sendCommand(cmdStartMeasurement, cmd)
}

func (s *SPS30Sensor) Stop() error {
    return s.sendCommand(cmdStopMeasurement, nil)
}

func (s *SPS30Sensor) Clean() error {
    return s.sendCommand(cmdFanCleaning, nil)
}

func (s *SPS30Sensor) Read() (Measurement, error) {
	if err := s.sendCommand(cmdReadMeasurement, nil); err != nil {
		return Measurement{}, fmt.Errorf("sendCommand error: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 60)
	var txErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := s.bus.Tx(nil, buf); err == nil {
			txErr = nil
			break
		} else {
			txErr = err
			time.Sleep(50 * time.Millisecond)
		}
	}
	if txErr != nil {
		return Measurement{}, fmt.Errorf("I2C Tx error after retries: %w", txErr)
	}

	values := make([]float32, 0, 10)
	for i := 0; i < 60; i += 6 {
		if !validCRC(buf[i:i+2], buf[i+2]) {
			return Measurement{}, fmt.Errorf("CRC error at bytes %d-%d", i, i+1)
		}
		if !validCRC(buf[i+3:i+5], buf[i+5]) {
			return Measurement{}, fmt.Errorf("CRC error at bytes %d-%d", i+3, i+4)
		}

		// собираем float32 из 4 байт (2 + 2)
		dataBytes := []byte{buf[i], buf[i+1], buf[i+3], buf[i+4]}
		val := math.Float32frombits(binary.BigEndian.Uint32(dataBytes))
		values = append(values, val)
	}

	return Measurement{
		PM1Mass:      values[0],
		PM2_5Mass:    values[1],
		PM4Mass:      values[2],
		PM10Mass:     values[3],
		PM0_5Num:     values[4],
		PM1Num:       values[5],
		PM2_5Num:     values[6],
		PM4Num:       values[7],
		PM10Num:      values[8],
		ParticleSize: values[9],
	}, nil
}

func (s *SPS30Sensor) sendCommand(cmd uint16, args []byte) error {
    if args != nil && len(args)%2 != 0 {
        return fmt.Errorf("arguments length must be even (pairs of bytes)")
    }

    buf := []byte{byte(cmd >> 8), byte(cmd & 0xFF)}
    if args != nil {
        for i := 0; i < len(args); i += 2 {
            chunk := args[i : i+2]
            crc := calcCRC(chunk)
            buf = append(buf, chunk[0], chunk[1], crc)
        }
    }

    for attempt := 1; attempt <= 3; attempt++ {
        if err := s.bus.Tx(buf, nil); err == nil {
            return nil
        }
        time.Sleep(50 * time.Millisecond)
    }
    return fmt.Errorf("I2C Tx failed after retries")
}

func (s *SPS30Sensor) IsMeasuring() (bool, error) {
	// Отправляем команду 0x0202 (Status)
	if err := s.sendCommand(cmdStatusMeasurement, nil); err != nil {
		return false, fmt.Errorf("sendCommand failed: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	buf := make([]byte, 3)
	if err := s.bus.Tx(nil, buf); err != nil {
		return false, fmt.Errorf("Tx failed: %w", err)
	}

	if !validCRC(buf[:2], buf[2]) {
		return false, fmt.Errorf("CRC error in status response")
	}

	status := binary.BigEndian.Uint16(buf[:2])
	return status == 0x0001, nil
}

func validCRC(data []byte, crc byte) bool {
    return calcCRC(data) == crc
}

func calcCRC(data []byte) byte {
    crc := byte(0xFF)
    for _, b := range data {
        crc ^= b
        for i := 0; i < 8; i++ {
            if crc&0x80 != 0 {
                crc = (crc << 1) ^ 0x31
            } else {
                crc <<= 1
            }
        }
    }
    return crc
}
