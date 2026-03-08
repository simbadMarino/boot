// Package config holds all hardware configuration for the robot.
//
// Keeping configuration here means:
//   - No magic numbers scattered across hardware/ or robot/ packages
//   - Easy to swap pin assignments, I2C addresses, or servo limits
//     without touching any logic
//   - Later you can load this from a YAML/JSON file instead of hardcoding
package config

import (
	"gobot.io/x/gobot/drivers/i2c"

	"boot/hardware/servo"
)

// Config is the top-level configuration for the entire robot.
type Config struct {
	// I2CAddress is the PCA9685 I2C address.
	// Default is 0x40. Change if you've soldered the address jumpers.
	I2CAddress int

	// I2CBus is the I2C bus number on your board.
	// Raspberry Pi uses bus 1 by default.
	I2CBus int

	// Servos maps a human-readable name to its physical configuration.
	// Use names that describe function: "pan", "tilt", "gripper", etc.
	Servos map[string]servo.Config
}

// Default returns a ready-to-use Config with sensible defaults.
// Edit this function to match your physical wiring.
func Default() *Config {
	return &Config{
		I2CAddress: 0x40,
		I2CBus:     1,
		Servos: map[string]servo.Config{
			// A pan/tilt head with two servos on channels 0 and 1.
			// Adjust MinPulse/MaxPulse if your servos don't reach full travel.
			"pan": {
				Channel:  15,
				MinPulse: 1.0,
				MaxPulse: 2.0,
			},
			"tilt": {
				Channel:  14,
				MinPulse: 1.0,
				MaxPulse: 2.0,
			},
		},
	}
}

// PCA9685Options converts config values into Gobot i2c option functions.
// robot.go calls this when initialising the PCA9685 driver.
//
// This is the only place where the config package touches Gobot types —
// everything else in config stays framework-agnostic.
func (c *Config) PCA9685Options() []func(i2c.Config) {
	return []func(i2c.Config){
		i2c.WithBus(c.I2CBus),
		i2c.WithAddress(c.I2CAddress),
	}
}
