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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

// VaultMirror reconciles a VaultMirror object
type VaultMirrorReconciler struct {
	client.Client
	indexer  client.FieldIndexer
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

type VaultMirrorReconcilerOptions struct {
	MaxConcurrentReconciles int
}

// SetupWithManager adding controllers
func (r *VaultMirrorReconciler) SetupWithManager(mgr ctrl.Manager, opts VaultMirrorReconcilerOptions) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.VaultMirror{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: opts.MaxConcurrentReconciles}).
		Complete(r)
}

// +kubebuilder:rbac:groups=core,resources=VaultMirrors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=VaultMirrors/status,verbs=get;update;patch

// Reconcile VaultMirrors
func (r *VaultMirrorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Namespace", req.Namespace, "Name", req.NamespacedName)
	logger.Info("reconciling VaultMirror")

	// Fetch the VaultMirror instance
	binding := v1beta1.VaultMirror{}

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

func (r *VaultMirrorReconciler) reconcile(ctx context.Context, binding v1beta1.VaultMirror, logger logr.Logger) (v1beta1.VaultMirror, ctrl.Result, error) {
	h, err := FromBinding(&binding, logger)

	// Failed to setup vault client, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Connection to vault failed: %s", err.Error())
		r.Recorder.Event(&binding, "Normal", "error", msg)
		return v1beta1.VaultMirrorNotBound(binding, v1beta1.VaultConnectionFailedReason, msg), ctrl.Result{Requeue: true}, err
	}

	_, err = h.ApplySecret(&binding, secret)

	// Failed to setup vault client, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Update vault failed: %s", err.Error())
		r.Recorder.Event(&binding, "Normal", "error", msg)
		return v1beta1.VaultMirrorNotBound(binding, v1beta1.VaultUpdateFailedReason, msg), ctrl.Result{Requeue: true}, err
	}

	msg := "Vault fields successfully bound"
	r.Recorder.Event(&binding, "Normal", "info", msg)
	return v1beta1.VaultMirrorBound(binding, v1beta1.VaultUpdateSuccessfulReason, msg), ctrl.Result{}, err
}

func (r *VaultMirrorReconciler) patchStatus(ctx context.Context, binding *v1beta1.VaultMirror) error {
	key := client.ObjectKeyFromObject(binding)
	latest := &v1beta1.VaultMirror{}
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
