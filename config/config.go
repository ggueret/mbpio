package config

import (
	"os"
	"fmt"
	"time"
	"gopkg.in/yaml.v2"
	"github.com/ggueret/mbpio/gpio"
)

type InputPoller struct {
	Type			string
	Value			*string
}

type Input struct {
	Pin				gpio.Pin
	Poller			*InputPoller
}

type OutputPwm struct {
	Freq			*int	`yaml:",omitempty"`
	Cycle			*uint32	`yaml:",omitempty"`
}

type Output struct {
	Pin				gpio.Pin
	Pwm				*OutputPwm	`yaml:",omitempty"`
}

type Config struct {
	Inputs			map[int]Input	`yaml:",flow"`
	Outputs			map[int]Output	`yaml:",flow"`

	ListenOn		string	`yaml:"listen_on"`

	PollEvery		int

	EnableRTU		bool
	RTUAddress		string
	RTUBaudRate		int
	RTUDataBits		int
	RTUStopBits		int
	RTUParity		string
	RTUTimeout		time.Duration
	// todo: RS485 config
}

func Load(path string) (config *Config, err error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	// load default settings
	config = &Config{
		ListenOn: "127.0.0.1:502",
		EnableRTU: false,
		RTUAddress: "/dev/ttyS0",
		RTUBaudRate: 19200,
		RTUDataBits: 8,
		RTUStopBits: 1,
		RTUParity: "E",
		RTUTimeout: 0,
	}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("file decoding errored: %s", err)
	}

	return config, nil
}