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

package controllers

import (
	"context"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	VBReconciler *VaultBindingReconciler
}

type SecretReconcilerOptions struct {
	MaxConcurrentReconciles int
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

// Reconcile VaultBindings
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Namespace", req.Namespace, "Name", req.NamespacedName)
	logger.Info("reconciling Secret")

	// Fetch the Secret instance
	sec := &corev1.Secret{}

	err := r.Client.Get(ctx, req.NamespacedName, sec)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var bindings v1beta1.VaultBindingList
	err = r.List(ctx, &bindings)
	if err != nil {
		logger.Error(err, "failed to lookup matching VaultBindings")

		return ctrl.Result{}, nil
	}

	for _, vb := range bindings.Items {
		if vb.Spec.Secret.Name != sec.GetName() {
			continue
		}

		logger.Info("found referencing VaultBinding, trigger reconile", "vaultbinding", vb.GetName())
		req := ctrl.Request{
			NamespacedName: client.ObjectKeyFromObject(&vb),
		}

		r.VBReconciler.Reconcile(ctx, ctrl.Request(req))
	}

	return ctrl.Result{}, nil
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager, opts SecretReconcilerOptions) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: opts.MaxConcurrentReconciles}).
		Complete(r)
}
