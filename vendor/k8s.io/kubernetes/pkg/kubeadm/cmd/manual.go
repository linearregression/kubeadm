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
	"fmt"
	"io"

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	kubeadmapi "k8s.io/kubernetes/pkg/kubeadm/api"
	kubemaster "k8s.io/kubernetes/pkg/kubeadm/master"
	kubenode "k8s.io/kubernetes/pkg/kubeadm/node"
	kubeadmutil "k8s.io/kubernetes/pkg/kubeadm/util"
	// TODO: cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdManual(out io.Writer, params *kubeadmapi.BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manual",
		Short: "Advanced, less-automated functionality, for power users.",
		// TODO put example usage in the Long description here
	}
	cmd.AddCommand(NewCmdManualBootstrap(out, params))
	return cmd
}

func NewCmdManualBootstrap(out io.Writer, params *kubeadmapi.BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Manually bootstrap a cluster 'out-of-band'",
		Long: dedent.Dedent(`
			Manually bootstrap a cluster 'out-of-band', by generating and distributing a CA
			certificate to all your servers and specifying and (list of) API server URLs.
		`),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.AddCommand(NewCmdManualBootstrapInitMaster(out, params))
	cmd.AddCommand(NewCmdManualBootstrapJoinNode(out, params))

	return cmd
}

func NewCmdManualBootstrapInitMaster(out io.Writer, params *kubeadmapi.BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-master",
		Short: "Manually bootstrap a master 'out-of-band'",
		Long: dedent.Dedent(`
			Manually bootstrap a master 'out-of-band'.
			Will create TLS certificates and set up static pods for Kubernetes master
			components.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if params.Discovery.ListenIP == "" {
				ip, err := kubeadmutil.GetDefaultHostIP()
				if err != nil {
					return err
				}
				params.Discovery.ListenIP = ip
			}
			if err := kubemaster.CreateTokenAuthFile(params); err != nil {
				return err
			}
			if err := kubemaster.WriteStaticPodManifests(params); err != nil {
				return err
			}
			kubeconfigs, err := kubemaster.CreatePKIAssets(params)
			if err != nil {
				return err
			}

			for name, kubeconfig := range kubeconfigs {
				if err := kubeadmutil.WriteKubeconfigIfNotExists(params, name, kubeconfig); err != nil {
					out.Write([]byte(fmt.Sprintf("Unable to write admin for master:\n%s\n", err)))
					return nil
				}
			}

			out.Write([]byte(fmt.Sprintf(dedent.Dedent(`
				Static pods written and kubelet's kubeconfig written.
				Kubelet should be able to start soon (try 'systemctl restart kubelet' or equivalent
				if it doesn't). CA cert is written to /etc/kubernetes/pki/ca.pem. Please copy this file
				(scp, rsync or through other means) to all your nodes and then run on them:
				kubeadm manual bootstrap node --ca-cert-file <path-to-ca-cert> --token %s --api-server-urls https://%s:443/

			`),
				params.Discovery.BearerToken, params.Discovery.ListenIP,
			)))

			return nil
		},
	}

	params.Discovery.ApiServerURLs = "http://127.0.0.1:8080/" // On the master, assume you can talk to the API server
	cmd.PersistentFlags().StringVarP(&params.Discovery.ApiServerDNSName, "api-dns-name", "", "",
		`(optional) DNS name for the API server, will be encoded into
		subjectAltName in the resulting (generated) TLS certificates`)
	cmd.PersistentFlags().StringVarP(&params.Discovery.ListenIP, "listen-ip", "", "",
		`(optional) IP address to listen on, in case autodetection fails.`)
	cmd.PersistentFlags().StringVarP(&params.Discovery.BearerToken, "token", "", "",
		`(optional) Shared secret used to secure bootstrap. Will be generated and displayed if not provided.`)

	return cmd
}

func NewCmdManualBootstrapJoinNode(out io.Writer, params *kubeadmapi.BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join-node",
		Short: "Manually bootstrap a node 'out-of-band', joining it into a cluster with extant control plane",

		Run: func(cmd *cobra.Command, args []string) {
			if params.Discovery.CaCertFile == "" {
				out.Write([]byte(fmt.Sprintf("Must specify --ca-cert-file (see --help)\n")))
				return
			}

			if params.Discovery.ApiServerURLs == "" {
				out.Write([]byte(fmt.Sprintf("Must specify --api-server-urls (see --help)\n")))
				return
			}

			kubeconfig, err := kubenode.PerformTLSBootstrap(params)
			if err != nil {
				out.Write([]byte(fmt.Sprintf("Failed to perform TLS bootstrap: %s\n", err)))
				return
			}
			fmt.Println("recieved signed certificate from the API server, will write `/etc/kubernetes/kubelet.conf`...")

			err = kubeadmutil.WriteKubeconfigIfNotExists(params, "kubelet", kubeconfig)
			if err != nil {
				out.Write([]byte(fmt.Sprintf("Unable to write config for node:\n%s\n", err)))
				return
			}
			out.Write([]byte(dedent.Dedent(`
			    Kubelet informed of new config.
			    Run 'kubectl get nodes' on the master to see it join.
			`)))
		},
	}
	cmd.PersistentFlags().StringVarP(&params.Discovery.CaCertFile, "ca-cert-file", "", "",
		`Path to a CA cert file in PEM format. The same CA cert must be distributed to
		all servers.`)
	cmd.PersistentFlags().StringVarP(&params.Discovery.ApiServerURLs, "api-server-urls", "", "",
		`Comma separated list of API server URLs. Typically this might be just
		https://<address-of-master>:8080/`)
	cmd.PersistentFlags().StringVarP(&params.Discovery.BearerToken, "token", "", "",
		`Shared secret used to secure bootstrap. Must match output of 'init-master'.`)

	return cmd
}
