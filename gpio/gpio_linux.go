package gpio

import "github.com/stianeikeland/go-rpio"

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

	PullOff = rpio.PullOff
	PullDown = rpio.PullDown
	PullUp = rpio.PullUp

	Low = rpio.Low
	High = rpio.High

)

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
	rpio.PinMode(rpio.Pin(pin), mode)
}

func WritePin(pin Pin, state rpio.State) {
	rpio.WritePin(rpio.Pin(pin), state)
}

func ReadPin(pin Pin) rpio.State {
	return rpio.State(rpio.ReadPin(rpio.Pin(pin)))
}

func TogglePin(pin Pin) {
	rpio.TogglePin(rpio.Pin(pin))
}

func DetectEdge(pin Pin, edge rpio.Edge) {
	rpio.DetectEdge(rpio.Pin(pin), edge)
}

func EdgeDetected(pin Pin) bool {
	return rpio.EdgeDetected(rpio.Pin(pin))
}

func PullMode(pin Pin, pull rpio.Pull) {
	rpio.PullMode(rpio.Pin(pin), pull)
}

func SetFreq(pin Pin, freq int) {
	rpio.SetFreq(rpio.Pin(pin), freq)
}

func SetDutyCycle(pin Pin, dutyLen, cycleLen uint32) {
	rpio.SetDutyCycle(rpio.Pin(pin), dutyLen, cycleLen)
}

func StopPwm() {
	rpio.StopPwm()
}

func StartPwm() {
	rpio.StartPwm()
}

func EnableIRQs(irqs uint64) {
	rpio.EnableIRQs(irqs)
}

func DisableIRQs(irqs uint64) {
	rpio.DisableIRQs(irqs)
}

func Open() (err error) {
	return rpio.Open()
}

func Close() error {
	return rpio.Close()
}
