package provider

import (
	"fmt"

	"github.com/cloudquery/cloudquery/cmd/utils"
	"github.com/cloudquery/cloudquery/pkg/core"
	"github.com/cloudquery/cloudquery/pkg/errors"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cloudquery/pkg/ui/console"
	"github.com/cloudquery/cq-provider-sdk/provider/diag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const syncShort = "Download the providers specified in config and re-create their database schema"

func newCmdProviderSync() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [providers,...]",
		Short: syncShort,
		Long:  syncShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := console.CreateClient(cmd.Context(), viper.ConfigFileUsed(), false, nil, utils.InstanceId)
			if err != nil {
				return err
			}
			_, diags := c.SyncProviders(cmd.Context(), args...)
			errors.CaptureDiagnostics(diags, map[string]string{"command": "provider_sync"})
			if diags.HasErrors() {
				return fmt.Errorf("failed to sync providers %w", diags)
			}
			return nil
		},
	}
	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
	defer diags.PrintDiagnostics("Sync", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
	providers := c.Providers
	if len(pp) > 0 {
		providers = c.Providers.GetMany(pp...)
	}
	ui.ColorizedOutput(ui.ColorProgress, "Syncing CloudQuery providers %s\n\n", providers)
	if len(providers) == 0 {
		return nil, diag.FromError(fmt.Errorf("one or more providers not found: %s", pp), diag.USER,
			diag.WithDetails("providers not found, are they defined in configuration?. Defined: %s", c.Providers))
	}
	diags = diags.Add(c.DownloadProviders(ctx))
	if diags.HasErrors() {
		return nil, diags
	}

	for _, p := range providers {
		sync, dd := core.Sync(ctx, c.StateManager, c.PluginManager, p)
		if dd.HasErrors() {
			ui.ColorizedOutput(ui.ColorError, "%s failed to sync provider %s.\n", emojiStatus[ui.StatusError], p.String())
			// TODO: should we just append diags and continue to sync others or stop syncing?
			return nil, dd
		}
		if sync.State != core.NoChange {
			ui.ColorizedOutput(ui.ColorSuccess, "%s sync provider %s to %s successfully. [%s]\n", emojiStatus[ui.StatusOK], p.Name, p.Version, sync.State)
		}
		diags = diags.Add(dd)
		if sync != nil {
			results = append(results, sync)
		}
	}
	ui.ColorizedOutput(ui.ColorProgress, "\nFinished syncing providers...\n\n")
	return results, diags
}
