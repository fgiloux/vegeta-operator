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

import (
	"strconv"

	vegetav1alpha1 "github.com/fgiloux/vegeta-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	containerName = "vegeta"
	resultsPath   = "/results"
	bodyPath      = "/opt/config/body.txt"
)

// aPod4Attack generates the definition of the attack pod
func (r *VegetaReconciler) aPod4Attack(v *vegetav1alpha1.Vegeta) *corev1.Pod {
	volumes, mounts := getAPVolumesAndMounts(v)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: v.Name,
			Namespace:    v.Namespace,
			Labels:       r.Labels.LabelsMap,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:        r.Image,
				Name:         containerName,
				Command:      getAttackCmd(v),
				Resources:    v.Spec.Resources,
				VolumeMounts: mounts,
				// TODO: I am not sure this needs to be made configurable. What is defined in the image should be just fine.
				WorkingDir: resultsPath,
			}},
			RestartPolicy: "Never",
			Volumes:       volumes,
		},
	}

	return pod
}

// getAttackCmd assembles an attack command based on the parameters configured in the vegeta resource
func getAttackCmd(veg *vegetav1alpha1.Vegeta) []string {

	/* TODO

	   Mounts:
	   // - BodyConfigMap body.txt Specifies a config map containing the body of every request unless overridden per attack target.
	   // - CertSecret client.crt Specifies the secret containing the TLS client PEM encoded certificate file.
	   // - HeadersConfigMap headers.txt Specifies a config map containing request headers to be used in all targets defined
	   // - KeySecret client.key Specifies the secret containing the PEM encoded TLS client certificate private key
	   // - RootCertsConfigMap *.crt
	   	// - TargetsConfigMap targets.json or targets.http (depending on format)
	*/

	cmd := []string{}
	if veg.Spec.Attack.Target != "" {
		cmd = append(cmd, "echo", `"GET "`+veg.Spec.Attack.Target+`"`, "|")
	}
	cmd = append(cmd, "vegeta", "attack")

	if veg.Spec.Attack.BodyConfigMap != "" {
		cmd = append(cmd, "-body", bodyPath)
	}

	if veg.Spec.Attack.Chunked {
		cmd = append(cmd, "-chunked")
	}

	if veg.Spec.Attack.Connections > 0 {
		cmd = append(cmd, "-connections", strconv.Itoa(veg.Spec.Attack.Connections))
	}

	if veg.Spec.Attack.Duration != "" {
		cmd = append(cmd, "-duration", veg.Spec.Attack.Duration)
	}

	if veg.Spec.Attack.Format != "" {
		cmd = append(cmd, "-format", veg.Spec.Attack.Format.String())
	}

	if veg.Spec.Attack.H2C {
		cmd = append(cmd, "-h2c")
	}

	if veg.Spec.Attack.HTTP2 {
		cmd = append(cmd, "-http2")
	}

	if veg.Spec.Attack.Insecure {
		cmd = append(cmd, "-insecure")
	}

	if veg.Spec.Attack.KeepAlive {
		cmd = append(cmd, "-keepalive")
	}

	if veg.Spec.Attack.Lazy {
		cmd = append(cmd, "-lazy")
	}

	if veg.Spec.Attack.MaxBody != 0 {
		cmd = append(cmd, "-max-body", strconv.FormatUint(uint64(veg.Spec.Attack.MaxBody), 10))
	}

	if veg.Spec.Attack.MaxWorkers != 0 {
		cmd = append(cmd, "-max-workers", strconv.FormatUint(uint64(veg.Spec.Attack.MaxWorkers), 10))
	}

	if veg.Spec.Attack.Name != "" {
		cmd = append(cmd, "-name", veg.Spec.Attack.Name)
	}

	if veg.Spec.Attack.ProxyHeader != "" {
		cmd = append(cmd, "-proxy-header", veg.Spec.Attack.ProxyHeader)
	}

	if veg.Spec.Attack.Rate != "" {
		cmd = append(cmd, "-rate", veg.Spec.Attack.Rate)
	}

	if veg.Spec.Attack.Redirects != 0 {
		cmd = append(cmd, "-redirects", strconv.Itoa(veg.Spec.Attack.Redirects))
	}

	if veg.Spec.Attack.Timeout != "" {
		cmd = append(cmd, "-timeout", veg.Spec.Attack.Timeout)
	}

	if veg.Spec.Attack.Workers > 0 {
		cmd = append(cmd, "-workers", strconv.FormatUint(uint64(veg.Spec.Attack.Workers), 10))
	}

	// In case of results being sent to standard ouptut the report should be processed immediately. There is no way to process it afterwards.
	// TODO: use the enum
	if veg.Spec.Report.OutputType == "" || veg.Spec.Report.OutputType == vegetav1alpha1.StdoutOutput {
		cmd = append(cmd, "|")
		cmd = append(cmd, getReportCmd(veg)...)
	}

	// TODO:
	// The report command accepts multiple result files. It'll read and sort them by timestamp before generating reports.
	// For supporting distributed attacks this means that it is best to process the reports separately.
	// This will be done by a pod, which gets launched after the attack pods have successfully completed
	return cmd
}

