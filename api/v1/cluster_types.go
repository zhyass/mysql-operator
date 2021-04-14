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
	VolumeSpec VolumeSpec `json:"volumeSpec,omitempty"`
}

// MysqlOpts defines the options of MySQL container.
type MysqlOpts struct {
	// To specify the image that will be used for mysql server container.
	// If this is specified then the mysqlVersion is used as source for MySQL server version.
	// +optional
	Image string `json:"image,omitempty"`

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
}

// MysqlConf defines type for extra cluster configs. It's a simple map between
// string and string.
type MysqlConf map[string]intstr.IntOrString

// PodSpec defines type for configure cluster pod spec.
type PodSpec struct {
	ImagePullPolicy  core.PullPolicy             `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	Labels             map[string]string         `json:"labels,omitempty"`
	Annotations        map[string]string         `json:"annotations,omitempty"`
	Resources          core.ResourceRequirements `json:"resources,omitempty"`
	Affinity           *core.Affinity            `json:"affinity,omitempty"`
	XenonLifecycle     *core.Lifecycle           `json:"xenonLifecycle,omitempty"`
	NodeSelector       map[string]string         `json:"nodeSelector,omitempty"`
	PriorityClassName  string                    `json:"priorityClassName,omitempty"`
	Tolerations        []core.Toleration         `json:"tolerations,omitempty"`
	ServiceAccountName string                    `json:"serviceAccountName,omitempty"`

	// Volumes allows adding extra volumes to the statefulset
	// +optional
	Volumes []core.Volume `json:"volumes,omitempty"`

	// VolumesMounts allows mounting extra volumes to the mysql container
	// +optional
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`

	// InitContainers allows the user to specify extra init containers
	// +optional
	InitContainers []core.Container `json:"initContainers,omitempty"`

	// Containers allows for user to specify extra sidecar containers to run along with mysql
	// +optional
	Containers []core.Container `json:"containers,omitempty"`
}

// VolumeSpec is the desired spec for storing mysql data. Only one of its
// members may be specified.
type VolumeSpec struct {
	// EmptyDir to use as data volume for mysql. EmptyDir represents a temporary
	// directory that shares a pod's lifetime.
	// +optional
	EmptyDir *core.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// HostPath to use as data volume for mysql. HostPath represents a
	// pre-existing file or directory on the host machine that is directly
	// exposed to the container.
	// +optional
	HostPath *core.HostPathVolumeSource `json:"hostPath,omitempty"`

	// PersistentVolumeClaim to specify PVC spec for the volume for mysql data.
	// It has the highest level of precedence, followed by HostPath and
	// EmptyDir. And represents the PVC specification.
	// +optional
	PersistentVolumeClaim *core.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
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
