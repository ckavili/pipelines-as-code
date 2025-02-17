package version

import (
	"fmt"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/version"
	"github.com/spf13/cobra"
)

func Command(ioStreams *cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print tkn pac version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(ioStreams.ErrOut, version.Version)
		},
	}
	return cmd
}
