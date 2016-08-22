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

package kubeadmapi

type BootstrapParams struct {
	// A struct with methods that implement Discover()
	// kubeadm will do the CSR dance
	Discovery *OutOfBandDiscovery
	EnvParams map[string]string
}

type OutOfBandDiscovery struct {
	// 'join-node' side
	ApiServerURLs string // comma separated
	CaCertFile    string
	BearerToken   string // optional on master side, will be generated if not specified
	// 'init-master' side
	ApiServerDNSName string // optional, used in master bootstrap
	ListenIP         string // optional IP for master to listen on, rather than autodetect
}
