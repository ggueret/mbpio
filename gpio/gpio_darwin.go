package gpio

import (
	"github.com/stianeikeland/go-rpio"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type Mode rpio.Mode
type Pin rpio.Pin
//type State rpio.State
type Pull rpio.Pull
type Edge rpio.Edge


func State(i uint16) rpio.State {
	return rpio.State(i)
}

const (
	Input = rpio.Input
	Output = rpio.Output
	Clock = rpio.Clock
	Pwm = rpio.Pwm
	Spi = rpio.Spi

	Low = rpio.Low
	High = rpio.High

	PullOff = rpio.PullOff
	PullDown = rpio.PullDown
	PullUp = rpio.PullUp

	NoEdge = rpio.NoEdge
	RiseEdge = rpio.RiseEdge
	FallEdge = rpio.FallEdge
	AnyEdge = rpio.AnyEdge
)

var ModeStrings = map[rpio.Mode]string {
	rpio.Input: "Input",
	rpio.Output: "Output",
	rpio.Clock: "Clock",
	rpio.Pwm: "Pwm",
	rpio.Spi: "Spi",
}

var PullStrings = map[rpio.Pull]string {
	rpio.PullOff: "Off",
	rpio.PullDown: "Down",
	rpio.PullUp: "Up",
}

var StateStrings = map[rpio.State]string {
	rpio.Low: "Low",
	rpio.High: "High",
}

var EdgeStrings = map[rpio.Edge]string {
	rpio.NoEdge: "NoEdge",
	rpio.RiseEdge: "RiseEdge",
	rpio.FallEdge: "FallEdge",
	rpio.AnyEdge: "AnyEdge",
}

func (pin Pin) Input() {
	PinMode(pin, Input)
}

func (pin Pin) Output() {
	PinMode(pin, Output)
}

func (pin Pin) Clock() {
	PinMode(pin, Clock)
}

func (pin Pin) Pwm() {
	PinMode(pin, Pwm)
}

func (pin Pin) High() {
	WritePin(pin, High)
}

func (pin Pin) Low() {
	WritePin(pin, Low)
}

func (pin Pin) Toggle() {
	TogglePin(pin)
}

func (pin Pin) Freq(freq int) {
	SetFreq(pin, freq)
}

func (pin Pin) DutyCycle(dutyLen, cycleLen uint32) {
	SetDutyCycle(pin, dutyLen, cycleLen)
}

func (pin Pin) Mode(mode rpio.Mode) {
	PinMode(pin, mode)
}

func (pin Pin) Write(state rpio.State) {
	WritePin(pin, state)
}

func (pin Pin) Read() rpio.State {
	return ReadPin(pin)
}

func (pin Pin) Pull(pull rpio.Pull) {
	PullMode(pin, pull)
}

func (pin Pin) PullUp() {
	PullMode(pin, PullUp)
}

func (pin Pin) PullDown() {
	PullMode(pin, PullDown)
}

func (pin Pin) PullOff() {
	PullMode(pin, PullOff)
}

func (pin Pin) Detect(edge rpio.Edge) {
	DetectEdge(pin, edge)
}

func (pin Pin) EdgeDetected() bool {
	return EdgeDetected(pin)
}

func PinMode(pin Pin, mode rpio.Mode) {
	log.WithFields(logrus.Fields{"pin": pin, "mode": ModeStrings[mode]}).Debug("PinMode")
//	log.Debugf("GPIO - Set Pin Mode %s on Pin %d", ModeStrings[mode], pin)
}

func WritePin(pin Pin, state rpio.State) {
	log.WithFields(logrus.Fields{"pin": pin, "state": StateStrings[state]}).Debug("WritePin")
//	log.Debugf("GPIO - Write %s on Pin %d", StateStrings[state], pin)
}

func ReadPin(pin Pin) rpio.State {
	log.WithFields(logrus.Fields{"pin": pin}).Debug("ReadPin")
//	log.Debugf("GPIO - Read on Pin %d", pin)
	return Low
}

func TogglePin(pin Pin) {
	log.WithFields(logrus.Fields{"pin": pin}).Debug("TogglePin")
//	log.Debugf("GPIO - Toggle Pin %d", pin)
}

func DetectEdge(pin Pin, edge rpio.Edge) {
	log.WithFields(logrus.Fields{"pin": pin, "edge": EdgeStrings[edge]}).Debug("DetectEdge")
//	log.Debugf("GPIO - Detect Edge %s on Pin %d", EdgeStrings[edge], pin)
}

func EdgeDetected(pin Pin) bool {
	log.WithFields(logrus.Fields{"pin": pin}).Debug("EdgeDetected")
//	log.Debug("GPIO - Edge Detected on Pin %d", pin)
	return false
}

func PullMode(pin Pin, pull rpio.Pull) {
	log.WithFields(logrus.Fields{"pin": pin, "pull": PullStrings[pull]}).Debug("PullMode")
//	log.Debugf("GPIO - Set Pull Mode %s on Pin %d", PullStrings[pull], pin)
}

func SetFreq(pin Pin, freq int) {
	log.WithFields(logrus.Fields{"pin": pin, "freq": freq}).Debug("SetFreq")
//	log.Debugf("GPIO - Set Freq of %d Hz on Pin %d", freq, pin)
}

func SetDutyCycle(pin Pin, dutyLen, cycleLen uint32) {
	log.WithFields(logrus.Fields{"pin": pin, "duty": dutyLen, "cycle": cycleLen}).Debug("SetDutyCycle")
//	log.Debugf("GPIO - Set DutyCycle of %d/%d on Pin %d", dutyLen, cycleLen, pin)
}

func StopPwm() {
	log.Debug("StopPwm")
//	log.Debug("GPIO - Stop Pwm")
}

func StartPwm() {
	log.Debug("StartPwm")
//	log.Debug("GPIO - Start Pwm")
}

func EnableIRQs(irqs uint64) {
	log.Debug("GPIO - Enable IRQs")
}

func DisableIRQs(irqs uint64) {
	log.Debug("GPIO - Disable IRQs")
}

func Open() (err error) {
	log.Debug("Open")
	return nil
}

func Close() error {
	log.Debug("Close")
	return nil
}
