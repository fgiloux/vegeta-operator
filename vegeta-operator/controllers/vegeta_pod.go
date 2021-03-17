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
	"strings"

	vegetav1alpha1 "github.com/fgiloux/vegeta-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	containerName = "vegeta"
	resultsPath   = "/results"
	bodyPath      = "/opt/config/body.txt"
)

// aPod4Attack generates the definition of the attack pod
func (r *VegetaReconciler) aPod4Attack(v *vegetav1alpha1.Vegeta) *corev1.Pod {
	log := r.Log.WithValues("vegeta", v.Namespace)
	immediate := int64(0)
	volumes, mounts := getAPVolumesAndMounts(v)
	var image string
	if vImg := strings.TrimSpace(v.Spec.Image); vImg != "" {
		image = vImg
	} else {
		image = r.Image
	}
	log.V(1).Info("Root certificates", "RootCertsConfigMap", v.Spec.Attack.RootCertsConfigMap)
	log.V(1).Info("Volumes", "Volumes", volumes)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: v.Name + "-",
			Namespace:    v.Namespace,
			Labels: r.Labels.Merge(map[string]string{
				"app.kubernetes.io/name":       "vegeta",
				"app.kubernetes.io/instance":   v.Name,
				"app.kubernetes.io/managed-by": "vegeta-operator"}),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:           image,
				ImagePullPolicy: "Always",
				Name:            containerName,
				Command:         []string{"/bin/sh"},
				Args:            []string{"-c", getAttackCmd(v)},
				Resources:       v.Spec.Resources,
				VolumeMounts:    mounts,
				// TODO: I am not sure this needs to be made configurable. What is defined in the image should be just fine.
				WorkingDir: resultsPath,
			}},
			RestartPolicy:                 "Never",
			Volumes:                       volumes,
			SecurityContext:               &corev1.PodSecurityContext{},
			TerminationGracePeriodSeconds: &immediate,
		},
	}
	// Set Vegeta instance as the owner and controller
	ctrl.SetControllerReference(v, pod, r.Scheme)
	return pod
}

// getAttackCmd assembles an attack command based on the parameters configured in the vegeta resource
func getAttackCmd(veg *vegetav1alpha1.Vegeta) string {

	/* TODO

	   Mounts:
	   // - BodyConfigMap body.txt Specifies a config map containing the body of every request unless overridden per attack target.
	   // - CertSecret client.crt Specifies the secret containing the TLS client PEM encoded certificate file.
	   // - HeadersConfigMap headers.txt Specifies a config map containing request headers to be used in all targets defined
	   // - KeySecret client.key Specifies the secret containing the PEM encoded TLS client certificate private key
	   	// - TargetsConfigMap targets.json or targets.http (depending on format)
	*/

	var sb strings.Builder
	if veg.Spec.Attack.Target != "" {
		sb.WriteString(`echo GET `)
		sb.WriteString(veg.Spec.Attack.Target)
		sb.WriteString(` | `)
	}
	sb.WriteString("vegeta attack")
	if veg.Spec.Attack.BodyConfigMap != "" {
		sb.WriteString(" -body ")
		sb.WriteString(bodyPath)
	}

	if veg.Spec.Attack.Chunked {
		sb.WriteString(" -chunked")
	}

	if veg.Spec.Attack.Connections > 0 {
		sb.WriteString(" -connections ")
		sb.WriteString(strconv.Itoa(veg.Spec.Attack.Connections))
	}

	if veg.Spec.Attack.Duration != "" {
		sb.WriteString(" -duration ")
		sb.WriteString(veg.Spec.Attack.Duration)
	}

	if veg.Spec.Attack.Format != "" {
		sb.WriteString(" -format ")
		sb.WriteString(veg.Spec.Attack.Format.String())
	}

	if veg.Spec.Attack.H2C {
		sb.WriteString(" -h2c")
	}

	if veg.Spec.Attack.HTTP2 {
		sb.WriteString(" -http2")
	}

	if veg.Spec.Attack.Insecure {
		sb.WriteString(" -insecure")
	}

	if veg.Spec.Attack.KeepAlive {
		sb.WriteString(" -keepalive")
	}

	if veg.Spec.Attack.Lazy {
		sb.WriteString(" -lazy")
	}

	if veg.Spec.Attack.MaxBody != 0 {
		sb.WriteString(" -max-body ")
		sb.WriteString(strconv.FormatUint(uint64(veg.Spec.Attack.MaxBody), 10))
	}

	if veg.Spec.Attack.MaxWorkers != 0 {
		sb.WriteString(" -max-workers ")
		sb.WriteString(strconv.FormatUint(uint64(veg.Spec.Attack.MaxWorkers), 10))
	}

	if veg.Spec.Attack.Name != "" {
		sb.WriteString(" -max-name ")
		sb.WriteString(veg.Spec.Attack.Name)
	}

	if veg.Spec.Attack.ProxyHeader != "" {
		sb.WriteString(" -proxy-header ")
		sb.WriteString(veg.Spec.Attack.ProxyHeader)
	}

	if veg.Spec.Attack.Rate != "" {
		sb.WriteString(" -rate ")
		sb.WriteString(veg.Spec.Attack.Rate)
	}

	if veg.Spec.Attack.Redirects != 0 {
		sb.WriteString(" -redirects ")
		sb.WriteString(strconv.Itoa(veg.Spec.Attack.Redirects))
	}

	if veg.Spec.Attack.RootCertsConfigMap == "" {
		sb.WriteString(" -root-certs /etc/pki/tls/certs/ca-bundle.crt,/var/run/secrets/kubernetes.io/serviceaccount/ca.crt,/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt")
	} else {
		sb.WriteString(" -root-certs /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem")
	}

	if veg.Spec.Attack.Timeout != "" {
		sb.WriteString(" -timeout ")
		sb.WriteString(veg.Spec.Attack.Timeout)
	}

	if veg.Spec.Attack.Workers > 0 {
		sb.WriteString(" -workers ")
		sb.WriteString(strconv.FormatUint(uint64(veg.Spec.Attack.Workers), 10))
	}

	// In case of results being sent to standard ouptut the report should be processed immediately. There is no way to process it afterwards.
	if veg.Spec.Report == nil || veg.Spec.Report.OutputType == "" || veg.Spec.Report.OutputType == vegetav1alpha1.StdoutOutput {
		sb.WriteString(" | ")
		sb.WriteString(getReportCmd(veg))
	}

	// TODO:
	// The report command accepts multiple result files. It'll read and sort them by timestamp before generating reports.
	// For supporting distributed attacks this means that it is best to process the reports separately.
	// This will be done by a pod, which gets launched after the attack pods have successfully completed
	return sb.String()
}

