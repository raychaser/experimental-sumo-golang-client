package cmd

import (
	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	result := &cobra.Command{
		Use: "query",
	}
	result.AddCommand(newQueryLogsCmd().cmd)
	result.AddCommand(newQueryMetricsCmd().cmd)
	return result
}
