package tinyio

import (
	"io"
)

// UART represents a UART connection. It is implemented by the machine.UART type.
type UART interface {
	io.Reader
	io.Writer

	Buffered() int
}

// PWM represents a PWM peripheral. A PWM have many output channels
// but all are constrained to the same underlying frequency and Top set on the PWM.
type PWM interface {
	// SetPeriod sets the amount of time between the PWM's square wave rising flank.
	// period is in nanoseconds, just like time.Duration.
	SetPeriod(period int64) error

	Top() uint32
	// Set sets the PWM's channel to value. One can use the value returned by
	// Top to obtain a certain duty cycle. i.e:
	//  pwm.Set(0, pwm.Top()/4) // Sets channel 0 to 25% duty cycle.
	//  pwm.Set(1, 2*pwm.Top()/3) // Sets channel 1 to 66.67% duty cycle.
	// Set should only return an error on I/O errors such as with peripherals
	// external to a microcontroller or when channel exceeds amount of available channels.
	Set(channel uint8, value uint32) error
}

// SPI represents a SPI bus. It is implemented by the machine.SPI type.
type SPI interface {
	// Tx transmits the given buffer w and receives at the same time the buffer r.
	// The two buffers must be the same length. The only exception is when w or r are nil,
	// in which case Tx only transmits (without receiving) or only receives (while sending 0 bytes).
	Tx(w, r []byte) error

	// Transfer writes a single byte out on the SPI bus and receives a byte at the same time.
	// If you want to transfer multiple bytes, it is more efficient to use Tx instead.
	Transfer(b byte) (byte, error)
}

// I2C represents an I2C bus. It is notably implemented by the machine.I2C type.
type I2C interface {
	// Excludes WriteRegister and ReadRegister since these are rarely implemented
	// as hardware-level functions and more commonly use the contents of
	// machine/i2c.go. They should instead be implemented as tinyio top level
	// package functions or subpackage functions.

	Tx(addr uint16, w, r []byte) error
}
