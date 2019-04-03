package main

import (
	"sync"
	"runtime"
	"encoding/binary"
	"github.com/goburrow/serial"
	"github.com/tbrandon/mbserver"
	"github.com/ggueret/mbpio/gpio"
	"github.com/ggueret/mbpio/config"
	"github.com/ggueret/mbpio/modbus"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	mb		*mbserver.Server
	cfg		*config.Config
	mu		sync.Mutex
	done	chan struct{}
	quit	chan struct{}
	wg		sync.WaitGroup
	pollers	map[string]func([]int)
}

var (
	PWM_DEFAULT_CYCLE uint32 = 100
	PWM_DEFAULT_FREQ int = 10000  // 10kHz / 0,01MHz
	OUTPUT_DEFAULT_VALUE = 0
)

var Version = "0.0.1"

func NewServer(config_path string) (*Server, error) {

	cfg, err := config.Load(config_path)
	if err != nil {
		return nil, err
	}

	return &Server{
		mb: mbserver.NewServer(),
		cfg: cfg,
		done: make(chan struct{}),
		quit: make(chan struct{}),
		wg: sync.WaitGroup{},
		pollers: make(map[string]func([]int)),
	}, nil
}

func (s *Server) Start() error {
	log.Printf("starting mbpio v%s for %s/%s", Version, runtime.GOOS, runtime.GOARCH)

	err := gpio.Open()
	if err != nil {
		return err
	}
	defer gpio.Close()

	s.RegisterPollers()

	s.mb.RegisterFunctionHandler(0x1, s.ReadCoils)
	s.mb.RegisterFunctionHandler(0x2, s.ReadDiscreteInputs)
	s.mb.RegisterFunctionHandler(0x3, s.ReadHoldingRegisters)
	s.mb.RegisterFunctionHandler(0x4, s.ReadInputRegisters)

	s.mb.RegisterFunctionHandler(0x5, s.WriteSingleCoil)
	s.mb.RegisterFunctionHandler(0x6, s.WriteHoldingRegister)
	s.mb.RegisterFunctionHandler(0xf, s.WriteMultipleCoils)
	s.mb.RegisterFunctionHandler(0x10, s.WriteHoldingRegisters)

	// init outputs as mb.Coils and mb.HoldingRegisters for PWM
	for addr, output := range s.cfg.Outputs {
		if output.Pwm != nil {
			// Output is a HoldingRegister
			pwmFreq := output.Pwm.Freq
			if pwmFreq == nil {
				pwmFreq = &PWM_DEFAULT_FREQ
			}

			pwmCycle := output.Pwm.Cycle
			if pwmCycle == nil {
				pwmCycle = &PWM_DEFAULT_CYCLE
			}

			pwmDuty := uint32(OUTPUT_DEFAULT_VALUE)

			log.WithFields(log.Fields{"addr": addr, "pin": output.Pin, "freq": *pwmFreq, "duty": pwmDuty, "cycle": *pwmCycle}).Debug("Registering i/o output holding register")
			output.Pin.Mode(gpio.Pwm)
			output.Pin.Freq(*pwmFreq)
			output.Pin.DutyCycle(pwmDuty, *pwmCycle)
		} else {
			// Output is a Coil
			log.WithFields(log.Fields{"addr": addr, "pin": output.Pin, "state": gpio.Low}).Debug("Registering i/o output coil")
			output.Pin.Mode(gpio.Output)
			output.Pin.Write(gpio.Low)
		}
	}

	// init inputs as mb.DiscreteInputs for on/off and mb.InputRegisters for the others
	pollersInputs := make(map[string][]int)

	for addr, input := range s.cfg.Inputs {
		if input.Poller != nil {
			// Input is a InputRegister
			if _, ok := s.pollers[input.Poller.Type]; ok {
				log.WithFields(log.Fields{"addr": addr, "pin": input.Pin, "type": input.Poller.Type, "value": input.Poller.Value}).Debug("Registering i/o input register")
				pollersInputs[input.Poller.Type] = append(pollersInputs[input.Poller.Type], addr)
			}
		} else {
			// Input is a DiscreteInput
			log.WithFields(log.Fields{"addr": addr, "pin": input.Pin}).Debug("Registering i/o input discrete")
			input.Pin.Mode(gpio.Input)
		}
	}

	// run the selected inputs pollers
	for pollerName, pollerFunc := range s.pollers {
		if pollerInputs, ok := pollersInputs[pollerName]; ok {
			log.Debugf("Spawning the %s poller...", pollerName)
			go pollerFunc(pollerInputs)
		}
	}

	if s.cfg.EnableRTU == true {
		log.Infof("Listening to RTU address %s", s.cfg.RTUAddress)
		err = s.mb.ListenRTU(&serial.Config{
			Address: s.cfg.RTUAddress,
			BaudRate: s.cfg.RTUBaudRate,
			DataBits: s.cfg.RTUDataBits,
			StopBits: s.cfg.RTUStopBits,
			Parity: s.cfg.RTUParity,
			Timeout: s.cfg.RTUTimeout,
		})
		if err != nil {
			return err
		}
	}

	log.Infof("Listening to TCP address %s", s.cfg.ListenOn)
	err = s.mb.ListenTCP(s.cfg.ListenOn)
	if err != nil {
		return err
	}

	for {
		select {
		case <- s.quit:
			s.wg.Wait()
			close(s.done)
			return nil

		default:
		}
	}
	return nil
}

