package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudquery/cloudquery/cmd/fetch"
	initCmd "github.com/cloudquery/cloudquery/cmd/init"
	"github.com/cloudquery/cloudquery/cmd/policy"
	"github.com/cloudquery/cloudquery/cmd/provider"
	"github.com/cloudquery/cloudquery/cmd/utils"
	"github.com/cloudquery/cloudquery/internal/analytics"
	cqpflag "github.com/cloudquery/cloudquery/internal/pflag"
	"github.com/cloudquery/cloudquery/pkg/core"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cq-provider-sdk/helpers"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	zerolog "github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// fileDescriptorF gets set trough system relevant files like ulimit_unix.go
var fileDescriptorF func()

var (
	// Values for Commit and Date should be injected at build time with -ldflags "-X github.com/cloudquery/cloudquery/cmd.Variable=Value"

	Commit     = "development"
	Date       = "unknown"
	APIKey     = ""
	InstanceId = uuid.New()
	rootShort  = "CloudQuery CLI"
	rootLong   = `CloudQuery CLI

Query your cloud assets & configuration with SQL for monitoring security, compliance & cost purposes.

Find more information at:
	https://docs.cloudquery.io`
)

type rootOptions struct {
	Verbose bool
	Color   *cqpflag.Enum

	// Output
	OutputFormat string

	// Sentry
	SentryDSN   string
	SentryDebug bool

	// Logging
	LogConsole        bool
	NoLogFile         bool
	LogFormat         *cqpflag.Enum
	LogLevel          *cqpflag.Enum
	LogDirectory      string
	LogFilename       string
	LogFileMaxSize    int
	LogFileMaxBackups int
	LogFileMaxAge     int

	// Telemetry
	NoTelemetry     bool
	TelemtryInspect bool
	TelemtryDebug   bool
	TelemetryAPIKey string
}

func newCmdRoot() *cobra.Command {
	o := rootOptions{}
	cmd := &cobra.Command{
		Use:     "cloudquery",
		Short:   rootShort,
		Long:    rootLong,
		Version: core.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Don't print usage on command errors.
			// PersistentPreRunE runs after argument parsing, so errors during parsing will result in printing the help
			cmd.SilenceUsage = true

			initLogging(o)
			initSentry(o)
			err := initAnalytics(o)
			if err != nil {
				return err
			}

			if analytics.Enabled() {
				ui.ColorizedOutput(ui.ColorInfo, "Anonymous telemetry collection and crash reporting enabled. Run with --no-telemetry to disable, or check docs at https://docs.cloudquery.io/docs/cli/telemetry\n")
				if ui.IsTerminal() {
					if err := helpers.Sleep(cmd.Context(), 2*time.Second); err != nil {
						return err
					}
				}
			}
			logInvocationParams(cmd, args)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			analytics.Close()
		},
	}

	pflag.String("api-key", "", "API key for telemetry")

	// General Flags
	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "enable verbose output")
	o.Color = cqpflag.NewEnum([]string{"on", "off", "auto"}, "auto")
	cmd.PersistentFlags().Var(o.Color, "color", "Enable colorized output (on, off, auto)")

	// Logging Flags
	cmd.PersistentFlags().BoolVar(&o.LogConsole, "log-console", false, "enable console logging")
	o.LogLevel = cqpflag.NewEnum([]string{"debug", "info", "warn", "error", "fatal", "panic"}, "info")
	cmd.PersistentFlags().Var(o.LogLevel, "log-level", "log level (debug, info, warn, error, fatal, panic). default: info")
	cmd.PersistentFlags().BoolVar(&o.NoLogFile, "no-log-file", true, "enable file logging")
	cmd.PersistentFlags().StringVar(&o.LogDirectory, "log-directory", ".", "set output directory for logs")
	cmd.PersistentFlags().StringVar(&o.LogFilename, "log-filename", "cloudquery.log", "set output filename for logs")
	o.LogFormat = cqpflag.NewEnum([]string{"json", "keyvalue"}, "keyvalue")
	cmd.PersistentFlags().Var(o.LogFormat, "log-format", "Logging format (json, keyvalue)")
	cmd.PersistentFlags().IntVar(&o.LogFileMaxAge, "log-file-max-size", 30, "set max size in MB of the logfile before it's rolled")
	cmd.PersistentFlags().IntVar(&o.LogFileMaxBackups, "log-file-max-backups", 3, "set max number of rolled files to keep")
	cmd.PersistentFlags().IntVar(&o.LogFileMaxAge, "log-file-max-age", 3, "set max age in days to keep a logfile")

	cmd.PersistentFlags().BoolVar(&o.NoTelemetry, "no-telemetry", false, "disable telemetry collection")
	cmd.PersistentFlags().BoolVar(&o.TelemtryInspect, "telemetry-inspect", false, "enable telemetry inspection")
	cmd.PersistentFlags().BoolVar(&o.TelemtryDebug, "telemetry-debug", false, "enable telemetry debug logging")
	cmd.PersistentFlags().StringVar(&o.TelemetryAPIKey, "telemetry-apikey", APIKey, "set telemetry API key")

	hiddenFlags := []string{
		"telemetry-inspect", "telemtry-debug", "telemtry-apikey",
		"sentry-debug", "sentry-dsn",
		"log-max-age", "log-max-backups", "log-max-size"}
	for _, f := range hiddenFlags {
		err := cmd.PersistentFlags().MarkHidden(f)
		if err != nil {
			panic(err)
		}
	}

	registerSentryFlags(cmd)
	initViper(cmd)

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		initCmd.NewCmdInit(), fetch.NewCmdFetch(), policy.NewCmdPolicy(), provider.NewCmdProvider(),
		newCmdVersion(), newCmdDoc())
	cmd.DisableAutoGenTag = true
	return cmd
}

