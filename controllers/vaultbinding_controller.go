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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

const (
	// secretIndexKey is the key used for indexing VaultBindings based on
	// their secrets.
	secretIndexKey string = ".metadata.secret"
)

// VaultBinding reconciles a VaultBinding object
type VaultBindingReconciler struct {
	client.Client
	indexer  client.FieldIndexer
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

type VaultBindingReconcilerOptions struct {
	MaxConcurrentReconciles int
}

// SetupWithManager adding controllers
func (r *VaultBindingReconciler) SetupWithManager(mgr ctrl.Manager, opts VaultBindingReconcilerOptions) error {
	// Index the VaultBinding by the Secret references they point at
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &v1beta1.VaultBinding{}, secretIndexKey,
		func(o client.Object) []string {
			vb := o.(*v1beta1.VaultBinding)
			r.Log.Info(fmt.Sprintf("%s/%s", vb.GetNamespace(), vb.Spec.Secret.Name))
			return []string{
				fmt.Sprintf("%s/%s", vb.GetNamespace(), vb.Spec.Secret.Name),
			}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.VaultBinding{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.requestsForSecretChange),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: opts.MaxConcurrentReconciles}).
		Complete(r)
}

func (r *VaultBindingReconciler) requestsForSecretChange(o client.Object) []reconcile.Request {
	s, ok := o.(*corev1.Secret)
	if !ok {
		panic(fmt.Sprintf("expected a Secret, got %T", o))
	}

	ctx := context.Background()
	var list v1beta1.VaultBindingList
	if err := r.List(ctx, &list, client.MatchingFields{
		secretIndexKey: objectKey(s).String(),
	}); err != nil {
		return nil
	}

	var reqs []reconcile.Request
	for _, i := range list.Items {
		r.Log.Info("referenced secret from a vaultbinding changed detected, reconcile binding", "namespace", i.GetNamespace(), "name", i.GetName())
		reqs = append(reqs, reconcile.Request{NamespacedName: objectKey(&i)})
	}

	return reqs
}

// +kubebuilder:rbac:groups=core,resources=VaultBindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=VaultBindings/status,verbs=get;update;patch

// Reconcile VaultBindings
func (r *VaultBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Namespace", req.Namespace, "Name", req.NamespacedName)
	logger.Info("reconciling VaultBinding")

	// Fetch the VaultBinding instance
	binding := v1beta1.VaultBinding{}

	err := r.Client.Get(ctx, req.NamespacedName, &binding)
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

	binding, result, reconcileErr := r.reconcile(ctx, binding, logger)

	// Update status after reconciliation.
	if err = r.patchStatus(ctx, &binding); err != nil {
		log.Error(err, "unable to update status after reconciliation")
		return ctrl.Result{Requeue: true}, err
	}

	return result, reconcileErr
}

func (r *VaultBindingReconciler) reconcile(ctx context.Context, binding v1beta1.VaultBinding, logger logr.Logger) (v1beta1.VaultBinding, ctrl.Result, error) {
	// Fetch referencing secret
	secret := &corev1.Secret{}
	secretName := types.NamespacedName{
		Namespace: binding.GetNamespace(),
		Name:      binding.Spec.Secret.Name,
	}
	err := r.Client.Get(context.TODO(), secretName, secret)

	// Failed to fetch referenced secret, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Referencing secret was not found: %s", err.Error())
		r.Recorder.Event(&binding, "Normal", "error", msg)
		return v1beta1.VaultBindingNotBound(binding, v1beta1.SecretNotFoundReason, msg), ctrl.Result{Requeue: true}, err
	}

	h, err := FromBinding(&binding, logger)

	// Failed to setup vault client, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Connection to vault failed: %s", err.Error())
		r.Recorder.Event(&binding, "Normal", "error", msg)
		return v1beta1.VaultBindingNotBound(binding, v1beta1.VaultConnectionFailedReason, msg), ctrl.Result{Requeue: true}, err
	}

	_, err = h.ApplySecret(&binding, secret)

	// Failed to setup vault client, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Update vault failed: %s", err.Error())
		r.Recorder.Event(&binding, "Normal", "error", msg)
		return v1beta1.VaultBindingNotBound(binding, v1beta1.VaultUpdateFailedReason, msg), ctrl.Result{Requeue: true}, err
	}

	msg := "Vault fields successfully bound"
	r.Recorder.Event(&binding, "Normal", "info", msg)
	return v1beta1.VaultBindingBound(binding, v1beta1.VaultUpdateSuccessfulReason, msg), ctrl.Result{}, err
}

func (r *VaultBindingReconciler) patchStatus(ctx context.Context, binding *v1beta1.VaultBinding) error {
	key := client.ObjectKeyFromObject(binding)
	latest := &v1beta1.VaultBinding{}
	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}

	return r.Client.Status().Patch(ctx, binding, client.MergeFrom(latest))
}

// objectKey returns client.ObjectKey for the object.
func objectKey(object metav1.Object) client.ObjectKey {
	return client.ObjectKey{
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}
}
