// Package pca9685 wraps the Gobot PCA9685 I2C driver and exposes a clean,
// minimal interface that the rest of the project can depend on.
//
// Nothing outside this package needs to know about Gobot internals —
// if you ever swap Gobot for a different I2C library, only this file changes.
package pca9685

import (
	"fmt"

	"gobot.io/x/gobot/drivers/i2c"
)

// PWMFrequency is the standard frequency for hobby servos.
// 50 Hz means one pulse every 20 ms, which is what virtually all
// hobby servos expect.
const PWMFrequency = 50.0

// ticksPerCycle is the PCA9685's 12-bit resolution.
// Every 20 ms period is divided into 4096 steps (ticks).
const ticksPerCycle = 4096

// PeriodMs is the duration of one PWM cycle at 50 Hz (1000ms / 50Hz).
const PeriodMs = 1000.0 / PWMFrequency

// Driver wraps the Gobot PCA9685Driver and is the only type
// the rest of the project interacts with for low-level PWM control.
type Driver struct {
	dev *i2c.PCA9685Driver
}

// New creates a new Driver.
//
// adaptor is the platform-specific I2C connector (e.g. raspi.NewAdaptor()).
// It accepts optional Gobot i2c config functions, for example:
//
//	pca9685.New(adaptor, i2c.WithAddress(0x41))
//
// The PCA9685 default I2C address is 0x40. You only need WithAddress if
// you've changed the address jumpers on the board.
func New(adaptor i2c.Connector, opts ...func(i2c.Config)) (*Driver, error) {
	dev := i2c.NewPCA9685Driver(adaptor, opts...)

	// Start() initialises the I2C connection and configures the chip.
	// In Gobot, every driver must be Start()ed before use.
	if err := dev.Start(); err != nil {
		return nil, fmt.Errorf("pca9685: failed to start: %w", err)
	}

	d := &Driver{dev: dev}

	// Set PWM frequency immediately after starting.
	// All channels share the same frequency on the PCA9685.
	if err := d.SetFrequency(PWMFrequency); err != nil {
		return nil, err
	}

	return d, nil
}

// SetFrequency sets the PWM frequency in Hz for all channels.
// Call this once at startup. Most servo projects use 50 Hz.
func (d *Driver) SetFrequency(hz float32) error {
	if err := d.dev.SetPWMFreq(hz); err != nil {
		return fmt.Errorf("pca9685: SetFrequency(%v): %w", hz, err)
	}
	return nil
}

// SetPulse sets the PWM signal on a specific channel.
//
//   - channel: 0–15 (the PCA9685 has 16 channels)
//   - onTick:  the tick at which the pulse goes HIGH  (almost always 0)
//   - offTick: the tick at which the pulse goes LOW
//
// To convert a pulse width in milliseconds to an offTick:
//
//	offTick = uint16((pulseMs / PeriodMs) * ticksPerCycle)
//
// Example: a 1.5 ms pulse at 50 Hz → (1.5 / 20) * 4096 ≈ 307
func (d *Driver) SetPulse(channel int, onTick, offTick uint16) error {
	if channel < 0 || channel > 15 {
		return fmt.Errorf("pca9685: channel %d out of range (0–15)", channel)
	}
	if err := d.dev.SetPWM(channel, onTick, offTick); err != nil {
		return fmt.Errorf("pca9685: SetPulse(ch=%d): %w", channel, err)
	}
	return nil
}

// MsToTick converts a pulse width in milliseconds to a PCA9685 tick value.
// This is a convenience helper used by the servo package.
//
// Example:
//
//	MsToTick(1.0)  → 205   (≈ 0°)
//	MsToTick(1.5)  → 307   (≈ 90°)
//	MsToTick(2.0)  → 410   (≈ 180°)
func MsToTick(ms float64) uint16 {
	return uint16((ms / PeriodMs) * ticksPerCycle)
}

// Close stops all PWM output and releases the I2C device.
// Always call this when shutting down.
func (d *Driver) Close() error {
	if err := d.dev.Halt(); err != nil {
		return fmt.Errorf("pca9685: Close: %w", err)
	}
	return nil
}
