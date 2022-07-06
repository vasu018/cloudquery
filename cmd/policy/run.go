package policy

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/cloudquery/cloudquery/cmd/diags"
	"github.com/cloudquery/cloudquery/internal/analytics"
	"github.com/cloudquery/cloudquery/pkg/policy"
	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cloudquery/pkg/ui/console"
	"github.com/cloudquery/cq-provider-sdk/provider/diag"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v6/decor"
)

var (
	outputDir    string
	noResults    bool
	storeResults bool
)

const (
	runShort     = "Executes a policy on CloudQuery database"
	exampleShort = `
# Run an official policy
# Official policies are available on our hub: https://hub.cloudquery.io/policies
cloudquery policy run aws

# Run a sub-policy of an official policy
cloudquery policy run aws//cis_v1.2.0

# Run a policy from a GitHub repository
cloudquery policy run github.com/<repo-owner>/<repo-name>

# Run a policy from a local directory
cloudquery policy run ./<path-to-local-directory>

# See https://hub.cloudquery.io for additional policies
# See https://docs.cloudquery.io/docs/tutorials/policies/policies-overview for instructions on writing policies`
)

func newCmdPolicyRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <policy>",
		Short:   runShort,
		Long:    runShort,
		Example: exampleShort,
		RunE:    runPolicyRun,
		Args:    cobra.ExactArgs(1),
	}
	flags := cmd.Flags()
	flags.StringVar(&outputDir, "output-dir", "", "Generates a new file for each policy at the given dir with the output")
	flags.BoolVar(&noResults, "no-results", false, "Do not show policies results")
	flags.BoolVar(&storeResults, "enable-db-persistence", false, "Enable storage of policy output in database")
	flags.Bool("disable-fetch-check", false, "Disable checking if a respective fetch happened before running policies")
	_ = viper.BindPFlag("disable-fetch-check", flags.Lookup("disable-fetch-check"))
	return cmd
}

func runPolicyRun(cmd *cobra.Command, args []string) error {
	defer diags.PrintDiagnostics("", &diags, viper.GetBool("redact-diags"), viper.GetBool("verbose"))

	// use config value for dbPersistence if not already enabled through the cli
	if !dbPersistence && c.cfg.CloudQuery.Policy != nil {
		dbPersistence = c.cfg.CloudQuery.Policy.DBPersistence
	}

	policiesToRun, err := ParseAndDetect(policySource)
	if err != nil {
		ui.ColorizedOutput(ui.ColorError, err.Error())
		return diag.FromError(err, diag.RESOLVING)
	}
	log.Debug().Interface("policies", policiesToRun).Msg("policies to run")
	ui.ColorizedOutput(ui.ColorProgress, "Starting policies run...\n\n")
	var (
		policyRunProgress ui.Progress
		policyRunCallback policy.UpdateCallback
	)
	// if we are running in a terminal, build the progress bar
	if ui.DoProgress() {
		policyRunProgress, policyRunCallback = buildPolicyRunProgress(ctx, policiesToRun)
	}
	// Policies run request
	resp, diags := policy.Run(cmd.Context(), c.StateManager, c.Storage, &policy.RunRequest{
		Policies:      policiesToRun,
		Directory:     c.cfg.CloudQuery.PolicyDirectory,
		OutputDir:     outputDir,
		RunCallback:   policyRunCallback,
		DBPersistence: dbPersistence,
	})
	if resp != nil {
		policiesToRun = resp.Policies
	}
	for _, p := range policiesToRun {
		analytics.Capture("policy run", c.Providers, p.Analytic(dbPersistence), diags)
	}

	if policyRunProgress != nil {
		policyRunProgress.MarkAllDone()
		policyRunProgress.Wait()
	}
	if !noResults && resp != nil {
		printPolicyResponse(resp.Executions)
	}

	if diags.HasErrors() {
		ui.SleepBeforeError(cmd.Context())
		ui.ColorizedOutput(ui.ColorError, "❌ Failed to run policies\n\n")
		return diags
	}

	ui.ColorizedOutput(ui.ColorProgress, "Finished policies run...\n\n")
	return nil
	return nil
}

