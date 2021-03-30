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
	containerName   = "vegeta"
	configPath      = "/opt/config/"
	credentialsPath = "/opt/config/credentials/"
	resultsPath     = "/results/"
)

// aPod4Attack generates the definition of the attack pod
func (r *VegetaReconciler) aPod4Attack(v *vegetav1alpha1.Vegeta) *corev1.Pod {
	immediate := int64(0)
	volumes, mounts := getAPVolumesAndMounts(v)
	var image string
	if vImg := strings.TrimSpace(v.Spec.Image); vImg != "" {
		image = vImg
	} else {
		image = r.Image
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: v.Name + "-",
			Namespace:    v.Namespace,
			Labels: r.Labels.Merge(map[string]string{
				"app.kubernetes.io/name":       "vegeta",
				"app.kubernetes.io/instance":   v.Name,
				"app.kubernetes.io/managed-by": "vegeta-operator",
				"vegeta.testing.io/type":       "attack"}),
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
				Env:        getAttackEnv(v),
				EnvFrom:    getEnvFrom(v),
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

// aPod4Report generates the definition of the report pod
func (r *VegetaReconciler) aPod4Report(v *vegetav1alpha1.Vegeta) *corev1.Pod {
	immediate := int64(0)
	volumes, mounts := getRPVolumesAndMounts(v)
	var image string
	if vImg := strings.TrimSpace(v.Spec.Image); vImg != "" {
		image = vImg
	} else {
		image = r.Image
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: v.Name + "-report-",
			Namespace:    v.Namespace,
			Labels: r.Labels.Merge(map[string]string{
				"app.kubernetes.io/name":       "vegeta",
				"app.kubernetes.io/instance":   v.Name,
				"app.kubernetes.io/managed-by": "vegeta-operator",
				"vegeta.testing.io/type":       "report"}),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:           image,
				ImagePullPolicy: "Always",
				Name:            containerName,
				Command:         []string{"/bin/sh"},
				Args:            []string{"-c", getReportCmd(v)},
				Resources:       v.Spec.Resources,
				VolumeMounts:    mounts,
				// TODO: I am not sure this needs to be made configurable. What is defined in the image should be just fine.
				WorkingDir: resultsPath,
				Env:        getReportEnv(v),
				EnvFrom:    getEnvFrom(v),
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

	/*
	   Mounts:
	   // - BodyConfigMap body.txt Specifies a config map containing the body of every request unless overridden per attack target.
	   // - CertSecret client.crt Specifies the secret containing the TLS client PEM encoded certificate file.
	   // - HeadersConfigMap headers.txt Specifies a config map containing request headers to be used in all targets defined
	   // - KeySecret client.key Specifies the secret containing the PEM encoded TLS client certificate private key
	   	// - TargetsConfigMap targets.json or targets.http (depending on format)
	*/

	var sb strings.Builder

	if veg.Spec.Attack.TargetsConfigMap != "" {
		sb.WriteString("vegeta attack -targets ")
		sb.WriteString(configPath)
		sb.WriteString("targets")
		if veg.Spec.Attack.Format == vegetav1alpha1.JSONFormat {
			sb.WriteString(".json ")
		} else {
			sb.WriteString(".http ")
		}
	} else if veg.Spec.Attack.Target != "" {
		sb.WriteString("echo ")
		sb.WriteString(veg.Spec.Attack.Target)
		sb.WriteString(" | ")
		sb.WriteString("vegeta attack")
	}

	if veg.Spec.Attack.BodyConfigMap != "" {
		sb.WriteString(" -body ")
		sb.WriteString(configPath)
		sb.WriteString("body.txt")
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

	for _, h := range veg.Spec.Attack.Headers {
		sb.WriteString(" -header \"")
		sb.WriteString(h)
		sb.WriteString("\"")
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

	if veg.Spec.Attack.KeySecret != "" {
		sb.WriteString(" -key ")
		sb.WriteString(credentialsPath)
		sb.WriteString("client.key")
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

	sb.WriteString(" -root-certs /var/run/secrets/kubernetes.io/serviceaccount/ca.crt,/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt,/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem")

	if veg.Spec.Attack.Timeout != "" {
		sb.WriteString(" -timeout ")
		sb.WriteString(veg.Spec.Attack.Timeout)
	}

	if veg.Spec.Attack.Workers > 0 {
		sb.WriteString(" -workers ")
		sb.WriteString(strconv.FormatUint(uint64(veg.Spec.Attack.Workers), 10))
	}

	// In case of results being sent to standard ouptut the report should be processed immediately. There is no way to process it afterwards. Otherwise the output gets stored for later processing.
	if veg.Spec.Report == nil {
		sb.WriteString(" | ")
		sb.WriteString(getReportCmd(veg))
	} else {
		switch veg.Spec.Report.OutputType {
		case vegetav1alpha1.PvcOutput:
			sb.WriteString(" -output ")
			sb.WriteString(resultsPath)
			sb.WriteString(getResultFileName(veg))
			sb.WriteString("_res.gob")
		case vegetav1alpha1.ObcOutput:
			sb.WriteString(" -output ")
			sb.WriteString(resultsPath)
			sb.WriteString(getResultFileName(veg))
			sb.WriteString("_res.gob")
			// Additional step to upload the result file to the S3 bucket
			sb.WriteString("; s3 -command upload")
		default:
			sb.WriteString(" | ")
			sb.WriteString(getReportCmd(veg))
		}
	}

	return sb.String()
}

// getReportCmd generates the report command  based on the parameters configured in the vegeta resource
func getReportCmd(veg *vegetav1alpha1.Vegeta) string {
	var sb strings.Builder
	upload := ""
	if veg.Spec.Report != nil && veg.Spec.Report.OutputType == vegetav1alpha1.ObcOutput {
		// Get the files downloaded from the S3 bucket
		sb.WriteString("s3 -command download; ")
		upload = "; s3 -command upload"
	}

	sb.WriteString("vegeta report ")

	if veg.Spec.Report != nil && veg.Spec.Report.Buckets != "" {
		sb.WriteString(" -buckets ")
		sb.WriteString(veg.Spec.Report.Buckets)
	}

	if veg.Spec.Report != nil && veg.Spec.Report.Every != "" {
		sb.WriteString(" -every ")
		sb.WriteString(veg.Spec.Report.Every)
	}

	if veg.Spec.Report != nil && veg.Spec.Report.OutputType != vegetav1alpha1.StdoutOutput {
		// TODO: I am only generating reports in binary format. I may need to encode them in one of the available formats: (gob | json | csv)
		sb.WriteString(" -output ")
		fileName := getResultFileName(veg) + "_rep.gob "
		sb.WriteString(fileName)
	}

	if veg.Spec.Report != nil && veg.Spec.Report.Type.String() != "" {
		sb.WriteString(" -type ")
		sb.WriteString(veg.Spec.Report.Type.String())
	}

	if veg.Spec.Report != nil && veg.Spec.Report.OutputType != vegetav1alpha1.StdoutOutput {
		sb.WriteString(resultsPath)
		sb.WriteString(veg.ObjectMeta.GetCreationTimestamp().Format("20060102150405"))
		sb.WriteString("-")
		sb.WriteString(veg.Name)
		sb.WriteString("*_res.*")
	}
	sb.WriteString(upload)
	return sb.String()
}

// getResultFileName generates the name of the result file (used for result and report)
func getResultFileName(veg *vegetav1alpha1.Vegeta) string {
	return veg.ObjectMeta.GetCreationTimestamp().Format("20060102150405") + "-${HOSTNAME}"
}

func getResultBaseName(veg *vegetav1alpha1.Vegeta) string {
	return veg.ObjectMeta.GetCreationTimestamp().Format("20060102150405") + "-" + veg.Name
}

// getAPVolumesAndMounts generates the list of volumes and mounts for the attack pod
func getAPVolumesAndMounts(veg *vegetav1alpha1.Vegeta) ([]corev1.Volume, []corev1.VolumeMount) {
	// reports PV are mounted RW under /reports/
	// configs CM are mounted RO under /opt/config/ (except RootCertsConfigMap)
	// secrets are mounted RO under /opt/config/credentials/
	// Fields (all optionals):
	// - BodyConfigMap body.txt Specifies a config map containing the body of every request unless overridden per attack target.
	// - KeySecret client.key Specifies the secret containing the PEM encoded TLS client certificate private key
	// - TargetsConfigMap targets.json or targets.http (depending on format)
	var ro int32 = 292
	volumes := []corev1.Volume{}
	mounts := []corev1.VolumeMount{}

	if veg.Spec.Report == nil || veg.Spec.Report.OutputType != vegetav1alpha1.PvcOutput {
		volumes = append(volumes,
			corev1.Volume{
				Name: "vegeta-results",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
	} else {
		volumes = append(volumes,
			corev1.Volume{
				Name: "vegeta-results",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: veg.Spec.Report.OutputClaim,
					},
				},
			})
	}

	mounts = append(mounts,
		corev1.VolumeMount{
			Name:      "vegeta-results",
			MountPath: resultsPath,
		})

	if veg.Spec.Attack.RootCertsConfigMap != "" {
		caKey := veg.Spec.Attack.RootCertsFile
		if caKey == "" {
			caKey = "ca-bundle.crt"
		}
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
								Key:  caKey,
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
				MountPath: configPath,
				ReadOnly:  true,
			},
		)
	}

	if veg.Spec.Attack.KeySecret != "" {
		volumes = append(volumes,
			corev1.Volume{
				Name: "key",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: veg.Spec.Attack.KeySecret,
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{
								Key:  "client.key",
								Path: "client.key",
							},
						},
						DefaultMode: &ro,
					},
				},
			},
		)

		mounts = append(mounts,
			corev1.VolumeMount{
				Name:      "key",
				MountPath: credentialsPath,
				ReadOnly:  true,
			},
		)
	}

	if veg.Spec.Attack.TargetsConfigMap != "" {
		var file string
		if veg.Spec.Attack.Format == vegetav1alpha1.JSONFormat {
			file = "targets.json"
		} else {
			file = "targets.http"
		}
		volumes = append(volumes,
			corev1.Volume{
				Name: "targets",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: veg.Spec.Attack.TargetsConfigMap,
						},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{
								Key:  file,
								Path: file,
							},
						},
						DefaultMode: &ro,
					},
				},
			},
		)

		mounts = append(mounts,
			corev1.VolumeMount{
				Name:      "targets",
				MountPath: configPath,
				ReadOnly:  true,
			},
		)
	}

	return volumes, mounts
}

