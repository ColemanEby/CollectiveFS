package logger

import (
	"log"
	"os"
	"time"
)

var (
	verbose bool
)

// Init initializes logging with the specified verbosity level.
func Init(v bool) {
	verbose = v
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
}

// Debug logs a debug message if verbose mode is enabled.
func Debug(format string, v ...interface{}) {
	if verbose {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message.
func Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

// Error logs an error message.
func Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// Fatal logs a fatal error and exits.
func Fatal(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}

// SetOutput sets the output destination for the logger.
func SetOutput(w *os.File) {
	log.SetOutput(w)
}

// GetTimestamp returns a formatted timestamp matching Python's format.
func GetTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
