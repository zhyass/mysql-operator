/*
Copyright 2021 zhyass.

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

package v1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of pods.
	// +optional
	// +kubebuilder:validation:Enum=0;2;3;5
	// +kubebuilder:default:=3
	Replicas *int32 `json:"replicas,omitempty"`

	// MysqlOpts is the options of MySQL container.
	// +optional
	// +kubebuilder:default:={rootPassword: "", user: "qc_usr", password: "Qing@123", database: "qingcloud", initTokuDB: true, resources: {limits: {cpu: "500m", memory: "1Gi"}, requests: {cpu: "100m", memory: "256Mi"}}}
	MysqlOpts MysqlOpts `json:"mysqlOpts,omitempty"`

	// XenonOpts is the options of xenon container.
	// +optional
	// +kubebuilder:default:={image: "xenondb/xenon:1.1.5-alpha", admitDefeatHearbeatCount: 5, electionTimeout: 10000, resources: {limits: {cpu: "100m", memory: "256Mi"}, requests: {cpu: "50m", memory: "128Mi"}}}
	XenonOpts XenonOpts `json:"xenonOpts,omitempty"`

	// +optional
	// +kubebuilder:default:={image: "prom/mysqld-exporter:v0.12.1", resources: {limits: {cpu: "100m", memory: "128Mi"}, requests: {cpu: "10m", memory: "32Mi"}}, enabled: false}
	MetricsOpts MetricsOpts `json:"metricsOpts,omitempty"`

	// Represents the MySQL version that will be run. The available version can be found here:
	// This field should be set even if the Image is set to let the operator know which mysql version is running.
	// Based on this version the operator can take decisions which features can be used.
	// +optional
	// +kubebuilder:default:="5.7"
	MysqlVersion string `json:"mysqlVersion,omitempty"`

	// Pod extra specification
	// +optional
	// +kubebuilder:default:={imagePullPolicy: "IfNotPresent", serviceAccountName: "mysql", resources: {limits: {cpu: "100m", memory: "128Mi"}, requests: {cpu: "10m", memory: "32Mi"}}, sidecarImage: "zhyass/sidecar:0.0.1"}
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// PVC extra specifiaction
	// +optional
	// +kubebuilder:default:={enabled: true, accessModes: {"ReadWriteOnce"}, size: "10Gi"}
	Persistence Persistence `json:"persistence,omitempty"`
}

// MysqlOpts defines the options of MySQL container.
type MysqlOpts struct {
	// Password for the root user.
	// +optional
	// +kubebuilder:default:=""
	RootPassword string `json:"rootPassword,omitempty"`

	// Username of new user to create.
	// +optional
	// +kubebuilder:default:="qc_usr"
	User string `json:"user,omitempty"`

	// Password for the new user.
	// +optional
	// +kubebuilder:default:="Qing@123"
	Password string `json:"password,omitempty"`

	// Name for new database to create.
	// +optional
	// +kubebuilder:default:="qingcloud"
	Database string `json:"database,omitempty"`

	// Install tokudb engine.
	// +optional
	// +kubebuilder:default:=true
	InitTokuDB bool `json:"initTokuDB,omitempty"`

	// A map[string]string that will be passed to my.cnf file.
	// +optional
	MysqlConf MysqlConf `json:"mysqlConf,omitempty"`

	// +optional
	// +kubebuilder:default:={limits: {cpu: "500m", memory: "1Gi"}, requests: {cpu: "100m", memory: "256Mi"}}
	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

// XenonOpts defines the options of xenon container.
type XenonOpts struct {
	// To specify the image that will be used for xenon container.
	// +optional
	// +kubebuilder:default:="zhyass/xenon:1.1.5"
	Image string `json:"image,omitempty"`

	// High available component admit defeat heartbeat count.
	// +optional
	// +kubebuilder:default:=5
	AdmitDefeatHearbeatCount *int32 `json:"admitDefeatHearbeatCount,omitempty"`

	// High available component election timeout. The unit is millisecond.
	// +optional
	// +kubebuilder:default:=10000
	ElectionTimeout *int32 `json:"electionTimeout,omitempty"`

	// +optional
	// +kubebuilder:default:={limits: {cpu: "100m", memory: "256Mi"}, requests: {cpu: "50m", memory: "128Mi"}}
	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

type MetricsOpts struct {
	// +optional
	// +kubebuilder:default:="prom/mysqld-exporter:v0.12.1"
	Image string `json:"image,omitempty"`

	// +optional
	// +kubebuilder:default:={limits: {cpu: "100m", memory: "128Mi"}, requests: {cpu: "10m", memory: "32Mi"}}
	Resources core.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`
}

// MysqlConf defines type for extra cluster configs. It's a simple map between
// string and string.
type MysqlConf map[string]intstr.IntOrString

// PodSpec defines type for configure cluster pod spec.
type PodSpec struct {
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default:="IfNotPresent"
	ImagePullPolicy core.PullPolicy `json:"imagePullPolicy,omitempty"`

	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Affinity          *core.Affinity    `json:"affinity,omitempty"`
	PriorityClassName string            `json:"priorityClassName,omitempty"`
	Tolerations       []core.Toleration `json:"tolerations,omitempty"`
	SchedulerName     string            `json:"schedulerName,omitempty"`

	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:default:="mysql"
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// +optional
	// +kubebuilder:default:={limits: {cpu: "100m", memory: "128Mi"}, requests: {cpu: "10m", memory: "32Mi"}}
	Resources core.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	// +kubebuilder:default:="zhyass/sidecar:0.0.1"
	SidecarImage string `json:"sidecarImage,omitempty"`
}

// Persistence is the desired spec for storing mysql data. Only one of its
// members may be specified.
type Persistence struct {
	// +optional
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	// +kubebuilder:default:={"ReadWriteOnce"}
	AccessModes []core.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// +optional
	// +kubebuilder:default:="10Gi"
	Size string `json:"size,omitempty"`
}

type ClusterConditionType string

const (
	ClusterReady ClusterConditionType = "Ready"
	ClusterInit  ClusterConditionType = "Initializing"
	ClusterError ClusterConditionType = "Error"
)

// ClusterCondition defines type for cluster conditions.
type ClusterCondition struct {
	// type of cluster condition, values in (\"Ready\")
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of (\"True\", \"False\", \"Unknown\")
	Status core.ConditionStatus `json:"status"`

	// LastTransitionTime
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason
	Reason string `json:"reason"`
	// Message
	Message string `json:"message"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ReadyNodes represents number of the nodes that are in ready state
	ReadyNodes int `json:"readyNodes,omitempty"`
	// Conditions contains the list of the cluster conditions fulfilled
	Conditions []ClusterCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.readyNodes
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status",description="The cluster status"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="The number of desired nodes"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=mysql
// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
