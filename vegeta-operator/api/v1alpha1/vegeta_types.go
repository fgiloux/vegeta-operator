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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// NOTE: Run "make generate" to regenerate code and "make manifests" to regenerate the CRD manifests after modifying this file

// AttackSpec defines the desired attacks.
type AttackSpec struct {
	// Specifies a config map containing the body of every request unless overridden per attack target.
	// The config  map should contain a single file named body.txt
	//
	// +optional
	BodyConfigMap string `json:"bodyConfigMap,omitempty"`

	// Specifies the secret containing the TLS client PEM encoded certificate file.
	// The secret should contain a single file named client.crt
	//
	// +optional
	CertSecret string `json:"certSecret,omitempty"`

	// Specifies whether to send request bodies with the chunked transfer encoding.
	//
	// +optional
	Chunked bool `json:"chunked,omitempty"`

	// Specifies the maximum number of idle open connections per target host (defaults to 10000).
	//
	// +optional
	Connections int `json:"connections,omitempty"`

	// Specifies the amount of time to issue request to the targets. The internal concurrency structure's setup has this value as a variable. The actual run time of the test can be longer than specified due to the responses delay. Use 0 for an infinite attack.
	//
	// +kubebuilder:validation:Format=duration
	// +optional
	Duration string `json:"duration,omitempty"`

	// Specifies the format of the target provided in the targets file, see below. Valid values are: json and http.
	// +optional
	Format TargetFormatEnum `json:"format,omitempty"`

	// Specifies that HTTP2 requests are to be sent over TCP without TLS encryption.
	//
	// +optional
	H2C bool `json:"h2c,omitempty"`

	// Specifies a config map containing request headers to be used in all targets defined.
	// The config map should contain a single file named headers.txt. You can specify as many as needed by writing a new header on a new line.
	//
	// +optional
	HeadersConfigMap string `json:"headersConfigMap,omitempty"`

	// Specifies whether to enable HTTP/2 requests to servers which support it.
	//
	// +optional
	HTTP2 bool `json:"http2,omitempty"`

	// Specifies whether to ignore invalid server TLS certificates.
	//
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// Specifies whether to reuse TCP connections between HTTP requests (defaults to true).
	//
	// +optional
	KeepAlive bool `json:"keepAlive,omitempty"`

	// Specifies the secret containing the PEM encoded TLS client certificate private key file to be used with HTTPS requests. The secret should contain a single file named client.key.
	//
	// +optional
	KeySecret string `json:"keySecret,omitempty"`

	// TODO:
	// Specifies the local IP address to be used (defaults to 0.0.0.0).
	// This may be configurable with multus but is left for now.
	// Validation: https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions/v1beta1/types_jsonschema.go
	// kubebuilder:validation:Format=ipv4
	// optional
	//LAddr string `json:"laddr,omitempty"`

	// Specifies whether to read the input targets lazily instead of eagerly.
	//
	// +optional
	Lazy bool `json:"lazy,omitempty"`

	// Specifies the maximum number of bytes to capture from the body of each response. Remaining unread bytes will be fully read but discarded. [-1 = no limit] (defaults to -1).
	//
	// +optional
	MaxBody uint `json:"maxBody,omitempty"`

	// MaxWorkers specifies the Maximum number of workers, i.e. goroutines (defaults to 18446744073709551615).
	//
	// +optional
	MaxWorkers uint `json:"maxWorkers,omitempty"`

	// Specifies the name of the attack to be recorded in responses.
	//
	// +optional
	Name string `json:"name,omitempty"`

	// TODO: I am not sure it is a good idea to have it configurable (at least in a first iteration). For now the output is directly piped into the result processing command.
	// Specifies the output file to which the binary results will be written to. Made to be piped to the report command input. Defaults to stdout.
	//
	// optional
	// Output string `json:"output,omitempty"`

	// Specifies the Proxy CONNECT header.
	//
	// +optional
	ProxyHeader string `json:"proxyHeader,omitempty"`

	// Specifies the request rate per time unit to issue against the targets.
	// 0 or infinity means vegeta will send requests as fast as possible. Use together with MaxWorkers to model a fixed set of concurrent users sending requests serially (i.e. waiting for a response before sending the next request).Defaults to 50/1s.
	//
	// +optional
	Rate string `json:"rate,omitempty"`

	// Specifies the max number of redirects followed on each request. The default is 10. When the value is -1, redirects are not followed but the response is marked as successful.
	//
	// +optional
	Redirects int `json:"redirects,omitempty"`

	// Specifies custom DNS resolver addresses to use for name resolution instead of the ones configured by the operating system.
	// It is of no interest as pods allow more ellaborate DNS configuration:
	// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config
	// TODO: add the matching fields to the VegetaSpec, possibly in a subresource containing pod configuration.
	// Resolvers string `json:"resolvers,omitempty"`
	// TODO: Supporting additional pod configuration could be considered
	// - priorityClassName
	// - schedulerName
	// - serviceAccountName
	// - tolerations

	// Specifies a config map containing the trusted TLS root CAs certificate files If unspecified, the default system CAs certificates will be used.
	// The key for the file should be named: ca-bundle.crt
	// The file should be named: tls-ca-bundle.pem
	// With OpenShift this config map can get automatically populated by configuring cluster-wide trusted CA certificates and setting the following label to the empty config map: config.openshift.io/inject-trusted-cabundle=true, whose name is set into this field.
	//
	//
	// +optional
	RootCertsConfigMap string `json:"rootCertsConfigMap,omitempty"`

	// Target refers to the target endpoint for the load testing.
	// For multiple targets use TargetsConfigMap.
	//
	// +optional
	Target string `json:"target"`

	// Specifies a config map containing the file from which to read targets. The config map should contain a single file named targets with the format as extension, i.e. targets.json. See the format section to learn about the different target formats.
	//
	// +optional
	TargetsConfigMap string `json:"targetsConfigMap,omitempty"`

	// Specifies the timeout for each request. The default is 0 which disables timeouts.
	//
	// +kubebuilder:validation:Format=duration
	// +optional
	Timeout string `json:"timeout,omitempty"`

	// Unix socket is not covered at this point as its usage would make a few assumptions that are not given in K8:
	// - vegeta pod would run on the same host as the target
	// - the unix socket would be mounted in both pods

	// Specifies the initial number of workers, i.e. goroutines, used in the attack. It defaults to 10. The actual number of workers will increase if necessary in order to sustain the requested rate, unless it'd go beyond MaxWorkers.
	//
	// +optional
	Workers uint `json:"workers,omitempty"`
}

