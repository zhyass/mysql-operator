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
	"math"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

// nolint: megacheck, deadcode, varcheck
const (
	_        = iota // ignore first value by assigning to blank identifier
	kb int64 = 1 << (10 * iota)
	mb
	gb
)

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-mysql-radondb-io-v1-cluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=mysql.radondb.io,resources=clusters,verbs=create;update,versions=v1,name=mcluster.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Cluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cluster) Default() {
	clusterlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if len(r.Spec.MysqlOpts.MysqlConf) == 0 {
		r.Spec.MysqlOpts.MysqlConf = make(MysqlConf)
	}

	var defaultSize, maxSize, innodbBufferPoolSize int64
	innodbBufferPoolSize = 128 * mb
	conf, ok := r.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"]
	mem := r.Spec.MysqlOpts.Resources.Requests.Memory().Value()
	cpu := r.Spec.PodSpec.Resources.Limits.Cpu().MilliValue()
	if mem <= 1*gb {
		defaultSize = int64(0.45 * float64(mem))
		maxSize = int64(0.6 * float64(mem))
	} else {
		defaultSize = int64(0.6 * float64(mem))
		maxSize = int64(0.8 * float64(mem))
	}

	if !ok {
		innodbBufferPoolSize = max(defaultSize, innodbBufferPoolSize)
	} else {
		innodbBufferPoolSize = min(max(int64(conf.IntVal), innodbBufferPoolSize), maxSize)
	}

	instances := math.Max(math.Min(math.Ceil(float64(cpu)/float64(1000)), math.Floor(float64(innodbBufferPoolSize)/float64(gb))), 1)
	r.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"] = intstr.FromInt(int(innodbBufferPoolSize))
	r.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"] = intstr.FromInt(int(instances))
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-mysql-radondb-io-v1-cluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=mysql.radondb.io,resources=clusters,verbs=create;update,versions=v1,name=vcluster.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateUpdate(old runtime.Object) error {
	clusterlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
