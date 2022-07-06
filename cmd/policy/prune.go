package policy

import (
	"fmt"
	"time"

	"github.com/cloudquery/cloudquery/cmd/diags"
	"github.com/cloudquery/cloudquery/pkg/policy"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cq-provider-sdk/provider/diag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	pruneShort   = "Prune policy executions from the database which are older than the relative time specified"
	pruneExample = `
# Prune the policy executions which are older than the relative time specified
cloudquery policy prune 24h`
)

func newCmdPolicyPrune() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prune",
		Short:   pruneShort,
		Long:    pruneShort,
		Example: pruneExample,
		Args:    cobra.ExactArgs(1),
		RunE:    runPolicyRun,
	}
	return cmd
}

func runPrune(cmd *cobra.Command, args []string) error {
	defer diags.PrintDiagnostics("", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))
	duration, err := time.ParseDuration(args[1])
	if err != nil {
		ui.ColorizedOutput(ui.ColorError, err.Error())
		return diag.FromError(err, diag.USER)
	}
	pruneBefore := time.Now().Add(-duration)
	if !pruneBefore.Before(time.Now()) {
		return diag.FromError(fmt.Errorf("prune retention period can't be in the future"), diag.USER)
	}
	return policy.Prune(cmd.Context(), c.StateManager, pruneBefore)
}