// ReportSpec defines the desired report
type ReportSpec struct {

	// TODO: To check whether I want to provide a canonical way of storing reports:
	// - Specifying the name of a PVC where the files get recorded
	// - Name with generated part by default for keeping history
	// - Option to activate Prometheus scrapping

	// Buckets defines the histogram buckets, e.g.: "[0,1ms,10ms]".
	//
	// +optional
	Buckets string `json:"buckets,omitempty"`

	// The report is written to Output at Every given interval (e.g 100ms). The default of 0 means the report will only be written after all results have been processed.
	//
	// +kubebuilder:validation:Format=duration
	// +optional
	Every string `json:"every,omitempty"`

	// Specifies the output location. The value should match a persistent volume claim or an object bucket claim name.
	// In case of PVC the name of the result file is based on the attack, pod name and a timestamp. For now volumes are to be RWM in case of a distributed attack as they get mounted by each pod.
	//
	// +optional
	OutputClaim string `json:"outputClaim,omitempty"`

	// Specifies the type of storage to use for the output. Valid values are stdout, pvc, obc. Stdout is the default.
	//
	// +optional
	OutputType OutputTypeEnum `json:"outputType,omitempty"`

	// Type defines the report type to generate. Valid values are text, json, hist, hdrplot. It defaults to "text".
	//
	// +optional
	Type ReportTypeEnum `json:"type,omitempty"`
}

// VegetaSpec defines the desired state of Vegeta
type VegetaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies the attack parameters.
	//
	// +required
	Attack *AttackSpec `json:"attack"`

	// Specifies the number of pods running the attack. The attack as specified above will be run by each pod. This brings an additional level of parallelism and scalability to what workers provide.
	//
	// +optional
	Replicas uint `json:"replicas,omitempty"`

	// Specifies the report parameters.
	//
	// +optional
	Report *ReportSpec `json:"report,omitempty"`

	// Specifies the resource requests and limits of the vegeta container.
	//
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Image allows to select a different container image for the Vegeta attack than the one configured at the operator level
	// +optional
	Image string `json:"image,omitempty"`
}

