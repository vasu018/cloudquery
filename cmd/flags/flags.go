package flags

import "github.com/spf13/cobra"

func AddDBFlags(cmd *cobra.Command) {
	cmd.Flags().String("dsn", "postgres://postgres:pass@locahost:5432/postgres", "DSN of the database to connect to")
}
