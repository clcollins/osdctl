package cluster

import (
	"fmt"

	// "github.com/openshift/osdctl/pkg/k8s"
	k8spkg "github.com/openshift/osdctl/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mustGatherOptions struct {
	cluster string

	flags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	kubeCli                   client.Client
	k8sclusterresourcefactory k8spkg.ClusterResourceFactoryOptions
}

func newCmdMustGather(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *cobra.Command {
	ops := newMustGatherOptions(streams, flags)
	mustGatherCmd := &cobra.Command{
		Use:   "must-gather",
		Short: "retrieves a must-gather from the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(ops.complete(cmd, args))
			cmdutil.CheckErr(ops.run())
		},
	}

	mustGatherCmd.Flags().StringVarP(&ops.cluster, "cluster", "c", "", "ID of the cluster")
	mustGatherCmd.MarkFlagRequired("cluster")

	return mustGatherCmd
}

func newMustGatherOptions(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *mustGatherOptions {
	return &mustGatherOptions{
		flags:     flags,
		IOStreams: streams,
	}
}

func (o *mustGatherOptions) complete(cmd *cobra.Command, _ []string) error {
	fmt.Printf("%+v", o.flags.Context)
	k8svalid, err := o.k8sclusterresourcefactory.ValidateIdentifiers()
	if !k8svalid {
		if err != nil {
			return err
		}
	}

	// o.kubeCli, err = k8s.NewClient(o.flags)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (o *mustGatherOptions) run() error {
	fmt.Println(o.cluster)
	return nil
}
