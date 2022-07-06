package policy

import (
	"os"

	"github.com/cloudquery/cloudquery/pkg/policy"
	"github.com/cloudquery/cq-provider-sdk/database"
	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	testShort   = "Tests policy against a precompiled set of database snapshots"
	testExample = `
	# Download & Run the policies defined in your config
	cloudquery policy test path/to/policy.hcl path/to/snapshot/dir selector
		`
)

func newCmdPolicyTest() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "test",
		Short:   testShort,
		Long:    testShort,
		Example: testExample,
		RunE:    runTest,
		Args:    cobra.ExactArgs(2),
	}
	flags := cmd.Flags()
	flags.StringVar(&outputDir, "output-dir", "", "Generates a new file for each policy at the given dir with the output")
	flags.BoolVar(&noResults, "no-results", false, "Do not show policies results")
	return cmd
}

func runTest(cmd *cobra.Command, args []string) error {
	conn, err := database.New(cmd.Context(), hclog.NewNullLogger(), c.Storage.DSN())
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to new database")
		return err
	}
	defer conn.Close()
	uniqueTempDir, err := os.MkdirTemp(os.TempDir(), "*-myOptionalSuffix")
	if err != nil {
		return err
	}

	p, diags := policy.Load(cmd.Context(), c.cfg.CloudQuery.PolicyDirectory, &policy.Policy{Name: "test-policy", Source: policySource})
	if diags.HasErrors() {
		log.Error().Err(err).Msg("failed to load policy")
		return diags
	}

	e := policy.NewExecutor(conn, c.StateManager, nil)
	return p.Test(ctx, e, policySource, snapshotDestination, uniqueTempDir)
}
