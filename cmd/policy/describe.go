package policy

import (
	"github.com/spf13/cobra"
)

const (
	describeShort   = "Describe CloudQuery policy"
	describeExample = `
# Describe official policy
cloudquery policy describe aws

# The following will be the same as above
# Official policies are hosted here: https://github.com/cloudquery-policies
cloudquery policy describe aws//cis-1.2.0

# Describe community policy
cloudquery policy describe github.com/COMMUNITY_GITHUB_ORG/aws

# See https://hub.cloudquery.io for additional policies.
`
	describeDeprecated = "policy describe is deprecated"
)

func newCmdPolicyDescribe() *cobra.Command {
	describePolicyCmd := &cobra.Command{
		Use:        "describe",
		Short:      describeShort,
		Long:       describeShort,
		Example:    describeExample,
		Args:       cobra.ExactArgs(1),
		Deprecated: describeDeprecated,
	}
	return describePolicyCmd
}