// getReportCmd generates the report command  based on the parameters configured in the vegeta resource
func getReportCmd(veg *vegetav1alpha1.Vegeta) string {
	var sb strings.Builder
	sb.WriteString("vegeta report")

	if veg.Spec.Report != nil && veg.Spec.Report.Buckets != "" {
		sb.WriteString(" -buckets ")
		sb.WriteString(veg.Spec.Report.Buckets)
	}

	if veg.Spec.Report != nil && veg.Spec.Report.Every != "" {
		sb.WriteString(" -every ")
		sb.WriteString(veg.Spec.Report.Every)
	}

	if veg.Spec.Report != nil && veg.Spec.Report.OutputType != "" && veg.Spec.Report.OutputType != vegetav1alpha1.StdoutOutput {
		// TODO: I am only generating reports in binary format. I may need to encode them in one of the available formats: (gob | json | csv)
		sb.WriteString(" -output ")
		sb.WriteString(getReportFileName(veg))
	}

	if veg.Spec.Report != nil && veg.Spec.Report.Type.String() != "" {
		sb.WriteString(" -type ")
		sb.WriteString(veg.Spec.Report.Type.String())
	}
	return sb.String()
}

// getReportFileName generates the name of the report file
func getReportFileName(veg *vegetav1alpha1.Vegeta) string {
	fname := []string{}
	fname = append(fname, veg.GetName(), "-$(hostname)-$(date +%s).gob")
	return veg.GetName() + "-$(hostname)-$(date +%s).gob"
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
	// - TargetsConfigMap targets.json or targets.http (depending on format)
	var ro int32 = 292

	// TODO: check outputType and adapt the code for storage of results accordingly.
	volumes := []corev1.Volume{}
	volumes = append(volumes,
		corev1.Volume{
			Name: "vegeta-results",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

	mounts := []corev1.VolumeMount{}
	mounts = append(mounts,
		corev1.VolumeMount{
			Name:      "vegeta-results",
			MountPath: resultsPath,
		})

	if veg.Spec.Attack.RootCertsConfigMap != "" {
		volumes = append(volumes,
			corev1.Volume{
				Name: "trusted-ca",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: veg.Spec.Attack.RootCertsConfigMap,
						},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{
								Key:  "ca-bundle.crt",
								Path: "tls-ca-bundle.pem",
							},
						},
						DefaultMode: &ro,
					},
				},
			},
		)

		mounts = append(mounts,
			corev1.VolumeMount{
				Name:      "trusted-ca",
				MountPath: "/etc/pki/ca-trust/extracted/pem/",
				ReadOnly:  true,
			},
		)
	}

	if veg.Spec.Attack.BodyConfigMap != "" {
		volumes = append(volumes,
			corev1.Volume{
				Name: "body",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: veg.Spec.Attack.BodyConfigMap,
						},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{
								Key:  "body.txt",
								Path: "body.txt",
							},
						},
						DefaultMode: &ro,
					},
				},
			},
		)

		mounts = append(mounts,
			corev1.VolumeMount{
				Name:      "body",
				MountPath: "/opt/config/",
				ReadOnly:  true,
			},
		)
	}

	return volumes, mounts
}
