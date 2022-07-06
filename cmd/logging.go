package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output logging file should be located at /var/logging/service-xyz/service-xyz.logging and
// will be rolled accocdrding to configuration set.
func initLogging(options rootOptions) {
	var writers []io.Writer

	if options.LogConsole {
		if options.LogFormat.Value == "json" {
			writers = append(writers, os.Stdout)
		} else {
			// console := config.console
			// if console == nil {
			console := os.Stderr
			// }
			writers = append(writers, zerolog.ConsoleWriter{FormatLevel: formatLevel(options.Color.Value), Out: console, NoColor: options.Color.Value})
		}
	}

	if !options.NoLogFile {
		writers = append(writers, newRollingFile(options))
	}
	mw := io.MultiWriter(writers...)

	// Default level is info, unless verbose flag is on
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	logger := zerolog.New(mw).With().Timestamp().Str("instance_id", config.InstanceId).Logger()
	// override global logger
	log.Logger = logger
	level, err := zerolog.ParseLevel(options.LogLevel.Value)
	if err != nil {
		panic(err)
	}
	logger.Level(level)

	logger.Info().
		Bool("no-log-file", options.NoLogFile).
		Str("log-format", options.LogFormat.Value).
		Bool("log-console", options.LogConsole).
		Str("log-level", options.LogLevel.Value).
		Str("log-directory", options.LogDirectory).
		Str("log-filename", options.LogFilename).
		Int("log-file-max-size", options.LogFileMaxSize).
		Int("log-file-max-backups", options.LogFileMaxBackups).
		Int("log-file-max-age", options.LogFileMaxAge).
		Msg("logging configured")

	return logger
}

func newRollingFile(options rootOptions) io.Writer {
	if err := os.MkdirAll(options.LogDirectory, 0744); err != nil {
		log.Error().Err(err).Str("path", options.LogDirectory).Msg("can't create logging directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(options.LogDirectory, options.LogFilename),
		MaxBackups: options.LogFileMaxAge,  // files
		MaxSize:    options.LogFileMaxSize, // megabytes
		MaxAge:     options.LogFileMaxAge,  // days
	}
}

func formatLevel(noColor bool) func(i interface{}) string {
	// formatLevel is zerolog.Formatter that turns a level value into a string.
	return func(i interface{}) string {
		if level, ok := i.(string); ok {
			switch level {
			case "trace":
				return ui.Colorize(ui.ColorTrace, noColor, "TRC")
			case "debug":
				return ui.Colorize(ui.ColorDebug, noColor, "DBG")
			case "info":
				return ui.Colorize(ui.ColorInfo, noColor, "INF")
			case "warn":
				return ui.Colorize(ui.ColorWarning, noColor, "WRN")
			case "error":
				return ui.Colorize(ui.ColorError, noColor, "ERR")
			case "fatal":
				return ui.Colorize(ui.ColorError, noColor, "FTL")
			case "panic":
				return ui.Colorize(ui.ColorErrorBold, noColor, "PNC")
			default:
				return ui.Colorize(ui.ColorInfo, noColor, "???")
			}
		}
		if i == nil {
			return ui.Colorize(ui.ColorInfo, noColor, "???")
		}
		return strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
	}
}
