package policy

import (
	"fmt"

	"github.com/cloudquery/cloudquery/internal/getter"
	"github.com/cloudquery/cloudquery/pkg/policy"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/spf13/cobra"
)

const (
	snapshotShort = `Take database snapshot of all tables included in a CloudQuery policy`
)

func newCmdPolicySnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: snapshotShort,
		Long:  snapshotShort,
		Args:  cobra.ExactArgs(2),
		RunE:  runSnapshot,
	}
	return cmd
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	p, err := policy.Load(cmd.Context(), c.cfg.CloudQuery.PolicyDirectory, &policy.Policy{Name: p.Name, Source: p.Source})
	if err != nil {
		ui.ColorizedOutput(ui.ColorError, err.Error())
		return fmt.Errorf("failed to load policies: %w", err)
	}
	if !p.HasChecks() {
		return fmt.Errorf("no checks loaded")
	}

	_, subPath := getter.ParseSourceSubPolicy(args[0])
	pol := p.Filter(subPath)
	if pol.TotalQueries() != 1 {
		return fmt.Errorf("selector must specify only a single control")
	}
	return policy.Snapshot(cmd.Context(), c.StateManager, c.Storage, &pol, args[1], subPath)
}
