/**
 * Standard simple logging interface for Deskpass services. Ensures JSON logging
 * is used for all log messages on staging and production and also simplifies
 * setting up remote logging to Papertrail.
 */

package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"log/syslog"
	"os"
)

// Basic logger struct that will be returned by ConfigureLogger and contains
// raw local and remote loggers along with maps of functions for each log level
type Logger struct {
	localLogger     *zerolog.Logger
	remoteLogger    *zerolog.Logger
	localLoggerFns  map[string]func() *zerolog.Event
	remoteLoggerFns map[string]func() *zerolog.Event
}

var validLogLevels = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
}

// Define basic logging functions for each log level, all of which just call
// commonLog with the appropriate level
func (l *Logger) Debug(message string, meta map[string]interface{}) {
	l.commonLog(message, nil, &meta, "Debug")
}

func (l *Logger) Info(message string, meta map[string]interface{}) {
	l.commonLog(message, nil, &meta, "Info")
}

func (l *Logger) Warn(message string, meta map[string]interface{}) {
	l.commonLog(message, nil, &meta, "Warn")
}

func (l *Logger) Error(message string, err error, meta map[string]interface{}) {
	l.commonLog(message, err, &meta, "Err")
}

// Main logging function that all other logging functions call. This is where
// the actual logging happens, and it's also where the remote logging is
// handled if it's been set up.
func (l *Logger) commonLog(message string, err error, meta *map[string]interface{}, level string) {
	var localLogger *zerolog.Event

	// Create the local logger--this has to be handled slightly differently with
	// err level because of the error parameter
	if os.Getenv("REMOTE_LOG_ONLY") == "" {
		if level == "Err" {
			localLogger = l.localLogger.Err(err)
		} else {
			localLogger = l.localLoggerFns[level]()
		}

		// Tack on additional metadata if it's been provided
		if meta != nil {
			localLogger = localLogger.Dict("meta", buildDictFromMeta(meta, l))
		}

		// Log the message locally
		localLogger.Msg(message)
	}

	// Then deal with remote logging if it's been set up
	var remoteLogger *zerolog.Event

	if l.remoteLogger != nil {
		if level == "Err" {
			remoteLogger = l.remoteLogger.Err(err)
		} else {
			remoteLogger = l.remoteLoggerFns[level]()
		}

		if meta != nil {
			remoteLogger = remoteLogger.Dict("meta", buildDictFromMeta(meta, l))
		}

		// Log the message remotely
		remoteLogger.Msg(message)
	}
}

// Main function to set up the logger. appName is the name of the service, and
// remoteLoggerURL is the URL for the remote logger (if it's being used). The
// environment is used to determine whether to use JSON logging and to set up
// the remote logger. Will return a Logger struct that can be used to log
// messages.
func ConfigureLogger(appName string, remoteLoggerURL string, environment string) *Logger {
	// Set up local logger for starters
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro

	// Set logger level based on specific level env var or based on environment
	logLevel := "info"

	if os.Getenv("LOG_LEVEL") != "" {
		if _, exists := validLogLevels[os.Getenv("LOG_LEVEL")]; exists {
			logLevel = os.Getenv("LOG_LEVEL")
		} else {
			fmt.Println("Invalid log level specified:", os.Getenv("LOG_LEVEL"))
		}
	} else {
		if environment == "development" {
			logLevel = "debug"
		}
	}

	zerolog.SetGlobalLevel(validLogLevels[logLevel])

	// Then set up the local logger for printing to stdout
	localLogger := zerolog.New(os.Stdout).With().Timestamp().Str("app", appName).Logger()

	// Make logging output a bit easier to read in development
	if environment == "development" {
		localLogger = localLogger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	localLogger.Debug().Msg(fmt.Sprintf("Setting up logger for %s in %s environment", appName, environment))

	// Set up the combined local/remote logger that will be returned for use
	combinedLogger := Logger{
		localLogger:  &localLogger,
		remoteLogger: nil,
		localLoggerFns: map[string]func() *zerolog.Event{
			"Debug": localLogger.Debug,
			"Info":  localLogger.Info,
			"Warn":  localLogger.Warn,
		},
		remoteLoggerFns: map[string]func() *zerolog.Event{
			"Debug": nil,
			"Info":  nil,
			"Warn":  nil,
		},
	}

	if remoteLoggerURL != "" {
		// Configure the remote logger
		remoteLog, err := syslog.Dial(
			"udp",
			remoteLoggerURL,
			syslog.LOG_EMERG,
			appName+"-"+environment,
		)

		if err == nil {
			remoteLogger := zerolog.New(zerolog.SyslogLevelWriter(remoteLog)).With().Timestamp().Str("app", appName).Logger()

			if environment == "development" {
				remoteLogger = remoteLogger.Output(zerolog.ConsoleWriter{Out: remoteLog})
			}

			combinedLogger.remoteLogger = &remoteLogger

			// Save the remote logger functions into the logger struct so that they
			// can be referenced by string later
			combinedLogger.remoteLoggerFns["Debug"] = remoteLogger.Debug
			combinedLogger.remoteLoggerFns["Info"] = remoteLogger.Info
			combinedLogger.remoteLoggerFns["Warn"] = remoteLogger.Warn
		} else {
			localLogger.Err(err).Msg("Failed to set up remote logger!")
		}
	}

	// Configure the general logger
	return &combinedLogger
}

// Helper function to just flip through meta values and build a Dict from the
// values. This is used to add metadata to log messages.
func buildDictFromMeta(meta *map[string]interface{}, logger *Logger) *zerolog.Event {
	// Flip through meta fields, building Dict that can be passed to logger
	loggerDict := zerolog.Dict()

	for key, value := range *meta {
		loggerDict.Any(key, value)
	}

	return loggerDict
}
