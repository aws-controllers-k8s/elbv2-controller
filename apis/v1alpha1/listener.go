// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListenerSpec defines the desired state of Listener.
//
// Information about a listener.
type ListenerSpec struct {

	// [TLS listeners] The name of the Application-Layer Protocol Negotiation (ALPN)
	// policy. You can specify one policy name. The following are the possible values:
	//
	//   - HTTP1Only
	//
	//   - HTTP2Only
	//
	//   - HTTP2Optional
	//
	//   - HTTP2Preferred
	//
	//   - None
	//
	// For more information, see ALPN policies (https://docs.aws.amazon.com/elasticloadbalancing/latest/network/create-tls-listener.html#alpn-policies)
	// in the Network Load Balancers Guide.
	AlpnPolicy []*string `json:"alpnPolicy,omitempty"`
	// [HTTPS and TLS listeners] The default certificate for the listener. You must
	// provide exactly one certificate. Set CertificateArn to the certificate ARN
	// but do not set IsDefault.
	Certificates []*Certificate `json:"certificates,omitempty"`
	// The actions for the default rule.
	// +kubebuilder:validation:Required
	DefaultActions []*Action `json:"defaultActions"`
	// The Amazon Resource Name (ARN) of the load balancer.
	LoadBalancerARN *string                                  `json:"loadBalancerARN,omitempty"`
	LoadBalancerRef *ackv1alpha1.AWSResourceReferenceWrapper `json:"loadBalancerRef,omitempty"`
	// The mutual authentication configuration information.
	MutualAuthentication *MutualAuthenticationAttributes `json:"mutualAuthentication,omitempty"`
	// The port on which the load balancer is listening. You can't specify a port
	// for a Gateway Load Balancer.
	Port *int64 `json:"port,omitempty"`
	// The protocol for connections from clients to the load balancer. For Application
	// Load Balancers, the supported protocols are HTTP and HTTPS. For Network Load
	// Balancers, the supported protocols are TCP, TLS, UDP, and TCP_UDP. You can’t
	// specify the UDP or TCP_UDP protocol if dual-stack mode is enabled. You can't
	// specify a protocol for a Gateway Load Balancer.
	Protocol *string `json:"protocol,omitempty"`
	// [HTTPS and TLS listeners] The security policy that defines which protocols
	// and ciphers are supported.
	//
	// For more information, see Security policies (https://docs.aws.amazon.com/elasticloadbalancing/latest/application/create-https-listener.html#describe-ssl-policies)
	// in the Application Load Balancers Guide and Security policies (https://docs.aws.amazon.com/elasticloadbalancing/latest/network/create-tls-listener.html#describe-ssl-policies)
	// in the Network Load Balancers Guide.
	SSLPolicy *string `json:"sslPolicy,omitempty"`
	// The tags to assign to the listener.
	Tags []*Tag `json:"tags,omitempty"`
}

// ListenerStatus defines the observed state of Listener
type ListenerStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRs managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	// +kubebuilder:validation:Optional
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
}

// Listener is the Schema for the Listeners API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Listener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ListenerSpec   `json:"spec,omitempty"`
	Status            ListenerStatus `json:"status,omitempty"`
}

// ListenerList contains a list of Listener
// +kubebuilder:object:root=true
type ListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Listener `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Listener{}, &ListenerList{})
}
