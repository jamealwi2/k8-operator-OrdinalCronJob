/*
Copyright 2022.

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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OrdinalCronJobSpec defines the desired state of OrdinalCronJob
type OrdinalCronJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Edit dbscan_types.go to remove/update
	CronSchedule            string                       `json:"cronSchedule"`
	JobsRequired            string                       `json:"jobsRequired"`
	Image                   string                       `json:"image"`
	StartingDeadlineSeconds *int64                       `json:"startingDeadlineSeconds,omitempty"`
	Suspend                 *bool                        `json:"suspend,omitempty"`
	Command                 []string                     `json:"command,omitempty"`
	Args                    []string                     `json:"args,omitempty"`
	Annotations             map[string]string            `json:"annotations,omitempty"`
	Env                     []apiv1.EnvVar               `json:"env,omitempty"`
	Resources               apiv1.ResourceRequirements   `json:"resources,omitempty"`
	VolumeMounts            []apiv1.VolumeMount          `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath"`
	Volumes                 []apiv1.Volume               `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
	ImagePullSecrets        []apiv1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// OrdinalCronJobStatus defines the observed state of OrdinalCronJob
type OrdinalCronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OrdinalCronJob is the Schema for the ordinalcronjobs API
type OrdinalCronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrdinalCronJobSpec   `json:"spec,omitempty"`
	Status OrdinalCronJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OrdinalCronJobList contains a list of OrdinalCronJob
type OrdinalCronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrdinalCronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrdinalCronJob{}, &OrdinalCronJobList{})
}