func printPolicyResponse(results []*policy.ExecutionResult) {
	if len(results) == 0 {
		return
	}
	for _, execResult := range results {
		ui.ColorizedOutput(ui.ColorUnderline, "%s %s Results:\n\n", console.EmojiStatus[ui.StatusInfo], execResult.PolicyName)

		if !execResult.Passed {
			if execResult.Error != "" {
				ui.ColorizedOutput(ui.ColorHeader, ui.ColorErrorBold.Sprintf("%s Policy failed to run\nError: %s\n\n", console.EmojiStatus[ui.StatusError], execResult.Error))
			} else {
				ui.ColorizedOutput(ui.ColorHeader, ui.ColorErrorBold.Sprintf("%s Policy finished with violations\n\n", console.EmojiStatus[ui.StatusWarn]))
			}
		}
		for _, res := range execResult.Results {
			switch res.Type {
			case policy.ManualQuery:
				ui.ColorizedOutput(ui.ColorInfo, "%s: Policy %s - %s\n\n", color.YellowString("Manual"), res.Name, res.Description)
				ui.ColorizedOutput(ui.ColorInfo, "\n")
			case policy.AutomaticQuery:
				if res.Passed {
					ui.ColorizedOutput(ui.ColorInfo, "%s: Policy %s - %s\n\n", color.GreenString("Passed"), res.Name, res.Description)
				} else {
					ui.ColorizedOutput(ui.ColorInfo, "%s: Policy %s - %s\n\n", color.RedString("Failed"), res.Name, res.Description)
				}
			}
			if len(res.Rows) > 0 {
				createOutputTable(res)
				ui.ColorizedOutput(ui.ColorInfo, "\n\n")
			}
		}
	}
}

func createOutputTable(res *policy.QueryResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(res.Columns)

	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	sort.Sort(res.Rows)
	max := len(res.Columns)
	for _, row := range res.Rows {
		data := make([]string, 0, len(res.Columns))
		data = append(data, color.HiRedString(row.Status))
		for _, key := range res.Columns {
			if val, ok := row.Identifiers[key]; ok {
				data = append(data, cast.ToString(val))
			}
		}
		data = append(data, row.Reason)
		for _, key := range res.Columns {
			if val, ok := row.AdditionalData[key]; ok {
				data = append(data, cast.ToString(val))
			}
		}
		table.Append(data)
		if len(data) > max {
			max = len(data)
		}
	}
	table.SetFooter(append(make([]string, max-2, max), "Total:", strconv.Itoa(len(res.Rows))))
	table.Render()
}

func buildPolicyRunProgress(ctx context.Context, policies policy.Policies) (*Progress, policy.UpdateCallback) {
	policyRunProgress := NewProgress(ctx, func(o *ProgressOptions) {
		o.AppendDecorators = []decor.Decorator{decor.CountersNoUnit(" Finished Checks: %d/%d")}
	})

	for _, p := range policies {
		policyRunProgress.Add(p.Name, fmt.Sprintf("policy \"%s\" - ", p.Name), "evaluating - ", 1)
	}

	policyRunCallback := func(update policy.Update) {
		bar := policyRunProgress.GetBar(update.PolicyName)
		// try to get with policy source
		if bar == nil {
			bar = policyRunProgress.GetBar(update.Source)
		}
		if bar == nil {
			policyRunProgress.AbortAll()
			ui.ColorizedOutput(ui.ColorError, "❌ console UI failure, policy run will complete shortly\n")
			return
		}
		if update.Error != "" {
			policyRunProgress.Update(update.PolicyName, ui.StatusError, fmt.Sprintf("error: %s", update.Error), 0)
			return
		}

		// set the total queries to track
		if update.QueriesCount > 0 {
			bar.SetTotal(int64(update.QueriesCount), false)
		}

		bar.b.IncrBy(update.DoneCount() - int(bar.b.Current()))

		if update.AllDone() && bar.Status != ui.StatusWarn {
			policyRunProgress.Update(update.PolicyName, ui.StatusOK, "policy run complete - ", 0)
			bar.Done()
			return
		}
	}

	return policyRunProgress, policyRunCallback
}
