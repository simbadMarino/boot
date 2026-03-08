// main.go is intentionally thin.
//
// Its only responsibilities are:
//  1. Load config
//  2. Create the robot
//  3. Run the robot
//  4. Handle OS signals for clean shutdown
//
// No hardware logic lives here. If you find yourself writing
// servo or I2C code in main.go, it belongs in robot/ or hardware/ instead.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"boot/config"
	"boot/robot"
)

var version = "0.0.1"

func main() {
	// --version flag
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	// ── 1. Load config ────────────────────────────────────────────────────────
	// config.Default() returns hardcoded values.
	// Later you could replace this with config.FromFile("robot.yaml") etc.
	cfg := config.Default()

	// ── 2. Create the robot ───────────────────────────────────────────────────
	// robot.New() initialises all hardware defined in cfg.
	// If any component fails (I2C not available, wrong address, etc.)
	// we get a clear error here, before anything moves.
	r, err := robot.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialise robot: %v\n", err)
		os.Exit(1)
	}

	// ── 3. Ensure clean shutdown on Ctrl+C or SIGTERM ─────────────────────────
	// This is the standard Go pattern for graceful shutdown.
	// When the OS sends a termination signal (Ctrl+C, kill, systemd stop, etc.)
	// we catch it and call r.Stop() so servos centre and I2C is released cleanly.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit // block until a signal arrives
		fmt.Println("\nmain: received shutdown signal")
		if err := r.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "main: error during shutdown: %v\n", err)
		}
		os.Exit(0)
	}()

	// ── 4. Run the robot ──────────────────────────────────────────────────────
	// Everything from here is your robot's actual behaviour.
	// Swap DemoSweep() for your own logic.

	fmt.Println("main: robot ready")

	// Example A — run a sweep demo (blocking)
	if err := r.DemoSweep(); err != nil {
		fmt.Fprintf(os.Stderr, "main: demo error: %v\n", err)
		_ = r.Stop()
		os.Exit(1)
	}

	// Example B — move individual servos by name
	if pan := r.Servo("pan"); pan != nil {
		_ = pan.SetAngle(45)
	}
	if tilt := r.Servo("tilt"); tilt != nil {
		_ = tilt.SetAngle(90)
	}

	// Block forever (replace with your control loop, HTTP server, etc.)
	select {}
}
