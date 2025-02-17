package repository

import (
	"github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/spf13/cobra"
)

func Root(clients *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "repository",
		Aliases:      []string{"repo", "repsitories"},
		Short:        "Pipelines as Code repositories",
		Long:         `Manage Pipelines as Code repositories`,
		SilenceUsage: true,
	}
	cmd.AddCommand(ListCommand(clients, ioStreams))
	cmd.AddCommand(DescribeCommand(clients, ioStreams))
	cmd.AddCommand(CreateCommand(clients, ioStreams))

	return cmd
}
