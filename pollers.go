package main

import (
	"time"
	"errors"
//	"github.com/d2r2/go-dht"
//	"github.com/ggueret/mbpio/config"
	"github.com/ggueret/mbpio/gpio"
	log "github.com/sirupsen/logrus"
	"github.com/stianeikeland/go-rpio"
)

var DHTMaxCount = 32000
var TimeoutError = errors.New("Timeout")

func (s *Server) RegisterPollers() {
	var pollers = map[string]func([]int){
		"DHT22": s.PollDHT22,
		"LDR": s.PollLDR,
		"PB": s.PollPB,
	}

	for name, function := range pollers {
		s.pollers[name] = function
	}
}


func (s *Server) PollPB(inputs []int) {
	s.wg.Add(1)
	defer s.wg.Done()

	doPoll := func() {

		for _, addr := range inputs {
			input := s.cfg.Inputs[addr]
			state := input.Pin.Read()

			s.mu.Lock()
			s.mb.DiscreteInputs[addr] = uint8(state)
			s.mu.Unlock()
		}
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()


	doPoll()
	for {
		select {
			case <- ticker.C:
				doPoll()
			case <- s.quit:
				log.Info("PB poller terminated.")
				return
			default:
		}
	}
}

func (s *Server) PollLDR(inputs []int) {
	s.wg.Add(1)
	defer s.wg.Done()

	doPoll := func() {

		for _, addr := range inputs {
			input := s.cfg.Inputs[addr]

			input.Pin.Input()
			if input.Pin.Read() == gpio.Low {
				continue
			}

			input.Pin.Output()
			input.Pin.Low()
			time.Sleep(100 * time.Millisecond)
			input.Pin.Input()

			count := uint16(0)
			for input.Pin.Read() == gpio.Low {
				count++
				if count > 65535 {
					log.WithFields(log.Fields{"addr": addr, "pin": input.Pin}).Warning("LDR poller: timeout reached")
					break
				}
			}
			if count <= 65535 {
				s.mu.Lock()
				s.mb.InputRegisters[addr] = count
				s.mu.Unlock()
				log.WithFields(log.Fields{"addr": addr, "pin": input.Pin, "value": count}).Trace("LDR poller: value refreshed.")
			}
		}
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()


	doPoll()
	for {
		select {
			case <- ticker.C:
				doPoll()
			case <- s.quit:
				log.Info("LDR poller terminated.")
				return
			default:
		}
	}
}

func (s *Server) PollDHT22(inputs []int) {
	s.wg.Add(1)
	defer s.wg.Done()

	doPoll := func() {
		for _, addr := range inputs {
			input := s.cfg.Inputs[addr]

			lengths := make([]time.Duration, 40)
			iteration := 0

			// Send init
			input.Pin.Output()

			input.Pin.High()
			time.Sleep(250 * time.Millisecond)

			input.Pin.Low()
			time.Sleep(5 * time.Millisecond)

			input.Pin.High()
			time.Sleep(20 * time.Microsecond)

			input.Pin.Input()

			for {
				log.Print(input.Pin.Read())
				iteration++
				if iteration >= 80 {
					break
				}

			}
			continue
			for {
				duration, err := TimePulse(&input.Pin, gpio.High)
				if err != nil {
					log.WithFields(log.Fields{"addr": addr, "pin": input.Pin}).Warning("DHT22 poller: timeout reached")
					break
				}
				lengths[iteration] = duration
				iteration++
				if iteration >= 40 {
					break
				}
			}
			break
		}
	}

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	doPoll()
	for {
		select {
			case <- ticker.C:
				doPoll()
			case <- s.quit:
				log.Info("DHT22 poller terminated.")
				return
			default:
		}
	}
}


func TimePulse(pin *gpio.Pin, state rpio.State) (time.Duration, error) {
	aroundState := gpio.Low
	if state == gpio.Low {
		aroundState = gpio.High
	}
	cnt := 0
	for {
		if pin.Read() == aroundState {
			break
		}
		cnt++
		if cnt >= DHTMaxCount {
			return time.Duration(0), TimeoutError
		}
	}

	cnt = 0
	for {
		if pin.Read() == state {
			break
		}
		cnt++
		if cnt >= DHTMaxCount {
			return time.Duration(0), TimeoutError
		}
	}

	startTime := time.Now()

	cnt = 0
	for {
		if pin.Read() == aroundState {
			break
		}
		cnt++
		if cnt >= DHTMaxCount {
			return time.Duration(0), TimeoutError
		}
	}

	return time.Since(startTime), nil
}