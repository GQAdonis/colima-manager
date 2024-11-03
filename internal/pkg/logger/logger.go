package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

var (
	defaultLogger *Logger
	logFile       *os.File
)

func init() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create or open log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join("logs", fmt.Sprintf("colima-manager-%s.log", timestamp))
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create multi-writer for both file and stdout
	multiWriter := MultiWriter{writers: []Writer{
		&FileWriter{file: logFile},
		&ConsoleWriter{},
	}}

	// Initialize default logger
	defaultLogger = &Logger{
		infoLogger:  log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime),
		debugLogger: log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime),
	}
}

// Writer interface for different output destinations
type Writer interface {
	Write(p []byte) (n int, err error)
}

// MultiWriter writes to multiple writers
type MultiWriter struct {
	writers []Writer
}

func (mw MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}

// FileWriter writes to a file
type FileWriter struct {
	file *os.File
}

func (fw *FileWriter) Write(p []byte) (n int, err error) {
	return fw.file.Write(p)
}

// ConsoleWriter writes to stdout
type ConsoleWriter struct{}

func (cw *ConsoleWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	return defaultLogger
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Helper function to get file and line number
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Info logs an info message with caller information
func (l *Logger) Info(format string, v ...interface{}) {
	caller := getCallerInfo()
	l.infoLogger.Printf("%s - "+format, append([]interface{}{caller}, v...)...)
}

// Error logs an error message with caller information
func (l *Logger) Error(format string, v ...interface{}) {
	caller := getCallerInfo()
	l.errorLogger.Printf("%s - "+format, append([]interface{}{caller}, v...)...)
}

// Debug logs a debug message with caller information
func (l *Logger) Debug(format string, v ...interface{}) {
	caller := getCallerInfo()
	l.debugLogger.Printf("%s - "+format, append([]interface{}{caller}, v...)...)
}

// LogError logs an error and returns it
func (l *Logger) LogError(err error, format string, v ...interface{}) error {
	if err != nil {
		caller := getCallerInfo()
		msg := fmt.Sprintf(format, v...)
		l.errorLogger.Printf("%s - %s: %v", caller, msg, err)
	}
	return err
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	caller := getCallerInfo()
	l.errorLogger.Printf("%s - FATAL: "+format, append([]interface{}{caller}, v...)...)
	os.Exit(1)
}
