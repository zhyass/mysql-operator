/*
Copyright 2021.

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
	// Defaults to 3
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// MysqlOpts is the options of MySQL container.
	// +optional
	MysqlOpts MysqlOpts `json:"mysqlOpts,omitempty"`

	// XenonOpts is the options of xenon container.
	// +optional
	XenonOpts XenonOpts `json:"xenonOpts,omitempty"`

	// Represents the MySQL version that will be run. The available version can be found here:
	// This field should be set even if the Image is set to let the operator know which mysql version is running.
	// Based on this version the operator can take decisions which features can be used.
	// Defaults to 5.7
	// +optional
	MysqlVersion string `json:"mysqlVersion,omitempty"`

	// Pod extra specification
	// +optional
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// PVC extra specifiaction
	// +optional
	Persistence Persistence `json:"persistence,omitempty"`
}

// MysqlOpts defines the options of MySQL container.
type MysqlOpts struct {
	// Password for the root user.
	// +optional
	RootPassword string `json:"rootPassword,omitempty"`

	// Username of new user to create.
	// Defaults to qc_usr
	// +optional
	User string `json:"user,omitempty"`

	// Password for the new user.
	// Defaults to Qing@123
	// +optional
	Password string `json:"password,omitempty"`

	// Name for new database to create.
	// Defaults to qingcloud
	// +optional
	Database string `json:"database,omitempty"`

	// Install tokudb engine.
	// +optional
	InitTokudb bool `json:"initTokudb,omitempty"`

	// A map[string]string that will be passed to my.cnf file.
	// +optional
	MysqlConf MysqlConf `json:"mysqlConf,omitempty"`

	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

// XenonOpts defines the options of xenon container.
type XenonOpts struct {
	// To specify the image that will be used for xenon container.
	// +optional
	Image string `json:"image,omitempty"`

	// High available component admit defeat heartbeat count.
	// +optional
	AdmitDefeatHearbeatCount *int32 `json:"admitDefeatHearbeatCount,omitempty"`

	// High available component election timeout. The unit is millisecond.
	// +optional
	ElectionTimeout *int32 `json:"electionTimeout,omitempty"`

	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

// MysqlConf defines type for extra cluster configs. It's a simple map between
// string and string.
type MysqlConf map[string]intstr.IntOrString

// PodSpec defines type for configure cluster pod spec.
type PodSpec struct {
	ImagePullPolicy core.PullPolicy `json:"imagePullPolicy,omitempty"`

	Labels             map[string]string         `json:"labels,omitempty"`
	Annotations        map[string]string         `json:"annotations,omitempty"`
	Affinity           *core.Affinity            `json:"affinity,omitempty"`
	PriorityClassName  string                    `json:"priorityClassName,omitempty"`
	Tolerations        []core.Toleration         `json:"tolerations,omitempty"`
	SchedulerName      string                    `json:"schedulerName,omitempty"`
	ServiceAccountName string                    `json:"serviceAccountName,omitempty"`
	Resources          core.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	BusyboxImage string `json:"busyboxImage,omitempty"`

	// +optional
	MetricsImage string `json:"metricsImage,omitempty"`

	// Volumes allows adding extra volumes to the statefulset
	// +optional
	Volumes []core.Volume `json:"volumes,omitempty"`

	// VolumesMounts allows mounting extra volumes to the mysql container
	// +optional
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
}

// Persistence is the desired spec for storing mysql data. Only one of its
// members may be specified.
type Persistence struct {
	Enabled bool `json:"enabled,omitempty"`

	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	AccessModes []core.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	Size string `json:"size,omitempty"`
}

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

type ClusterConditionType string

const (
	ClusterReady ClusterConditionType = "Ready"
	ClusterInit  ClusterConditionType = "Initializing"
	ClusterError ClusterConditionType = "Error"
)

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
