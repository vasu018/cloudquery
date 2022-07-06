package fetch

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudquery/cloudquery/cmd/flags"
	"github.com/cloudquery/cloudquery/internal/firebase"
	"github.com/cloudquery/cloudquery/pkg/config"
	"github.com/cloudquery/cloudquery/pkg/core"
	"github.com/cloudquery/cloudquery/pkg/core/database"
	"github.com/cloudquery/cloudquery/pkg/core/state"
	"github.com/cloudquery/cloudquery/pkg/plugin"
	"github.com/cloudquery/cloudquery/pkg/plugin/registry"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cq-provider-sdk/cqproto"
	"github.com/cloudquery/cq-provider-sdk/provider/diag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v6/decor"
	"google.golang.org/grpc/status"
)

const (
	fetchShort = "Fetch resources from configured providers"
	fetchLong  = `Fetch resources from configured providers
	
	This requires a cloudquery.yml file which can be generated by "cloudquery init"
	`
	fetchExample = `  # Fetch configured providers to PostgreSQL as configured in cloudquery.yml
	cloudquery fetch`
)

func NewCmdFetch() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fetch",
		Short:   fetchShort,
		Long:    fetchLong,
		Example: fetchExample,
		RunE:    runFetch,
	}
	flags.AddDBFlags(cmd)
	cmd.Flags().String("config", "cloudquery.yml", "Path to the cloudquery fetch configuration file")
	cmd.Flags().Bool("redact-diags", false, "show redacted diagnostics only")
	_ = viper.BindPFlag("redact-diags", cmd.Flags().Lookup("redact-diags"))
	_ = cmd.Flags().MarkHidden("redact-diags")
	return cmd
}

