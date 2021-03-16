/*
Copyright 2020.

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

//

import (
	"context"
	"fmt"
	"time"

	"github.com/fgiloux/vegeta-operator/api/v1alpha1"
	vegetav1alpha1 "github.com/fgiloux/vegeta-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Vegeta controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		VegetaName = "test-vegeta"

		timeout = time.Second * 10
		// duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When an attack is performed with a successful pod", func() {
		It("Should update Vegeta.Status", func() {
			By("Creation")
			ctx := context.Background()
			vegeta := newVegeta(VegetaName)
			Expect(k8sClient.Create(ctx, vegeta)).Should(Succeed())
			msg := fmt.Sprintf("Name: %s, Namespage: %s \n", vegeta.Name, vegeta.Namespace)
			GinkgoWriter.Write([]byte(msg))
			vLookupKey := types.NamespacedName{Name: VegetaName, Namespace: TestNs}
			createdVegeta := &vegetav1alpha1.Vegeta{}

			// Creation may not immediately happen.
			Eventually(func() v1alpha1.PhaseEnum {
				_ = k8sClient.Get(ctx, vLookupKey, createdVegeta)
				return createdVegeta.Status.Phase
			}, timeout, interval).Should(Equal(v1alpha1.RunningPhase))
			Expect(createdVegeta.Spec.Replicas).Should(Equal(uint(1)))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(1))

			By("Completion")
			createdPod := &corev1.Pod{}
			podLookupKey := types.NamespacedName{Name: createdVegeta.Status.Active[0], Namespace: TestNs}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			createdPod.Status.Phase = corev1.PodSucceeded
			Expect(k8sClient.Status().Update(ctx, createdPod)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, vLookupKey, createdVegeta)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Eventually(func() v1alpha1.PhaseEnum {
				_ = k8sClient.Get(ctx, vLookupKey, createdVegeta)
				return createdVegeta.Status.Phase
			}, timeout, interval).Should(Equal(v1alpha1.SucceededPhase))
			msg = fmt.Sprintf("Vegeta phase: %s\n", createdVegeta.Status.Phase)
			GinkgoWriter.Write([]byte(msg))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(0))
			Expect(len(createdVegeta.Status.Succeeded)).Should(Equal(1))
		})
	})
	Context("When an attack is performed with a successful and an unsuccessful pod", func() {
		It("Should update Vegeta.Status", func() {
			By("Creation")
			ctx := context.Background()
			vegeta := newVegeta(VegetaName + "-fail")
			vegeta.Spec.Replicas = 2
			Expect(k8sClient.Create(ctx, vegeta)).Should(Succeed())
			msg := fmt.Sprintf("Name: %s, Namespage: %s \n", vegeta.Name, vegeta.Namespace)
			GinkgoWriter.Write([]byte(msg))
			vLookupKey := types.NamespacedName{Name: vegeta.Name, Namespace: TestNs}
			createdVegeta := &vegetav1alpha1.Vegeta{}

			// Creation may not immediately happen.
			Eventually(func() v1alpha1.PhaseEnum {
				_ = k8sClient.Get(ctx, vLookupKey, createdVegeta)
				return createdVegeta.Status.Phase
			}, timeout, interval).Should(Equal(v1alpha1.RunningPhase))
			Expect(createdVegeta.Spec.Replicas).Should(Equal(uint(2)))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(2))
			for _, podName := range createdVegeta.Status.Active {
				msg := fmt.Sprintf("Active pods: %s \n", podName)
				GinkgoWriter.Write([]byte(msg))
			}

			By("Failure")
			createdPod := &corev1.Pod{}
			podLookupKey := types.NamespacedName{Name: createdVegeta.Status.Active[0], Namespace: TestNs}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			createdPod.Status.Phase = corev1.PodFailed
			Expect(k8sClient.Status().Update(ctx, createdPod)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, vLookupKey, createdVegeta)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Eventually(func() v1alpha1.PhaseEnum {
				_ = k8sClient.Get(ctx, vLookupKey, createdVegeta)
				return createdVegeta.Status.Phase
			}, timeout, interval).Should(Equal(v1alpha1.FailedPhase))
			msg = fmt.Sprintf("Vegeta phase: %s\n", createdVegeta.Status.Phase)
			GinkgoWriter.Write([]byte(msg))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(1))
			Expect(len(createdVegeta.Status.Succeeded)).Should(Equal(0))
			Expect(len(createdVegeta.Status.Failed)).Should(Equal(1))

			// A second successful pod should NOT impact the status
			By("Success after Failure")
			podLookupKey = types.NamespacedName{Name: createdVegeta.Status.Active[0], Namespace: TestNs}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			createdPod.Status.Phase = corev1.PodSucceeded
			Expect(k8sClient.Status().Update(ctx, createdPod)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, vLookupKey, createdVegeta)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Eventually(func() int {
				_ = k8sClient.Get(ctx, vLookupKey, createdVegeta)
				return len(createdVegeta.Status.Succeeded)
			}, timeout, interval).Should(Equal(1))
			msg = fmt.Sprintf("Vegeta phase: %s\n", createdVegeta.Status.Phase)
			GinkgoWriter.Write([]byte(msg))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(0))
			Expect(len(createdVegeta.Status.Failed)).Should(Equal(1))
			Expect(createdVegeta.Status.Phase).Should(Equal(v1alpha1.FailedPhase))
		})
	})
	Context("When RootCertsConfigMap is specified", func() {
		It("Should create pods mounting the matching config map", func() {
			By("Creation of the vegeta resource")
			ctx := context.Background()
			vegeta := newVegeta(VegetaName + "-cm")
			vegeta.Spec.Attack.RootCertsConfigMap = "rootcerts"
			Expect(k8sClient.Create(ctx, vegeta)).Should(Succeed())
			msg := fmt.Sprintf("Name: %s, Namespage: %s \n", vegeta.Name, vegeta.Namespace)
			GinkgoWriter.Write([]byte(msg))
			vLookupKey := types.NamespacedName{Name: vegeta.Name, Namespace: TestNs}
			createdVegeta := &vegetav1alpha1.Vegeta{}

			// Creation may not immediately happen.
			Eventually(func() v1alpha1.PhaseEnum {
				_ = k8sClient.Get(ctx, vLookupKey, createdVegeta)
				return createdVegeta.Status.Phase
			}, timeout, interval).Should(Equal(v1alpha1.RunningPhase))
			Expect(createdVegeta.Spec.Replicas).Should(Equal(uint(1)))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(1))

			By("Creation of the pod")
			createdPod := &corev1.Pod{}
			podLookupKey := types.NamespacedName{Name: createdVegeta.Status.Active[0], Namespace: TestNs}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdPod.Spec.Volumes[1].Name).Should(Equal("trusted-ca"))
		})
	})
})

func newVegeta(name string) *vegetav1alpha1.Vegeta {
	return &vegetav1alpha1.Vegeta{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "vegeta.vegeta.testing.io/v1alpha1",
			Kind:       "Vegeta",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: TestNs,
		},
		Spec: vegetav1alpha1.VegetaSpec{
			Attack: &vegetav1alpha1.AttackSpec{
				Duration: "10s",
				Rate:     "1s",
				Target:   "",
			},
			Replicas: 1,
		},
	}
}
