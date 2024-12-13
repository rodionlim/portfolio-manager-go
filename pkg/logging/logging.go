package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Logger struct {
	verbose bool
	logFile *os.File
}

var (
	instance *Logger
	once     sync.Once
)

// InitializeLogger initializes the logger with the specified log level and log file path
func InitializeLogger(verboseLogging bool, logFilePath string) (*Logger, error) {
	var err error
	once.Do(func() {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

		instance = &Logger{
			verbose: verboseLogging,
		}

		if logFilePath != "" {
			// Open the log file for writing
			instance.logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				instance = nil
				return
			}

			// Tee the log output to both stdout and the log file
			log.SetOutput(io.MultiWriter(os.Stdout, instance.logFile))
		} else {
			// Log only to stdout
			log.SetOutput(os.Stdout)
		}
	})
	return instance, err
}

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	if instance == nil {
		instance, _ = InitializeLogger(false, "")
		instance.Warn("Logger not initialized. Using default logger.")
	}

	return instance
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	if l.verbose {
		log.Output(2, fmt.Sprintf("DEBUG: %s", fmt.Sprintln(v...)))
	}
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.verbose {
		log.Output(2, fmt.Sprintf("DEBUG: "+format, v...))
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	log.Output(2, fmt.Sprintf("INFO: %s", fmt.Sprintln(v...)))
}

// Infof logs an info message with formatting
func (l *Logger) Infof(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf("INFO: "+format, v...))
}

// Warn logs a warning message
func (l *Logger) Warn(v ...interface{}) {
	log.Output(2, fmt.Sprintf("WARN: %s", fmt.Sprintln(v...)))
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf("WARN: "+format, v...))
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	log.Output(2, fmt.Sprintf("ERROR: %s", fmt.Sprintln(v...)))
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf("ERROR: "+format, v...))
}

// Fatalf logs a fatal error message and exits the application
func (l *Logger) Fatalf(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf("FATAL: "+format, v...))
	os.Exit(1)
}

// CloseLogger closes the log file
func (l *Logger) CloseLogger() error {
	if l.logFile != nil {
		err := l.logFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
