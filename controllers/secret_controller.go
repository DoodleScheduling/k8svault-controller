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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1beta1 "github.com/DoodleScheduling/k8svault-controller/pkg/apis/infra.doodle.com/v1beta1"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get;update;patch

// Reconcile secrets
func (r *SecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("Namespace", req.Namespace, "Name", req.NamespacedName)
	logger.Info("Reconciling Secret")

	// Fetch the Secret instance
	instance := &corev1.Secret{}

	// Operate on a copy
	desired := instance.DeepCopy()

	err := r.Client.Get(context.TODO(), req.NamespacedName, desired)
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

	if _, ok := desired.Annotations[infrav1beta1.AnnotationPath]; ok {
		m, err := mapFromSecret(desired)

		// Annotation parser failed, return early
		if err != nil {
			logger.Error(err, "failed to parse annotations")
			return reconcile.Result{}, err
		}

		logger.Info("Parsed annotation into mapping", "mapping", m)
		v, err := FromMapping(m, logger)

		// Failed to setup vault client, requeue immediately
		if err != nil {
			logger.Error(err, "Failed to setup vault client")
			return reconcile.Result{Requeue: true}, err
		}

		err = v.ApplySecret(m, desired, r.Recorder)

		// Failed applying state to vault, requeue immediately
		if err != nil {
			logger.Error(err, "Failed apply desired state to vault")
			r.Recorder.Event(instance, "Normal", "error", fmt.Sprintf("Failed to sync secret to vault %s/%s: %s", instance.Namespace, instance.Name, err.Error()))
			return reconcile.Result{Requeue: true}, err
		}
	}

	// Return if secret does not have necessary annotation
	return ctrl.Result{}, nil
}

// SetupWithManager adding controllers
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
