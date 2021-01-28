/*


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

package main

import (
	"flag"
	"os"
	"strings"

	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/DoodleScheduling/k8svault-controller/controllers"
	"github.com/prometheus/common/log"
	"github.com/spf13/pflag"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

var (
	metricsAddr             = ":9556"
	enableLeaderElection    = true
	leaderElectionNamespace string
	vaultAddress            string
	tlsInsecure             bool
	tlsServerName           string
	tlsCAPath               string
	tlsCACert               string
	tlsClientCert           string
	tlsClientKey            string
)

func main() {
	flag.StringVar(&metricsAddr, "metrics-addr", ":9556", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace", "",
		"Specify a different leader election namespace. It will use the one where the controller is deployed by default.")
	flag.StringVar(&vaultAddress, "vault-addr", "",
		"Fallback vault server if no one was specified in annotations.")
	flag.BoolVar(&tlsInsecure, "tls-insecure", false,
		"Allow insecure TLS communication to vault (no certificate validation).")
	flag.StringVar(&tlsServerName, "tls-server-name", "",
		"Used to set the SNI host when connecting to vault.")
	flag.StringVar(&tlsCAPath, "tls-capath", "",
		"CAPath is the path to a directory of PEM-encoded CA cert files to verify CAPath string.")
	flag.StringVar(&tlsCACert, "tls-cacert", "",
		"CACert is the path to a PEM-encoded CA cert file used to verify the Vault server SSL certificate.")
	flag.StringVar(&tlsClientCert, "tls-client-cert", "",
		"The path to the certificate for Vault communication.")
	flag.StringVar(&tlsClientKey, "tls-client-key", "",
		"The private key for Vault communication.")

	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Import flags into viper and bind them to env vars
	// flags are converted to upper-case, - is replaced with _
	// secret-length -> SECRET_LENGTH
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Error(err, "Failed parsing command line arguments")
		os.Exit(1)
	}

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      viper.GetString("metrics-addr"),
		Port:                    9555,
		LeaderElection:          viper.GetBool("enable-leader-election"),
		LeaderElectionNamespace: viper.GetString("leader-election-namespace"),
		LeaderElectionID:        "1d602b51.doodle.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	if err != nil {
		log.Error(err, "Could not add liveness probe")
		os.Exit(1)
	}

	// Add readiness probe
	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	if err != nil {
		log.Error(err, "Could not add readiness probe")
		os.Exit(1)
	}

	if err = (&controllers.SecretReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Secret"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Secret")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
