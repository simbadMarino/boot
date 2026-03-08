// Package robot owns and orchestrates all hardware components.
//
// It is the single place where everything is wired together.
// main.go creates a Robot, calls Start(), and the robot takes it from there.
//
// Adding a new hardware component (sensor, motor, LED) means:
//  1. Add it to the Robot struct
//  2. Initialise it in New()
//  3. Clean it up in Stop()
//  4. Expose whatever methods make sense for your use case
//
// Nothing in this package imports Gobot or hardware-specific libraries —
// it only imports our own hardware/ packages.
package robot

import (
	"fmt"
	"time"

	"boot/config"
	"boot/hardware/pca9685"
	"boot/hardware/servo"

	// Platform adaptor — swap this single import for your board:
	//   Jetson Nano:  "gobot.io/x/gobot/platforms/jetson"
	//   BeagleBone:   "gobot.io/x/gobot/platforms/beaglebone"
	"gobot.io/x/gobot/platforms/raspi"
)

// Robot holds all hardware references.
// It is created once in main.go and lives for the duration of the program.
type Robot struct {
	cfg    *config.Config
	pca    *pca9685.Driver
	servos map[string]*servo.Servo
}

// New initialises all hardware based on the provided Config.
//
// It returns an error if any hardware fails to initialise, so main.go
// can handle the failure cleanly before the robot starts doing anything.
func New(cfg *config.Config) (*Robot, error) {
	// 1. Create the platform adaptor.
	//    This is the only place that knows what physical board we're running on.
	adaptor := raspi.NewAdaptor()

	// 2. Start the PCA9685 driver.
	//    Pass optional I2C options from config (e.g. custom address).
	pca, err := pca9685.New(adaptor, cfg.PCA9685Options()...)
	if err != nil {
		return nil, fmt.Errorf("robot: failed to init PCA9685: %w", err)
	}

	// 3. Initialise all servos defined in config.
	servos := make(map[string]*servo.Servo)
	for name, scfg := range cfg.Servos {
		s, err := servo.New(pca, scfg)
		if err != nil {
			// If any servo fails, close the pca9685 cleanly before returning.
			_ = pca.Close()
			return nil, fmt.Errorf("robot: failed to init servo %q: %w", name, err)
		}
		servos[name] = s
		fmt.Printf("robot: servo %q ready on channel %d\n", name, scfg.Channel)
	}

	return &Robot{
		cfg:    cfg,
		pca:    pca,
		servos: servos,
	}, nil
}

// Servo returns a servo by name (as defined in config).
// Returns nil if the name doesn't exist — callers should check.
//
// Example:
//
//	r.Servo("pan").SetAngle(45)
func (r *Robot) Servo(name string) *servo.Servo {
	return r.servos[name]
}

// DemoSweep runs a simple demo that sweeps all servos back and forth.
// This is a blocking call — run it in a goroutine if you need it in the background.
func (r *Robot) DemoSweep() error {
	fmt.Println("robot: starting sweep demo...")

	for name, s := range r.servos {
		fmt.Printf("robot: sweeping servo %q\n", name)
		if err := s.Sweep(0, 180, 5, 20*time.Millisecond); err != nil {
			return fmt.Errorf("robot: sweep failed for servo %q: %w", name, err)
		}
	}

	return nil
}

// Stop gracefully shuts down all hardware.
// Always call this before your program exits — either in main() or
// via a signal handler (see main.go for the pattern).
func (r *Robot) Stop() error {
	fmt.Println("robot: shutting down...")

	// Centre all servos before cutting power — avoids sudden jumps on next start.
	for name, s := range r.servos {
		if err := s.Center(); err != nil {
			fmt.Printf("robot: warning — could not centre servo %q: %v\n", name, err)
		}
	}

	if err := r.pca.Close(); err != nil {
		return fmt.Errorf("robot: error closing PCA9685: %w", err)
	}

	fmt.Println("robot: shutdown complete")
	return nil
}
