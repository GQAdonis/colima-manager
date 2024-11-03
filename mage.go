//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Default target to run when none is specified
var Default = Build

// Build builds the binary
func Build() error {
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", getBinaryName(), ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InstallDeps installs project dependencies
func InstallDeps() error {
	fmt.Println("Installing dependencies...")
	cmd := exec.Command("go", "mod", "download")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install builds and installs the binary to /usr/local/bin
func Install() error {
	// First build the binary
	if err := Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	srcPath := getBinaryName()
	destPath := filepath.Join("/usr/local/bin", getBinaryName())

	fmt.Printf("Installing %s to %s...\n", srcPath, destPath)

	// Create /usr/local/bin if it doesn't exist
	if err := os.MkdirAll("/usr/local/bin", 0755); err != nil {
		return fmt.Errorf("failed to create /usr/local/bin: %w", err)
	}

	// Copy the binary
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write binary to %s: %w", destPath, err)
	}

	fmt.Printf("Successfully installed %s\n", destPath)
	return nil
}

// Clean removes the binary
func Clean() error {
	fmt.Println("Cleaning...")
	if err := os.Remove(getBinaryName()); err != nil && !os.IsNotExist(err) {
		return err
	}
	// Also try to remove from /usr/local/bin
	binPath := filepath.Join("/usr/local/bin", getBinaryName())
	if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Test runs the test suite
func Test() error {
	fmt.Println("Running tests...")
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TestCoverage runs tests with coverage and generates a coverage report
func TestCoverage() error {
	fmt.Println("Running tests with coverage...")

	// Create coverage output directory if it doesn't exist
	if err := os.MkdirAll("coverage", 0755); err != nil {
		return fmt.Errorf("failed to create coverage directory: %w", err)
	}

	// Run tests with coverage
	cmd := exec.Command("go", "test", "./...", "-coverprofile=coverage/coverage.out")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Generate HTML coverage report
	cmd = exec.Command("go", "tool", "cover", "-html=coverage/coverage.out", "-o=coverage/coverage.html")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("Coverage report generated at coverage/coverage.html")
	return nil
}

// Helpers
func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "colima-manager.exe"
	}
	return "colima-manager"
}