// getReportCmd generates the report command  based on the parameters configured in the vegeta resource
func getReportCmd(veg *vegetav1alpha1.Vegeta) []string {
	cmd := []string{}
	cmd = append(cmd, "vegeta", "report")

	if veg.Spec.Report.Buckets != "" {
		cmd = append(cmd, "-buckets", veg.Spec.Report.Buckets)
	}

	if veg.Spec.Report.Every != "" {
		cmd = append(cmd, "-every", veg.Spec.Report.Every)
	}

	if veg.Spec.Report.OutputType != "" && veg.Spec.Report.OutputType != vegetav1alpha1.StdoutOutput {
		// TODO: I am only generating reports in binary format. I may need to encode them in one of the available formats: (gob | json | csv)
		cmd = append(cmd, "-output")
		cmd = append(cmd, getReportFileName(veg)...)
	}

	if veg.Spec.Report.Type.String() != "" {
		cmd = append(cmd, "-type", veg.Spec.Report.Type.String())
	}
	return cmd
}

// getReportFileName generates the name of the report file
func getReportFileName(veg *vegetav1alpha1.Vegeta) []string {
	fname := []string{}
	fname = append(fname, veg.GetName(), "-$(hostname)-$(date +%s).gob")
	return fname
}

// getAPVolumesAndMounts generates the list of volumes and mounts for the attack pod
func getAPVolumesAndMounts(veg *vegetav1alpha1.Vegeta) ([]corev1.Volume, []corev1.VolumeMount) {
	// TODO: make vegeta report storage configurable as per the related field in the vegeta resource
	// default mode 644 should be fine for results ((RW by ownwer RO by others)
	// for configmaps and secrets I may want to mount them readonly: volumeMount.readOnly: true
	// Volume: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#volume-v1-core
	// ConfigMap: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#configmapvolumesource-v1-core
	// Secret: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#secretvolumesource-v1-core
	// EmptyDir: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#emptydirvolumesource-v1-core
	// VolumeMount: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#volumemount-v1-core
	//
	// reports PV to be mounted RW under /reports
	// configs CM RO under /opt/config/ (except RootCertsConfigMap)
	// CA CM RO under /opt/config/credentials
	// secrets under /opt/config/credentials
	// Fields (all optionals):
	// - BodyConfigMap body.txt Specifies a config map containing the body of every request unless overridden per attack target.
	// - CertSecret client.crt Specifies the secret containing the TLS client PEM encoded certificate file.
	// - HeadersConfigMap headers.txt Specifies a config map containing request headers to be used in all targets defined
	// - KeySecret client.key Specifies the secret containing the PEM encoded TLS client certificate private key
	// - RootCertsConfigMap *.crt
	//  volumes:
	//  - name: trusted-ca
	//  configMap:
	//    name: trusted-ca
	//    items:
	// 	 - key: ca-bundle.crt
	// 	   path: tls-ca-bundle.pem
	//  volumeMounts:
	//    - name: trusted-ca
	// 	 mountPath: /etc/pki/ca-trust/extracted/pem
	// 	 readOnly: true
	// - TargetsConfigMap targets.json or targets.http (depending on format)

	// TODO: check outputType and adapt the code for storage of results accordingly.
	volumes := []corev1.Volume{
		{
			Name: "vegeta-results",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	mounts := []corev1.VolumeMount{
		{
			Name:      "vegeta-results",
			MountPath: resultsPath,
		},
	}

	volumes = append(volumes, corev1.Volume{
		Name: "vegeta-results",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	return volumes, mounts
}
