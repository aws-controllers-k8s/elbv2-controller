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

package load_balancer

import (
	"context"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"
)

// setResourceAdditionalFields will describe the fields that are not return by the
// describeLoadBalancer API calls.
func (rm *resourceManager) setResourceAdditionalFields(
	ctx context.Context,
	ko *svcapitypes.LoadBalancer,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceAdditionalFields")
	defer exit(err)

	ko.Spec.Attributes, err = rm.getLoadBalancerAttributes(ctx, ko)
	if err != nil {
		return err
	}
	return nil
}

// getLoadBalancerAttributes returns the attributes of the load balancer.
func (rm *resourceManager) getLoadBalancerAttributes(
	ctx context.Context,
	ko *svcapitypes.LoadBalancer,
) ([]*svcapitypes.LoadBalancerAttribute, error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getLoadBalancerAttributes")
	var err error
	defer func() { exit(err) }()

	attributes := []*svcapitypes.LoadBalancerAttribute{}
	var resp *svcsdk.DescribeLoadBalancerAttributesOutput

	resp, err = rm.sdkapi.DescribeLoadBalancerAttributesWithContext(ctx, &svcsdk.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: (*string)(ko.Status.ACKResourceMetadata.ARN),
	})
	rm.metrics.RecordAPICall("READ_ONE", "DescribeLoadBalancerAttributes", err)
	if err != nil {
		return nil, err
	}

	// Convert the attributes SDK type to the k8s API type
	for _, attr := range resp.Attributes {
		attribute := &svcapitypes.LoadBalancerAttribute{
			Key:   attr.Key,
			Value: attr.Value,
		}
		attributes = append(attributes, attribute)
	}
	return attributes, nil
}

// attributesHaveChanged returns true if one of desired attributes (a) have
// drifted from the latest attributes (b).
func attributesHaveChanged(a, b []*svcapitypes.LoadBalancerAttribute) bool {
	for _, attrA := range a {
		if !containsExactAttribute(b, attrA) {
			return false
		}
	}
	return true
}

// containsExactAttribute returns true if the key is in the attributes slice
// and has the same value.
func containsExactAttribute(attributes []*svcapitypes.LoadBalancerAttribute, targetAttribute *svcapitypes.LoadBalancerAttribute) bool {
	for _, attribute := range attributes {
		if *attribute.Key == *targetAttribute.Key && *attribute.Value == *targetAttribute.Value {
			return true
		}
	}
	return false
}

// customPreCompare is a custom pre compare function that compares the attributes
// of the load balancer.
func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if !attributesHaveChanged(a.ko.Spec.Attributes, b.ko.Spec.Attributes) {
		delta.Add("Spec.Attributes", a.ko.Spec.Attributes, b.ko.Spec.Attributes)
	}
}

// customUpdateLoadBalancer is a custom update function that updates the attributes/tags
// of the load balancer.
func (rm *resourceManager) customUpdateLoadBalancer(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	var err error
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateLoadBalancer")
	defer func() { exit(err) }()

	if delta.DifferentAt("Spec.Attributes") && len(desired.ko.Spec.Attributes) > 0 {
		if err := rm.updateLoadBalancerAttributes(ctx, desired, latest); err != nil {
			return nil, err
		}
	}
	// Leaving room for tag updates...
	// if delta.DifferentAt("Spec.Tags") {...}

	return desired, nil
}

// updateLoadBalancerAttributes updates the attributes of the load balancer.
func (rm *resourceManager) updateLoadBalancerAttributes(
	ctx context.Context,
	desired *resource,
	latest *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateLoadBalancerAttributes")
	var err error
	defer func() { exit(err) }()

	sdkAttributes := []*svcsdk.LoadBalancerAttribute{}
	for _, attr := range desired.ko.Spec.Attributes {
		// Only set non-empty attributes
		if attr.Key == nil || attr.Value == nil || *attr.Key == "" || *attr.Value == "" {
			continue
		}
		sdkAttribute := &svcsdk.LoadBalancerAttribute{
			Key:   aws.String(*attr.Key),
			Value: aws.String(*attr.Value),
		}
		sdkAttributes = append(sdkAttributes, sdkAttribute)
	}

	input := &svcsdk.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: (*string)(desired.ko.Status.ACKResourceMetadata.ARN),
		Attributes:      sdkAttributes,
	}
	_, err = rm.sdkapi.ModifyLoadBalancerAttributesWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ModifyLoadBalancerAttributes", err)
	if err != nil {
		return err
	}
	return nil
}
