/*
Copyright The Volcano Authors.

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

// AutoscalingPolicyBindingSpec defines the desired state of AutoscalingPolicyBinding.
// +kubebuilder:validation:XValidation:rule="has(self.optimizerConfiguration) != has(self.scalingConfiguration)",message="Either optimizerConfiguration or scalingConfiguration must be set, but not both."
type AutoscalingPolicyBindingSpec struct {
	// PolicyRef references the AutoscalingPolicy that defines the scaling rules and metrics.
	PolicyRef corev1.LocalObjectReference `json:"policyRef"`

	// OptimizerConfiguration enables multi-target optimization that dynamically allocates
	// replicas across heterogeneous ModelServing deployments based on overall compute requirements.
	// This is ideal for mixed hardware environments (e.g., H100/A100 clusters) where you want to
	// optimize resource utilization by adjusting deployment ratios between different hardware types
	// using mathematical optimization methods (e.g. integer programming).
	OptimizerConfiguration *OptimizerConfiguration `json:"optimizerConfiguration,omitempty"`

	// ScalingConfiguration defines traditional autoscaling behavior that adjusts replica counts
	// based on monitoring metrics and target values for a single ModelServing deployment.
	ScalingConfiguration *ScalingConfiguration `json:"scalingConfiguration,omitempty"`
}

// AutoscalingTargetType defines the type of target for autoscaling operations.
type AutoscalingTargetType string

// MetricEndpoint defines the endpoint configuration for scraping metrics from pods.
type MetricEndpoint struct {
	// URI is the path where metrics are exposed (e.g., "/metrics").
	// +optional
	// +kubebuilder:default="/metrics"
	Uri string `json:"uri,omitempty"`
	// Port is the network port where metrics are exposed by the pods.
	// +optional
	// +kubebuilder:default=8100
	Port int32 `json:"port,omitempty"`
}

// ScalingConfiguration defines the scaling parameters for a single target deployment.
type ScalingConfiguration struct {
	// Target specifies the ModelServing deployment to monitor and scale.
	Target Target `json:"target,omitempty"`
	// MinReplicas is the minimum number of replicas to maintain.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000000
	MinReplicas int32 `json:"minReplicas"`
	// MaxReplicas is the maximum number of replicas allowed.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000000
	MaxReplicas int32 `json:"maxReplicas"`
}

// OptimizerConfiguration defines parameters for multi-target optimization across
// multiple ModelServing deployments with different hardware characteristics.
type OptimizerConfiguration struct {
	// Params contains the optimization parameters for each ModelServing group.
	// Each entry defines a different deployment type (e.g., different hardware) to optimize.
	// +kubebuilder:validation:MinItems=1
	Params []OptimizerParam `json:"params,omitempty"`
	// CostExpansionRatePercent defines the acceptable cost expansion percentage
	// when optimizing across multiple deployment types. A higher value allows more
	// flexibility in resource allocation but may increase overall costs.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=200
	// +optional
	CostExpansionRatePercent int32 `json:"costExpansionRatePercent,omitempty"`
}

// Target defines a ModelServing deployment that can be monitored and scaled.
type Target struct {
	// TargetRef references the ModelServing object to monitor and scale.
	TargetRef corev1.ObjectReference `json:"targetRef"`
	// AdditionalMatchLabels provides additional label selectors to refine
	// which pods within the ModelServing deployment should be monitored.
	// +optional
	AdditionalMatchLabels map[string]string `json:"additionalMatchLabels,omitempty"`
	// MetricEndpoint configures how to scrape metrics from the target pods.
	// If not specified, defaults to port 8100 and path "/metrics".
	// +optional
	MetricEndpoint MetricEndpoint `json:"metricEndpoint,omitempty"`
}

// OptimizerParam defines optimization parameters for a specific ModelServing deployment type.
type OptimizerParam struct {
	// Target specifies the ModelServing deployment and its monitoring configuration.
	Target Target `json:"target,omitempty"`
	// Cost represents the relative cost factor for this deployment type.
	// Used in optimization calculations to balance performance vs. cost.
	// +kubebuilder:validation:Minimum=0
	// +optional
	Cost int32 `json:"cost,omitempty"`
	// MinReplicas is the minimum number of replicas to maintain for this deployment type.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000000
	MinReplicas int32 `json:"minReplicas"`
	// MaxReplicas is the maximum number of replicas allowed for this deployment type.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000000
	MaxReplicas int32 `json:"maxReplicas"`
}

// AutoscalingPolicyBinding binds AutoscalingPolicy rules to specific ModelServing deployments,
// enabling either traditional metric-based scaling or multi-target optimization across
// heterogeneous hardware deployments.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
type AutoscalingPolicyBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoscalingPolicyBindingSpec   `json:"spec,omitempty"`
	Status AutoscalingPolicyBindingStatus `json:"status,omitempty"`
}

// AutoscalingPolicyBindingStatus defines the observed state of AutoscalingPolicyBinding.
type AutoscalingPolicyBindingStatus struct {
}

// +kubebuilder:object:root=true

// AutoscalingPolicyBindingList contains a list of AutoscalingPolicyBinding objects.
type AutoscalingPolicyBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoscalingPolicyBinding `json:"items"`
}
