/*
Copyright AppsCode Inc. and Contributors.

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

package cmds

import (
	catalogv1alpha1 "go.klusters.dev/capi-ops-manager/apis/catalog/v1alpha1"
	opsv1alpha1 "go.klusters.dev/capi-ops-manager/apis/ops/v1alpha1"

	"github.com/spf13/cobra"
	v "gomodules.xyz/x/version"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(scheme))
	utilruntime.Must(opsv1alpha1.AddToScheme(scheme))
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "capi-ops-manager [command]",
		Short:             `CAPI Ops Manager by AppsCode`,
		DisableAutoGenTag: true,
	}

	rootCmd.AddCommand(v.NewCmdVersion())

	ctx := genericapiserver.SetupSignalContext()
	rootCmd.AddCommand(NewCmdOperator(ctx))
	rootCmd.AddCommand(NewCmdWebhook(ctx))

	return rootCmd
}