func (s *Server) Stop() {
	close(s.quit)
	<-s.done
}

func (s *Server) ReadCoils(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, numRegs, endRegister := modbus.RegisterAddressAndNumber(frame)
	if endRegister > 65535 {
		return []byte{}, &mbserver.IllegalDataAddress
	}
	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, value := range mb.Coils[register:endRegister] {
		if value != 0 {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, &mbserver.Success
}

func (s *Server) ReadDiscreteInputs(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, numRegs, endRegister := modbus.RegisterAddressAndNumber(frame)
	if endRegister > 65535 {
		return []byte{}, &mbserver.IllegalDataAddress
	}
	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, value := range mb.DiscreteInputs[register:endRegister] {
		if value != 0 {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, &mbserver.Success
}

func (s *Server) ReadHoldingRegisters(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, numRegs, endRegister := modbus.RegisterAddressAndNumber(frame)
	if endRegister > 65536 {
		return []byte{}, &mbserver.IllegalDataAddress
	}
	return append([]byte{byte(numRegs * 2)}, Uint16ToBytes(mb.HoldingRegisters[register:endRegister])...), &mbserver.Success
}

func (s *Server) ReadInputRegisters(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, numRegs, endRegister := modbus.RegisterAddressAndNumber(frame)
	if endRegister > 65536 {
		return []byte{}, &mbserver.IllegalDataAddress
	}
	return append([]byte{byte(numRegs * 2)}, Uint16ToBytes(mb.InputRegisters[register:endRegister])...), &mbserver.Success
}

func (s *Server) WriteSingleCoil(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, value := modbus.RegisterAddressAndValue(frame)
	// TODO Should we use 0 for off and 65,280 (FF00 in hexadecimal) for on?
	if value != 0 {
		value = 1
	}
	if output, ok := s.cfg.Outputs[register]; ok {
		if output.Pwm != nil {
			duty := uint32(0)
			if value == 1 {
				duty = *output.Pwm.Cycle
			}
			output.Pin.DutyCycle(duty, *output.Pwm.Cycle)
		} else {
			output.Pin.Write(gpio.State(value))
		}
		mb.Coils[register] = byte(value)
		return frame.GetData()[0:4], &mbserver.Success
	}
	return []byte{}, &mbserver.IllegalDataAddress
}

func (s *Server) WriteHoldingRegister(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, value := modbus.RegisterAddressAndValue(frame)
	if output, ok := s.cfg.Outputs[register]; ok {
		if output.Pwm != nil {
			if uint32(value) > *output.Pwm.Cycle {
				value = uint16(*output.Pwm.Cycle)
			}
			output.Pin.DutyCycle(uint32(value), *output.Pwm.Cycle)
			mb.HoldingRegisters[register] = value
			return frame.GetData()[0:4], &mbserver.Success
		}
	}
	return []byte{}, &mbserver.IllegalDataAddress
}

func (s *Server) WriteMultipleCoils(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, numRegs, endRegister := modbus.RegisterAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if endRegister > 65536 {
		return []byte{}, &mbserver.IllegalDataAddress
	}

	bitCount := 0
	for i, value := range valueBytes {
		for bitPos := uint(0); bitPos < 8; bitPos++ {
			addr := register+(i*8)+int(bitPos)
			addrVal := bitAtPosition(value, bitPos)

			if output, ok := s.cfg.Outputs[addr]; ok {
				if output.Pwm != nil {
					duty := uint32(0)
					if addrVal == 1 {
						duty = *output.Pwm.Cycle
					}
					output.Pin.DutyCycle(duty, *output.Pwm.Cycle)
				} else {
					output.Pin.Write(gpio.State(uint16(addrVal)))
				}
				mb.Coils[addr] = addrVal
			}
			bitCount++
			if bitCount >= numRegs {
				break
			}
		}
		if bitCount >= numRegs {
			break
		}
	}

	return frame.GetData()[0:4], &mbserver.Success
}

func (s *Server) WriteHoldingRegisters(mb *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	s.mu.Lock()
	defer s.mu.Unlock()
	register, numRegs, _ := modbus.RegisterAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]
	var exception *mbserver.Exception
	var data []byte

	if len(valueBytes)/2 != numRegs {
		exception = &mbserver.IllegalDataAddress
	}

	// Copy data to memory
	values := mbserver.BytesToUint16(valueBytes)
	for i, value := range values {
		if output, ok := s.cfg.Outputs[register+i]; ok {
			if output.Pwm != nil {
				output.Pin.DutyCycle(uint32(value), *output.Pwm.Cycle)
				mb.HoldingRegisters[register+i] = value
			} else {
				exception = &mbserver.IllegalDataAddress
				break
			}
		} else {
			exception = &mbserver.IllegalDataAddress
			break
		}
	}
	exception = &mbserver.Success
	data = frame.GetData()[0:4]

	return data, exception
}

func Uint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)*2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i*2:(i+1)*2], value)
	}
	return bytes
}

func bitAtPosition(value uint8, pos uint) uint8 {
	return (value >> pos) & 0x01
}