func runFetch(cmd *cobra.Command, args []string) error {
	cfg, ok := loadFetchConfig(viper.ConfigFileUsed())
	if _, dd := c.SyncProviders(cmd.Context(), cfg.cfg.Providers.Names()...); dd.HasErrors() {
		return nil, dd
	}

	hub := registry.NewRegistryHub(firebase.CloudQueryRegistryURL, registry.WithPluginDirectory(cfg.CloudQuery.PluginDirectory), registry.WithProgress(progressUpdater))
	pm, err := plugin.NewManager(hub, plugin.WithAllowReattach())
	if err != nil {
		return err
	}
	pr, err := pm.DownloadProviders(cmd.Context(), providers, true)
	if err != nil {
		return err
	}

	storage := database.NewStorage(cfg.CloudQuery.Connection.DSN, dialect)
	stateManager, err := state.NewClient(cmd.Context(), storage.DSN())
	if err != nil {
		return err
	}

	ui.ColorizedOutput(ui.ColorProgress, "Starting provider fetch...\n\n")
	var (
		fetchProgress ui.Progress
		fetchCallback core.FetchUpdateCallback
	)
	if ui.DoProgress() {
		fetchProgress, fetchCallback = buildFetchProgress(cmd.Context(), c.cfg.Providers)
	}

	providers := make([]core.ProviderInfo, len(c.cfg.Providers))
	for i, p := range c.cfg.Providers {
		rp, ok := c.Providers.Get(p.Name)
		if !ok {
			diags := diag.FromError(fmt.Errorf("failed to find provider %s in configuration", p.Name), diag.USER)
			diags.PrintDiagnostics("Fetch", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
			return nil, diags
		}
		providers[i] = core.ProviderInfo{Provider: rp, Config: p, ConfigFormat: cqproto.ConfigYAML}
	}
	result, diags := core.Fetch(cmd.Context(), c.StateManager, c.Storage, c.PluginManager, &core.FetchOptions{
		UpdateCallback: fetchCallback,
		ProvidersInfo:  providers,
		FetchId:        c.instanceId,
	})
	// first wait for progress to complete correctly
	if fetchProgress != nil {
		fetchProgress.MarkAllDone()
		fetchProgress.Wait()
	}
	// Check if any errors are found
	if diags.HasErrors() {
		// Ignore context cancelled error
		if st, ok := status.FromError(diags); ok && st.Code() == gcodes.Canceled {
			diags.PrintDiagnostics("", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
			ui.ColorizedOutput(ui.ColorProgress, "Provider fetch canceled.\n\n")
			return result, diags
		}
	}
	ui.ColorizedOutput(ui.ColorProgress, "Provider fetch complete.\n\n")
	diags.PrintDiagnostics("Fetch", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
	if result == nil {
		return nil, diags
	}
	for _, summary := range result.ProviderFetchSummary {
		printProviderSummary(summary)
	}
	return result, diags
	return nil
}

func loadFetchConfig(file string) (*config.Config, bool) {
	cfg, diags := config.NewParser().LoadConfigFile(file)
	if diags.HasDiags() {
		ui.ColorizedOutput(ui.ColorHeader, "Configuration Error Diagnostics:\n")
		for _, d := range diags {
			c := ui.ColorInfo
			switch d.Severity() {
			case diag.ERROR:
				c = ui.ColorError
			case diag.WARNING:
				c = ui.ColorWarning
			}
			ui.ColorizedOutput(c, "❌ %s; %s\n", d.Description().Summary, d.Description().Detail)
		}
		if diags.HasErrors() {
			return nil, false
		}
	}
	return cfg, true
}

// printProviderSummary is a helper to print the fetch summary in an easily readable format.
func printProviderSummary(summary *core.ProviderFetchSummary) {
	s := emojiStatus[ui.StatusOK]
	if summary.Status == core.FetchCanceled {
		s = emojiStatus[ui.StatusError] + " (canceled)"
	}
	key := summary.Name
	if summary.Name != summary.Alias {
		key = summary.Name + `(` + summary.Alias + `)`
	}
	diags := summary.Diagnostics().Squash()
	ui.ColorizedOutput(
		ui.ColorHeader,
		"Provider %s fetch summary: %s Total Resources fetched: %d",
		key,
		s,
		summary.TotalResourcesFetched,
	)

	// errors
	errors := formatIssues(diags, diag.ERROR, diag.PANIC)
	if len(errors) > 0 {
		ui.ColorizedOutput(ui.ColorHeader, "\t ❌ Errors: %s", errors)
	}

	// warnings
	warnings := formatIssues(diags, diag.WARNING)
	if len(warnings) > 0 {
		ui.ColorizedOutput(ui.ColorHeader, "\t ⚠️ Warnings: %s", warnings)
	}

	// ignored issues
	ignored := formatIssues(diags, diag.IGNORE)
	if len(ignored) > 0 {
		ui.ColorizedOutput(ui.ColorHeader, "\t ❓ Ignored issues: %s", ignored)
		ui.ColorizedOutput(ui.ColorHeader,
			"\nProvider %s finished with %s ignored issues."+
				"\nThis may be normal, however, you can use `--verbose` flag to see more details.",
			key, ignored)
	}

	ui.ColorizedOutput(ui.ColorHeader, "\n\n")
}

// formatIssues will pretty-print the diagnostics by the requested severities:
// - for no issues "" is returned
// - for any deep issues the "base (deep)" amounts are printed
// - for basic case with no deep issues but rather the base ones, the "base" amount is printed
func formatIssues(diags diag.Diagnostics, severities ...diag.Severity) string {
	basic, deep := countSeverity(diags, severities...)
	switch {
	case deep > 0:
		return strconv.FormatUint(basic, 10) + `(` + strconv.FormatUint(deep, 10) + `)`
	case basic > 0:
		return strconv.FormatUint(basic, 10)
	default:
		return ``
	}
}

func countSeverity(d diag.Diagnostics, sevs ...diag.Severity) (basic, deep uint64) {
	for _, sev := range sevs {
		basic += d.CountBySeverity(sev, false)
	}

	if !viper.GetBool("verbose") {
		return basic, 0
	}

	for _, sev := range sevs {
		deep += d.CountBySeverity(sev, true)
	}
	return basic, deep
}

func buildFetchProgress(ctx context.Context, providers []*config.Provider) (*Progress, core.FetchUpdateCallback) {
	fetchProgress := NewProgress(ctx, func(o *ProgressOptions) {
		o.AppendDecorators = []decor.Decorator{decor.CountersNoUnit(" Finished Resources: %d/%d")}
	})

	for _, p := range providers {
		if len(p.Resources) == 0 {
			ui.ColorizedOutput(ui.ColorWarning, "%s Skipping provider %s[%s] configured with no resource to fetch\n", emojiStatus[ui.StatusWarn], p.Name, p.Alias)
			continue
		}

		if p.Alias != p.Name {
			fetchProgress.Add(fmt.Sprintf("%s_%s", p.Name, p.Alias), fmt.Sprintf("cq-provider-%s (%s)", p.Name, p.Alias), "fetching", int64(len(p.Resources)))
		} else {
			fetchProgress.Add(fmt.Sprintf("%s_%s", p.Name, p.Alias), fmt.Sprintf("cq-provider-%s", p.Name), "fetching", int64(len(p.Resources)))
		}
	}
	fetchCallback := func(update core.FetchUpdate) {
		name := fmt.Sprintf("%s_%s", update.Name, update.Name)
		if update.Alias != "" {
			name = fmt.Sprintf("%s_%s", update.Name, update.Alias)
		}
		if update.DiagnosticCount > 0 {
			fetchProgress.Update(name, ui.StatusWarn, fmt.Sprintf("diagnostics: %d", update.DiagnosticCount), 0)
		}
		bar := fetchProgress.GetBar(name)
		if bar == nil {
			fetchProgress.AbortAll()
			ui.ColorizedOutput(ui.ColorError, "❌ console UI failure, fetch will complete shortly\n")
			return
		}
		if bar.Total < int64(len(update.FinishedResources)) {
			bar.SetTotal(int64(len(update.FinishedResources)), false)
		}

		bar.b.IncrBy(update.DoneCount() - int(bar.b.Current()))

		if bar.Status == ui.StatusError {
			if update.AllDone() {
				bar.SetTotal(0, true)
			}
			return
		}
		if update.AllDone() && bar.Status != ui.StatusWarn {
			fetchProgress.Update(name, ui.StatusOK, "fetch complete", 0)
			return
		}
	}
	return fetchProgress, fetchCallback
}
