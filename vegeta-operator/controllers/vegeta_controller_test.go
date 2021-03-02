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

	Context("When updating Vegeta Status", func() {
		It("Should increase Vegeta.Status.Active count when new Pods are created", func() {
			By("Creating a new Vegeta resource")
			ctx := context.Background()
			vegeta := newVegeta(VegetaName)
			Expect(k8sClient.Create(ctx, vegeta)).Should(Succeed())
			msg := fmt.Sprintf("Name: %s, Namespage: %s \n", vegeta.Name, vegeta.Namespace)
			GinkgoWriter.Write([]byte(msg))
			vLookupKey := types.NamespacedName{Name: VegetaName, Namespace: TestNs}
			createdVegeta := &vegetav1alpha1.Vegeta{}

			// Creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, vLookupKey, createdVegeta)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdVegeta.Spec.Replicas).Should(Equal(uint(1)))
			Expect(createdVegeta.Status.Phase).Should(Equal(v1alpha1.RunningPhase))
			Expect(len(createdVegeta.Status.Active)).Should(Equal(1))
		})

		It("Should increase Vegeta.Status.Succeeded count when Pods successfully completed", func() {
			By("Checking the number of pods that successfully completed")
		})

		It("Should increase Vegeta.Status.Failed count when Pods failed", func() {
			By("Checking the number of pods that failed")
		})

		It("Should have a Vegeta.Status.Phase set to FailedPhase when Pods failed", func() {
			By("Checking whether a pod failed")
		})

		It("Should have a Vegeta.Status.Phase set to RunningPhase when no Pod failed and Pods are still running", func() {
			By("Checking the number of running and failed pods")
		})

		It("Should have a Vegeta.Status.Phase set to SucceededPhase when all Pods successfully completed", func() {
			By("Checking that the total number of pods equals the number of successful ones")
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
