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
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	infrav1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("VaultBindingReconciler", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Second * 1
	)

	Context("VaultBinding", func() {
		var (
			namespace *corev1.Namespace
			//			container *vaultContainer
			err error
		//	tokenFile string
		)

		_, err = setupvaultContainer(context.TODO())
		Expect(err).NotTo(HaveOccurred(), "failed to start vault container")

		BeforeEach(func() {
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "vaultbinding-" + randStringRunes(5)},
			}
			err = k8sClient.Create(context.Background(), namespace)
			Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

			file, err := os.CreateTemp(os.TempDir(), "jwt")
			Expect(err).NotTo(HaveOccurred(), "failed to create temp jwt file")
			defer os.Remove(file.Name())
		})

		AfterEach(func() {
			Eventually(func() error {
				return k8sClient.Delete(context.Background(), namespace)
			}, timeout, interval).Should(Succeed(), "failed to delete test namespace")
		})

		It("fails if referenced secret is not found", func() {
			key := types.NamespacedName{
				Name:      "vaultbinding-" + randStringRunes(5),
				Namespace: namespace.Name,
			}
			created := &infrav1beta1.VaultBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: infrav1beta1.VaultBindingSpec{
					VaultSpec: &infrav1beta1.VaultSpec{
						Address: "https://does-not-exists",
						Path:    "/dest/not-found",
					},
					Secret: &corev1.SecretReference{
						Name: "does-not-exists",
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			got := &infrav1beta1.VaultBinding{}
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), key, got)
				return len(got.Status.Conditions) == 1 &&
					got.Status.Conditions[0].Reason == infrav1beta1.SecretNotFoundReason &&
					got.Status.Conditions[0].Status == "False" &&
					got.Status.Conditions[0].Type == infrav1beta1.BoundCondition
			}, timeout, interval).Should(BeTrue())
		})

		It("fails if vault can't be contacted", func() {
			By("Adding secret")
			keySecret := types.NamespacedName{
				Name:      "secret-" + randStringRunes(5),
				Namespace: namespace.Name,
			}

			secret := randStringRunes(5)
			createdSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      keySecret.Name,
					Namespace: keySecret.Namespace,
				},
				Data: map[string][]byte{
					"berries": []byte(secret),
				},
			}

			Expect(k8sClient.Create(context.Background(), createdSecret)).Should(Succeed())

			key := types.NamespacedName{
				Name:      "vaultbinding-" + randStringRunes(5),
				Namespace: namespace.Name,
			}
			created := &infrav1beta1.VaultBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: infrav1beta1.VaultBindingSpec{
					VaultSpec: &infrav1beta1.VaultSpec{
						Address: "https://does-not-exists",
						Path:    "/dest/not-found",
					},
					Secret: &corev1.SecretReference{
						Name: keySecret.Name,
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			got := &infrav1beta1.VaultBinding{}
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), key, got)
				return len(got.Status.Conditions) == 1 &&
					got.Status.Conditions[0].Reason == infrav1beta1.VaultConnectionFailedReason &&
					got.Status.Conditions[0].Status == "False" &&
					got.Status.Conditions[0].Type == infrav1beta1.BoundCondition
			}, timeout, interval).Should(BeTrue())
		})
	})
})
