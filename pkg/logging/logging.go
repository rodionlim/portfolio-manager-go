package logging

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
	verbose bool
	logFile *os.File
}

// InitializeLogger initializes the logger with the specified log level and log file path
func InitializeLogger(verboseLogging bool, logFilePath string) (*Logger, error) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	l := &Logger{
		verbose: verboseLogging,
	}

	if logFilePath != "" {
		// Open the log file for writing
		var err error
		l.logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		// Tee the log output to both stdout and the log file
		log.SetOutput(io.MultiWriter(os.Stdout, l.logFile))
	} else {
		// Log only to stdout
		log.SetOutput(os.Stdout)
	}

	return l, nil
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	if l.verbose {
		log.Output(2, fmt.Sprintln(v...))
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
}

// Warn logs a warning message
func (l *Logger) Warn(v ...interface{}) {
	log.Output(2, fmt.Sprintf("WARN: %s", fmt.Sprintln(v...)))
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	log.Output(2, fmt.Sprintf("ERROR: %s", fmt.Sprintln(v...)))
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
