package cluster

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewCmdClusterHealth implements the base cluster health command
func NewCmdCluster(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:               "cluster",
		Short:             "cluster related utilities",
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	clusterCmd.AddCommand(newCmdHealth(streams, flags))
	clusterCmd.AddCommand(newCmdMustGather(streams, flags)) // capture must-gather
	return clusterCmd
}