func Execute() error {
	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().Recover(err)
			panic(err)
		}
	}()
	return newCmdRoot().Execute()
}

func initUlimit() {
	if fileDescriptorF != nil {
		fileDescriptorF()
	}
}

func initViper(cmd *cobra.Command) {
	bindFlags(cmd)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CQ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigFile(".cloudqueryrc")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.cq")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			viper.BindEnv(f.Name, fmt.Sprintf("CQ_%s", envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func checkForUpdate(ctx context.Context) {
	v, err := core.CheckCoreUpdate(ctx, afero.Afero{Fs: afero.NewOsFs()}, time.Now().Unix(), core.UpdateCheckPeriod)
	if err != nil {
		log.Warn().Err(err).Msg("update check failed")
		return
	}
	if v != nil {
		ui.ColorizedOutput(ui.ColorInfo, "An update to CloudQuery core is available: %s!\n\n", v)
		log.Debug().Str("new_version", v.String()).Msg("update check succeeded")
	} else {
		log.Debug().Msg("update check succeeded, no new version")
	}
}

// func initLogging() {
// 	if funk.ContainsString(os.Args, "completion") {
// 		return
// 	}
// 	if !ui.IsTerminal() {
// 		logging.GlobalConfig.ConsoleLoggingEnabled = true // always true when no terminal
// 	}
// 	logging.GlobalConfig.InstanceId = utils.InstanceId.String()

// 	zerolog.Logger = logging.Configure(logging.GlobalConfig).With().Logger()
// }

func initAnalytics(o rootOptions) error {
	opts := []analytics.Option{
		analytics.WithVersionInfo(core.Version, Commit, Date),
		analytics.WithTerminal(ui.IsTerminal()),
		analytics.WithApiKey(o.TelemetryAPIKey),
		analytics.WithInstanceId(utils.InstanceId.String()),
	}
	if o.NoTelemetry {
		opts = append(opts, analytics.WithDisabled())
	}
	if o.TelemtryDebug {
		opts = append(opts, analytics.WithDebug())
	}
	if o.TelemtryInspect {
		opts = append(opts, analytics.WithInspect())
	}

	return analytics.Init(opts...)
}

func logInvocationParams(cmd *cobra.Command, args []string) {
	l := zerolog.Info().Str("core_version", core.Version)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name == "dsn" {
			l = l.Str("pflag:"+f.Name, "(redacted)")
			return
		}

		l = l.Str("pflag:"+f.Name, f.Value.String())
	})
	cmd.Flags().Visit(func(f *pflag.Flag) {
		l = l.Str("flag:"+f.Name, f.Value.String())
	})

	l.Str("command", cmd.CommandPath()).Strs("args", args).Msg("Invocation parameters")
}