// VegetaStatus defines the observed state of Vegeta
type VegetaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO: Ideally this would reflect the result of the test execution.

	// Active contains the names of currently running pods.
	// +optional
	Active []string `json:"active,omitempty"`

	// Failed contains the names of pods that failed.
	// +optional
	Failed []string `json:"failed,omitempty"`

	// Succeeded contains the names of pods that sucessfully completed.
	// +optional
	Succeeded []string `json:"succeeded,omitempty"`

	// Phase of the processing of the Vegeta request. Possible values are: pending (no pod started), running (not all pods have terminated yet and no pod has failed), failed (one of the pod has failed), succeeded (all pods have successfully terminated but report has not been generated yet), completed (all pods have successfully terminated and report has been generated)
	Phase PhaseEnum `json:"type,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Vegeta is the Schema for the vegeta API
type Vegeta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VegetaSpec   `json:"spec,omitempty"`
	Status VegetaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VegetaList contains a list of Vegeta
type VegetaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vegeta `json:"items"`
}

// OutputTypeEnum is an enumeration of possible output types
// +kubebuilder:validation:Enum=stdout;pvc;obc
type OutputTypeEnum string

const (
	// StdoutOutput requests that results are sent to stdout
	StdoutOutput OutputTypeEnum = "stdout"
	// PvcOutput requests that results are stored in a persistent volume
	PvcOutput OutputTypeEnum = "pvc"
	// ObcOutput requests that results are stored in an object buket
	ObcOutput OutputTypeEnum = "obc"
)

func (e OutputTypeEnum) String() string {
	switch e {
	case StdoutOutput:
		return "stdout"
	case PvcOutput:
		return "pvc"
	case ObcOutput:
		return "obc"
	default:
		return ""
	}
}

// PhaseEnum is an enumaration of possible phases for  the vegeta resource
type PhaseEnum string

const (
	// PendingPhase means that no pod has started
	PendingPhase PhaseEnum = "pending"
	// RunningPhase means that not all pods have terminated yet and no pod has failed
	RunningPhase PhaseEnum = "running"
	// SucceededPhase means that all pods have successfully terminated but report has not been generated yet
	SucceededPhase PhaseEnum = "succeeded"
	// FailedPhase means that one of the pods has failed
	FailedPhase PhaseEnum = "failed"
	// CompletedPhase means that all pods have successfully terminated and report has been generated
	CompletedPhase PhaseEnum = "completed"
)

func (e PhaseEnum) String() string {
	switch e {
	case PendingPhase:
		return "pending"
	case RunningPhase:
		return "running"
	case SucceededPhase:
		return "succeeded"
	case FailedPhase:
		return "failed"
	case CompletedPhase:
		return "completed"
	default:
		return ""
	}
}

// ReportTypeEnum is an enumeration of possible types of reports
// +kubebuilder:validation:Enum=text;json;hist;hdrplot
type ReportTypeEnum string

const (
	// TextReport specifies that reports should be generated in text format
	TextReport ReportTypeEnum = "text"
	// JSONReport specifies that reports should be generated in json format
	JSONReport ReportTypeEnum = "json"
	// HistReport specifies that reports should be generated as text based histogram for the given buckets
	HistReport ReportTypeEnum = "hist"
	// HDRPlotReport specifies that the reports should be written in a format plottable by https://hdrhistogram.github.io/HdrHistogram/plotFiles.html.
	HDRPlotReport ReportTypeEnum = "hdrplot"
)

func (e ReportTypeEnum) String() string {
	switch e {
	case TextReport:
		return "text"
	case JSONReport:
		return "json"
	case HistReport:
		return "hist"
	case HDRPlotReport:
		return "hdrplot"
	default:
		return ""
	}
}

// TargetFormatEnum is an enumaration of possible formats for the target file
// +kubebuilder:validation:Enum=json;http
type TargetFormatEnum string

const (
	// JSONFormat specifies that the target file is in json format
	JSONFormat TargetFormatEnum = "json"
	// HTTPFormat specifies that the target file is in http format
	HTTPFormat TargetFormatEnum = "http"
)

func (e TargetFormatEnum) String() string {
	switch e {
	case JSONFormat:
		return "json"
	case HTTPFormat:
		return "http"
	default:
		return ""
	}
}

func init() {
	SchemeBuilder.Register(&Vegeta{}, &VegetaList{})
}
