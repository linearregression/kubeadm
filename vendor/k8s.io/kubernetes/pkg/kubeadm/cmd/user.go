/*
Copyright 2016 The Kubernetes Authors.

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

package kubecmd

import (
	"io"

	"github.com/spf13/cobra"
	kubeadmapi "k8s.io/kubernetes/pkg/kubeadm/api"
)

func NewCmdUser(out io.Writer, params *kubeadmapi.BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Get initial admin credentials for a cluster.", // using TLS bootstrap
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	return cmd
}
