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

package kubemaster

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	// TODO: "k8s.io/client-go/client/tools/clientcmd/api"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kubeadmapi "k8s.io/kubernetes/pkg/kubeadm/api"
	"k8s.io/kubernetes/pkg/kubeadm/tlsutil"
	kubeadmutil "k8s.io/kubernetes/pkg/kubeadm/util"
)

func createClientCertsAndConfigs(params *kubeadmapi.BootstrapParams, clientNames []string, caCert *x509.Certificate, caKey *rsa.PrivateKey) (map[string]*clientcmdapi.Config, error) {

	basicClientConfig := kubeadmutil.CreateBasicClientConfig(
		"kubernetes",
		fmt.Sprintf("https://%s:443", params.Discovery.ListenIP),
		tlsutil.EncodeCertificatePEM(caCert),
	)

	configs := map[string]*clientcmdapi.Config{}

	for _, client := range []string{"kubelet", "admin"} {
		key, cert, err := newClientKeyAndCert(caCert, caKey)
		if err != nil {
			return nil, err
		}
		config := kubeadmutil.MakeClientConfigWithCerts(
			basicClientConfig,
			"kubernetes",
			client,
			tlsutil.EncodePrivateKeyPEM(key),
			tlsutil.EncodeCertificatePEM(cert),
		)
		configs[client] = config
	}

	return configs, nil
}
