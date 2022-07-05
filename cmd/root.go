package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudquery/cloudquery/cmd/fetch"
	initCmd "github.com/cloudquery/cloudquery/cmd/init"
	"github.com/cloudquery/cloudquery/cmd/options"
	"github.com/cloudquery/cloudquery/cmd/policy"
	"github.com/cloudquery/cloudquery/cmd/provider"
	"github.com/cloudquery/cloudquery/cmd/utils"
	"github.com/cloudquery/cloudquery/internal/analytics"
	"github.com/cloudquery/cloudquery/internal/logging"
	"github.com/cloudquery/cloudquery/pkg/core"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cq-provider-sdk/helpers"
	"github.com/getsentry/sentry-go"
	zerolog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
)

// fileDescriptorF gets set trough system relevant files like ulimit_unix.go
var fileDescriptorF func()

var (
	// Values for Commit and Date should be injected at build time with -ldflags "-X github.com/cloudquery/cloudquery/cmd.Variable=Value"

	Commit    = "development"
	Date      = "unknown"
	APIKey    = ""
	rootShort = "CloudQuery CLI"
	rootLong  = `CloudQuery CLI

Query your cloud assets & configuration with SQL for monitoring security, compliance & cost purposes.

Find more information at:
	https://docs.cloudquery.io`
)

type rootOptions struct {
	Verbose bool
	NoColor bool

	// Output
	OutputFormat string

	// Sentry
	SentryDSN   string
	SentryDebug bool

	// Logging
	LogConsole        bool
	NoLogFile         bool
	LogFormat         string
	LogLevel          string
	LogDirectory      string
	LogFilename       string
	LogFileMaxSize    int
	LogFileMaxBackups int
	LogFileMaxAge     int
}

func newCmdRoot() *cobra.Command {
	o := rootOptions{}
	rootCmd := &cobra.Command{
		Use:     "cloudquery",
		Short:   rootShort,
		Long:    rootLong,
		Version: core.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Don't print usage on command errors.
			// PersistentPreRunE runs after argument parsing, so errors during parsing will result in printing the help
			cmd.SilenceUsage = true

			initSentry(o)

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

	// Logging Flags
	rootCmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&o.NoColor, "no-color", false, "disable color in output")
	rootCmd.PersistentFlags().BoolVar(&o.LogConsole, "log-console", false, "enable console logging")
	rootCmd.PersistentFlags().BoolVar(&o.NoLogFile, "no-log-file", true, "enable file logging")
	rootCmd.PersistentFlags().StringVar(&o.LogDirectory, "log-directory", ".", "set output directory for logs")
	rootCmd.PersistentFlags().StringVar(&o.LogFilename, "log-filename", "cloudquery.log", "set output filename for logs")
	rootCmd.PersistentFlags().StringVar(&o.LogFormat, "log-format", "default", "enable JSON log format, instead of key/value")
	rootCmd.PersistentFlags().IntVar(&o.LogFileMaxAge, "log-file-max-size", 30, "set max size in MB of the logfile before it's rolled")
	rootCmd.PersistentFlags().IntVar(&o.LogFileMaxBackups, "log-file-max-backups", 3, "set max number of rolled files to keep")
	rootCmd.PersistentFlags().IntVar(&o.LogFileMaxAge, "log-file-max-age", 3, "set max age in days to keep a logfile")

	rootCmd.PersistentFlags().Bool("no-telemetry", false, "disable telemetry collection")
	rootCmd.PersistentFlags().Bool("telemetry-inspect", false, "enable telemetry inspection")
	rootCmd.PersistentFlags().Bool("telemetry-debug", false, "enable telemetry debug logging")

	_ = rootCmd.PersistentFlags().MarkHidden("telemetry-inspect")
	_ = rootCmd.PersistentFlags().MarkHidden("telemetry-debug")

	registerSentryFlags(rootCmd)
	initViper()

	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cobra.OnInitialize(initLogging, initUlimit, initSentry, initAnalytics)
	rootCmd.AddCommand(
		initCmd.NewCmdInit(), fetch.NewCmdFetch(), policy.NewCmdPolicy(), provider.NewCmdProvider(),
		options.NewCmdOptions(), newCmdVersion(), newCmdDoc())
	rootCmd.DisableAutoGenTag = true
	return rootCmd
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

func initViper() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CQ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigFile("cloudquery")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.cq")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func initLogging() {
	if funk.ContainsString(os.Args, "completion") {
		return
	}
	if !ui.IsTerminal() {
		logging.GlobalConfig.ConsoleLoggingEnabled = true // always true when no terminal
	}
	logging.GlobalConfig.InstanceId = utils.InstanceId.String()

	zerolog.Logger = logging.Configure(logging.GlobalConfig).With().Logger()
}

func initAnalytics() {
	opts := []analytics.Option{
		analytics.WithVersionInfo(core.Version, Commit, Date),
		analytics.WithTerminal(ui.IsTerminal()),
		analytics.WithApiKey(viper.GetString("telemetry-apikey")),
		analytics.WithInstanceId(utils.InstanceId.String()),
	}
	userId := analytics.GetCookieId()
	if viper.GetBool("no-telemetry") || analytics.CQTeamID == userId.String() {
		opts = append(opts, analytics.WithDisabled())
	}
	if viper.GetBool("debug-telemetry") {
		opts = append(opts, analytics.WithDebug())
	}
	if viper.GetBool("inspect-telemetry") {
		opts = append(opts, analytics.WithInspect())
	}

	_ = analytics.Init(opts...)
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
