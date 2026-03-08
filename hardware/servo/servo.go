// Package servo provides a high-level abstraction for hobby servos
// connected to a PCA9685 PWM board.
//
// This package knows about angles and physical limits.
// It does NOT know about I2C, ticks, or Gobot — that's pca9685's job.
package servo

import (
	"fmt"
	"time"

	"boot/hardware/pca9685" // our own driver — not Gobot directly
)

// PulseWriter is the interface this package depends on.
// It only demands one method: set a pulse on a channel.
//
// Why an interface instead of *pca9685.Driver directly?
// Because in tests you can pass a fake PulseWriter that doesn't
// need real hardware. The servo logic is testable without a Raspberry Pi.
type PulseWriter interface {
	SetPulse(channel int, onTick, offTick uint16) error
}

// Servo represents a single hobby servo motor attached to one
// PCA9685 channel.
type Servo struct {
	channel  int
	driver   PulseWriter
	minPulse float64 // pulse width in ms at 0°
	maxPulse float64 // pulse width in ms at 180°
	current  float64 // last angle commanded, in degrees
}

// Config holds the physical characteristics of one servo.
// These live in config.go and are passed in at construction time.
type Config struct {
	Channel  int
	MinPulse float64 // ms — pulse width that produces 0°
	MaxPulse float64 // ms — pulse width that produces 180°
}

// DefaultConfig returns safe starting values for a typical hobby servo.
// Tune MinPulse / MaxPulse per your specific servo if it doesn't
// reach full travel or goes past its mechanical limits.
func DefaultConfig(channel int) Config {
	return Config{
		Channel:  channel,
		MinPulse: 1.0, // 1.0 ms → 0°
		MaxPulse: 2.0, // 2.0 ms → 180°
	}
}

// New creates a Servo and immediately centres it (90°).
func New(driver PulseWriter, cfg Config) (*Servo, error) {
	s := &Servo{
		channel:  cfg.Channel,
		driver:   driver,
		minPulse: cfg.MinPulse,
		maxPulse: cfg.MaxPulse,
	}

	// Move to centre immediately so the servo is in a known position.
	if err := s.SetAngle(90); err != nil {
		return nil, fmt.Errorf("servo ch%d: failed to centre on init: %w", cfg.Channel, err)
	}

	return s, nil
}

// SetAngle moves the servo to the given angle (0–180 degrees).
func (s *Servo) SetAngle(degrees float64) error {
	if degrees < 0 || degrees > 180 {
		return fmt.Errorf("servo ch%d: angle %.1f out of range (0–180)", s.channel, degrees)
	}

	// Map degrees → pulse width in ms using linear interpolation.
	// At 0°   → minPulse ms
	// At 180° → maxPulse ms
	// At any angle in between → proportional value
	pulseMs := s.minPulse + (degrees/180.0)*(s.maxPulse-s.minPulse)

	// Convert ms to PCA9685 ticks using the helper in the pca9685 package.
	offTick := pca9685.MsToTick(pulseMs)

	if err := s.driver.SetPulse(s.channel, 0, offTick); err != nil {
		return fmt.Errorf("servo ch%d: SetAngle(%.1f°): %w", s.channel, degrees, err)
	}

	s.current = degrees
	return nil
}

// Center moves the servo to 90°.
func (s *Servo) Center() error {
	return s.SetAngle(90)
}

// Angle returns the last angle this servo was commanded to.
func (s *Servo) Angle() float64 {
	return s.current
}

// Sweep moves the servo back and forth between two angles, one step at a time,
// waiting `pause` between each step.
//
// This is a blocking call — it returns once the sweep is complete.
// For continuous motion in the background, call it inside a goroutine.
//
// Example: sweep from 0° to 180° in 5° increments, 20ms between steps:
//
//	s.Sweep(0, 180, 5, 20*time.Millisecond)
func (s *Servo) Sweep(from, to, step float64, pause time.Duration) error {
	if step <= 0 {
		return fmt.Errorf("servo ch%d: step must be > 0", s.channel)
	}

	// Move from → to
	for angle := from; angle <= to; angle += step {
		if err := s.SetAngle(angle); err != nil {
			return err
		}
		time.Sleep(pause)
	}

	// Move back to → from
	for angle := to; angle >= from; angle -= step {
		if err := s.SetAngle(angle); err != nil {
			return err
		}
		time.Sleep(pause)
	}

	return nil
}