// getRPVolumesAndMounts generates the list of volumes and mounts for the report pod
func getRPVolumesAndMounts(veg *vegetav1alpha1.Vegeta) ([]corev1.Volume, []corev1.VolumeMount) {
	var ro int32 = 292
	volumes := []corev1.Volume{}
	mounts := []corev1.VolumeMount{}

	if veg.Spec.Report == nil || veg.Spec.Report.OutputType != vegetav1alpha1.PvcOutput {
		volumes = append(volumes,
			corev1.Volume{
				Name: "vegeta-results",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
	} else {
		volumes = append(volumes,
			corev1.Volume{
				Name: "vegeta-results",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: veg.Spec.Report.OutputClaim,
					},
				},
			})
	}

	mounts = append(mounts,
		corev1.VolumeMount{
			Name:      "vegeta-results",
			MountPath: resultsPath,
		})

	if veg.Spec.Attack.RootCertsConfigMap != "" {
		caKey := veg.Spec.Attack.RootCertsFile
		if caKey == "" {
			caKey = "ca-bundle.crt"
		}
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
								Key:  caKey,
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

	return volumes, mounts
}

func getAttackEnv(veg *vegetav1alpha1.Vegeta) []corev1.EnvVar {
	env := []corev1.EnvVar{}
	if veg.Spec.Report != nil && veg.Spec.Report.OutputType == vegetav1alpha1.ObcOutput {
		env = append(env,
			corev1.EnvVar{
				Name:  "S3_UPLOAD_FILE",
				Value: resultsPath + getResultFileName(veg) + "_res.gob",
			})
	}
	if veg.Spec.Attack.RootCertsConfigMap != "" {

		env = append(env,
			corev1.EnvVar{
				Name:  "SSL_CERT_FILE",
				Value: "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem",
			})
	}
	return env
}

func getReportEnv(veg *vegetav1alpha1.Vegeta) []corev1.EnvVar {
	env := []corev1.EnvVar{}
	if veg.Spec.Report != nil && veg.Spec.Report.OutputType == vegetav1alpha1.ObcOutput {
		env = append(env,
			corev1.EnvVar{
				Name:  "S3_OBJECT_PREFIX",
				Value: getResultBaseName(veg),
			})
		env = append(env,
			corev1.EnvVar{
				Name:  "S3_UPLOAD_FILE",
				Value: resultsPath + getResultFileName(veg) + "_rep.gob",
			})
		if veg.Spec.Attack.RootCertsConfigMap != "" {
			env = append(env,
				corev1.EnvVar{
					Name:  "SSL_CERT_FILE",
					Value: "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem",
				})
		}
	}
	return env
}

func getEnvFrom(veg *vegetav1alpha1.Vegeta) []corev1.EnvFromSource {
	envFrom := []corev1.EnvFromSource{}
	if veg.Spec.Report != nil && veg.Spec.Report.OutputType == vegetav1alpha1.ObcOutput {
		envFrom = []corev1.EnvFromSource{
			{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: veg.Spec.Report.OutputClaim,
					},
				},
			},
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: veg.Spec.Report.OutputClaim,
					},
				},
			},
		}
	}

	return envFrom
}
