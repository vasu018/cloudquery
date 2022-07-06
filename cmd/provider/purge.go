package provider

import (
	"fmt"
	"time"

	"github.com/cloudquery/cloudquery/cmd/diags"
	"github.com/cloudquery/cloudquery/cmd/utils"
	"github.com/cloudquery/cloudquery/pkg/core"
	"github.com/cloudquery/cloudquery/pkg/errors"
	"github.com/cloudquery/cloudquery/pkg/plugin/registry"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cloudquery/pkg/ui/console"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	lastUpdate time.Duration
	dryRun     bool
	purgeShort = "Remove stale resources from one or more providers in database"
)

func newCmdProviderPurge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge [provider]",
		Short: purgeShort,
		Long:  purgeShort,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := console.CreateClient(cmd.Context(), viper.ConfigFileUsed(), false, nil, utils.InstanceId)
			if err != nil {
				return err
			}
			diags := c.RemoveStaleData(cmd.Context(), lastUpdate, dryRun, args)
			errors.CaptureDiagnostics(diags, map[string]string{"command": "provider_purge"})
			if diags.HasErrors() {
				return fmt.Errorf("failed to remove stale data: %w", diags)
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&lastUpdate, "last-update", time.Hour*1,
		"last-update is the duration from current time we want to remove resources from the database. "+
			"For example 24h will remove all resources that were not update in last 24 hours. Duration is a string with optional unit suffix such as \"2h45m\" or \"7d\"")
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "")
	return cmd
}

func runPurge(cmd *cobra.Command, args []string) error {
	defer printDiagnostics("", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
	if dd := c.DownloadProviders(ctx); dd.HasErrors() {
		return dd
	}
	pp := make([]registry.Provider, len(providers))
	for i, p := range providers {
		rp, ok := c.Providers.Get(p)
		if !ok {
			ui.ColorizedOutput(ui.ColorHeader, "unknown provider %s requested..\n\n", p)
		}
		pp[i] = rp
	}
	ui.ColorizedOutput(ui.ColorHeader, "Purging providers %s resources..\n\n", providers)
	defer diags.PrintDiagnostics("", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
	result, diags := core.PurgeProviderData(cmd.Context(), c.Storage, c.PluginManager, &core.PurgeProviderDataOptions{
		Providers:  pp,
		LastUpdate: lastUpdate,
		DryRun:     dryRun,
	})

	if dryRun && !diags.HasErrors() {
		ui.ColorizedOutput(ui.ColorWarning, "Expected resources to be purged: %d. Use --dry-run=false to purge these resources.\n", result.TotalAffected)
		for _, r := range result.Resources() {
			ui.ColorizedOutput(ui.ColorWarning, "\t%s: %d resources\n\n", r, result.AffectedResources[r])
		}
	}
	if diags.HasErrors() {
		ui.ColorizedOutput(ui.ColorProgress, "Purge for providers %s failed\n\n", providers)
		return diags
	}
	ui.ColorizedOutput(ui.ColorProgress, "Purge for providers %s was successful\n\n", providers)
	return diags
	return nil
}
