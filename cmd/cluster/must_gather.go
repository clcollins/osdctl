package cluster

import (
	// "context"
	"context"
	"fmt"

	"github.com/openshift/osdctl/pkg/k8s"
	"github.com/spf13/cobra"

	// "k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mustGatherOptions struct {
	flags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	kubeCli client.Client
}

// newCmdMustGather implements the mustGather command
// to retrieve a must-gather from the must-gather-operator on a cluster
func newCmdMustGather(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *cobra.Command {
	ops := newMustGatherOptions(streams, flags)
	mustGatherCmd := &cobra.Command{
		Use:               "must-gather",
		Short:             "Retrieves a must-gather from the currently logged-in cluster",
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(ops.complete(cmd, args))
			cmdutil.CheckErr(ops.run())
		},
	}

	// mustGatherCmd.Flags().StringVarP(&ops.output, "out", "o", "default", "Output format [default | json | env]")
	// mustGatherCmd.Flags().BoolVarP(&ops.verbose, "verbose", "", false, "Verbose output")

	return mustGatherCmd
}

func newMustGatherOptions(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *mustGatherOptions {
	return &mustGatherOptions{
		flags:     flags,
		IOStreams: streams,
	}
}

func (o *mustGatherOptions) run() error {
	podList := &corev1.PodList{}

	err := o.kubeCli.List(context.TODO(), podList, &client.ListOptions{Namespace: "openshift-monitoring"})
	if err != nil {
		return err
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found")
	}

	fmt.Printf("POD COUNT: %d\n", len(podList.Items))

	fmt.Printf("%+v\n", &podList.Items[0])

	return nil
}

type MustGather struct {
}

type MustGatherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MustGather `json:"items"`
}

func (o *mustGatherOptions) complete(cmd *cobra.Command, _ []string) error {
	var err error

	o.kubeCli, err = k8s.NewClient(o.flags)
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}

	return nil
}
