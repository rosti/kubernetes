/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alpha

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/componentconfigs"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
)

// newCmdAlphaConfig is a placeholder for alpha commands that would eventually graduate
// into the main `kubeadm config` command
func newCmdAlphaConfig(out io.Writer) *cobra.Command {
	var kubeConfigFile string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration for a kubeadm cluster",
	}

	// Add the --kubeconfig flag here to mimic what the `kubeadm config` command does
	// That way the changes will be smaller when stuff is graduated from alpha and is moved out of here
	options.AddKubeConfigFlag(cmd.PersistentFlags(), &kubeConfigFile)
	kubeConfigFile = cmdutil.GetKubeConfigPath(kubeConfigFile)

	cmd.AddCommand(newCmdAlphaConfigPrint(out, &kubeConfigFile))

	return cmd
}

// newCmdAlphaConfig is a placeholder for alpha commands that would eventually graduate
// into the main `kubeadm config print` command
func newCmdAlphaConfigPrint(out io.Writer, kubeConfigFile *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print configuration",
	}

	cmd.AddCommand(newCmdAlphaConfigPrintUpgradeable(out, kubeConfigFile))

	return cmd
}

// newCmdAlphaConfigPrintUpgradeable handles the `kubeadm alpha config print upgradeable` command
func newCmdAlphaConfigPrintUpgradeable(out io.Writer, kubeConfigPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgradeable",
		Short: "Print component configs that need manual upgrading",
		RunE: func(_ *cobra.Command, _ []string) error {
			// First, obtain a clientset from the kubeconfig file
			client, err := kubeconfig.ClientSetFromFile(*kubeConfigPath)
			if err != nil {
				return err
			}

			// Fetch only the kubeadm configuration from the cluster. Don't fetch the component configs
			// Also, mute this func as it can output a few messages, but we want to keep the output clean for the YAML
			cfg, err := config.FetchInitConfigurationFromCluster(client, ioutil.Discard, "", false, true)
			if err != nil {
				return err
			}

			// Get a DocumentMap with the unsupported component configs
			docmap, err := componentconfigs.FetchUnsupportedConfigsFromCluster(&cfg.ClusterConfiguration, client)
			if err != nil {
				return err
			}

			// We need to make sure, that our output is predictable, but the maps in Go are unordered.
			// Hence, we have to extract and sort the map keys.
			gvks := make([]schema.GroupVersionKind, 0, len(docmap))
			for gvk := range docmap {
				gvks = append(gvks, gvk)
			}
			sort.Slice(gvks, func(i, j int) bool {
				return gvks[i].String() < gvks[j].String()
			})

			// Finally, use the sorted keys to fetch the unsupported YAML and print it
			for _, gvk := range gvks {
				// Don't forget the YAML document separator. It has a trailing '\n' char, so we use just Fprint here
				fmt.Fprint(out, constants.YAMLDocumentSeparator)

				// Output the YAML document, while making sure that we don't have any spurious leading and/or trailing spaces
				fmt.Fprintln(out, strings.TrimSpace(string(docmap[gvk])))
			}

			return nil
		},
	}

	return cmd
}
