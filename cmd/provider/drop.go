package provider

import (
	"fmt"

	"github.com/cloudquery/cloudquery/internal/firebase"
	"github.com/cloudquery/cloudquery/pkg/core"
	"github.com/cloudquery/cloudquery/pkg/core/database"
	"github.com/cloudquery/cloudquery/pkg/core/state"
	"github.com/cloudquery/cloudquery/pkg/errors"
	"github.com/cloudquery/cloudquery/pkg/plugin"
	"github.com/cloudquery/cloudquery/pkg/plugin/registry"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cloudquery/pkg/ui/console"
	"github.com/cloudquery/cq-provider-sdk/provider/diag"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v6/decor"
)

var (
	providerForce bool
	dropShort     = "Drops provider schema from database"
)

func newCmdProviderDrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop [provider]",
		Short: dropShort,
		Long:  dropShort,
		Args:  cobra.ExactArgs(1),
		RunE:  runDrop,
	}
	cmd.Flags().BoolVar(&providerForce, "force", false, "Really drop tables for the provider")
	return cmd
}

func runDrop(cmd *cobra.Command, args []string) error {
	var progressUpdater ui.Progress

	if ui.DoProgress() {
		progressUpdater = console.NewProgress(cmd.Context(), func(o *console.ProgressOptions) {
			o.AppendDecorators = []decor.Decorator{decor.Percentage()}
		})
	}

	if !providerForce {
		ui.ColorizedOutput(ui.ColorWarning, "WARNING! This will drop all tables for the given provider. If you wish to continue, use the --force flag.\n")
		return diag.FromError(fmt.Errorf("if you wish to continue, use the --force flag"), diag.USER)
	}

	hub := registry.NewRegistryHub(firebase.CloudQueryRegistryURL, registry.WithPluginDirectory(cfg.CloudQuery.PluginDirectory), registry.WithProgress(progressUpdater))
	pm, err := plugin.NewManager(hub, plugin.WithAllowReattach())
	if err != nil {
		return err
	}
	providers := make([]registry.Provider, len(args))
	for i := range providers {
		providers[i] = registry.Provider{Name: args[i]}
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

	core.Drop(cmd.Context(), stateManager, pm, providers[0])

	errors.CaptureDiagnostics(diags, map[string]string{"command": "provider_drop"})
	if diags.HasErrors() {
		return fmt.Errorf("failed to drop provider %s %w", args[0], diags)
	}
	// pr[0].

	return nil
}
