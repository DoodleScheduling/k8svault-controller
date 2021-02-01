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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

// VaultClaim reconciles a VaultClaim object
type VaultClaimReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=VaultClaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=VaultClaims/status,verbs=get;update;patch

// Reconcile VaultClaims
func (r *VaultClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("Namespace", req.Namespace, "Name", req.NamespacedName)
	logger.Info("Reconciling VaultClaim")

	// Fetch the VaultClaim instance
	claim := &v1beta1.VaultClaim{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, claim)
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

	claim, result, err := r.reconcile(ctx, claim, logger)

	// Update status after reconciliation.
	if err = r.patchStatus(ctx, claim); err != nil {
		log.Error(err, "unable to update status after reconciliation")
		return ctrl.Result{Requeue: true}, err
	}

	// Return if VaultClaim does not have necessary annotation
	return result, nil
}

func (r *VaultClaimReconciler) reconcile(ctx context.Context, claim *v1beta1.VaultClaim, logger logr.Logger) (*v1beta1.VaultClaim, ctrl.Result, error) {
	// Fetch referencing secret
	secret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), claim.Spec.Secret.Name, secret)

	// Failed to fetch referenced secret, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Referencing secret was not found: %s", err.Error())
		r.Recorder.Event(claim, "Normal", events.EventSeverityError, msg)
		return v1beta1.VaultClaimNotBound(claim, v1beta1.SecretNotFoundReason, msg), ctrl.Result{Requeue: true}, reconcileErr
	}

	h, err := FromClaim(claim, logger)

	// Failed to setup vault client, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Connection to vault failed: %s", err.Error())
		r.Recorder.Event(claim, "Normal", events.EventSeverityError, msg)
		return v1beta1.VaultClaimNotBound(claim, v1beta1.VaultConnectionFailedReason, msg), ctrl.Result{Requeue: true}, reconcileErr
	}

	needUpdate, err = h.ApplyVaultClaim(claim, secret)

	// Failed to setup vault client, requeue immediately
	if err != nil {
		msg := fmt.Sprintf("Update vault failed: %s", err.Error())
		r.Recorder.Event(claim, "Normal", events.EventSeverityError, msg)
		return v1beta1.VaultClaimNotBound(claim, v1beta1.VaultUpdateFailedReason, msg), ctrl.Result{Requeue: true}, reconcileErr
	}

	if needUpdate == true {
		msg := "Vault fields successfully bound"
		r.Recorder.Event(claim, "Normal", events.EventSeverityInfo, msg)
		return v1beta1.VaultClaimBound(claim, v1beta1.VaultUpdateSuccessfulReason, msg), ctrl.Result{}, nil
	}

	return claim, ctrl.Result{}, err
}

func (r *VaultClaimReconciler) patchStatus(ctx context.Context, claim *v1beta1.VaultClaim) error {
	key := r.client.ObjectKeyFromObject(claim)
	latest := &v1beta1.VaultClaim{}
	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}
	return r.Client.Status().Patch(ctx, claim, client.MergeFrom(latest))
}

// SetupWithManager adding controllers
func (r *VaultClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.VaultClaim{}).
		Complete(r)
}
