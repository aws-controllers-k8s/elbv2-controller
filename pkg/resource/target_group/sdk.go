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

package target_group

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	smithy "github.com/aws/smithy-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &svcsdk.Client{}
	_ = &svcapitypes.TargetGroup{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
	_ = &ackcondition.NotManagedMessage
	_ = &reflect.Value{}
	_ = fmt.Sprintf("")
	_ = &ackrequeue.NoRequeue{}
	_ = &aws.Config{}
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkFind")
	defer func() {
		exit(err)
	}()
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromReadManyInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newListRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DescribeTargetGroupsOutput
	resp, err = rm.sdkapi.DescribeTargetGroups(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "DescribeTargetGroups", err)
	if err != nil {
		var awsErr smithy.APIError
		if errors.As(err, &awsErr) && awsErr.ErrorCode() == "TargetGroupNotFound" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	found := false
	for _, elem := range resp.TargetGroups {
		if elem.HealthCheckEnabled != nil {
			ko.Spec.HealthCheckEnabled = elem.HealthCheckEnabled
		} else {
			ko.Spec.HealthCheckEnabled = nil
		}
		if elem.HealthCheckIntervalSeconds != nil {
			healthCheckIntervalSecondsCopy := int64(*elem.HealthCheckIntervalSeconds)
			ko.Spec.HealthCheckIntervalSeconds = &healthCheckIntervalSecondsCopy
		} else {
			ko.Spec.HealthCheckIntervalSeconds = nil
		}
		if elem.HealthCheckPath != nil {
			ko.Spec.HealthCheckPath = elem.HealthCheckPath
		} else {
			ko.Spec.HealthCheckPath = nil
		}
		if elem.HealthCheckPort != nil {
			ko.Spec.HealthCheckPort = elem.HealthCheckPort
		} else {
			ko.Spec.HealthCheckPort = nil
		}
		if elem.HealthCheckProtocol != "" {
			ko.Spec.HealthCheckProtocol = aws.String(string(elem.HealthCheckProtocol))
		} else {
			ko.Spec.HealthCheckProtocol = nil
		}
		if elem.HealthCheckTimeoutSeconds != nil {
			healthCheckTimeoutSecondsCopy := int64(*elem.HealthCheckTimeoutSeconds)
			ko.Spec.HealthCheckTimeoutSeconds = &healthCheckTimeoutSecondsCopy
		} else {
			ko.Spec.HealthCheckTimeoutSeconds = nil
		}
		if elem.HealthyThresholdCount != nil {
			healthyThresholdCountCopy := int64(*elem.HealthyThresholdCount)
			ko.Spec.HealthyThresholdCount = &healthyThresholdCountCopy
		} else {
			ko.Spec.HealthyThresholdCount = nil
		}
		if elem.IpAddressType != "" {
			ko.Spec.IPAddressType = aws.String(string(elem.IpAddressType))
		} else {
			ko.Spec.IPAddressType = nil
		}
		if elem.LoadBalancerArns != nil {
			ko.Status.LoadBalancerARNs = aws.StringSlice(elem.LoadBalancerArns)
		} else {
			ko.Status.LoadBalancerARNs = nil
		}
		if elem.Matcher != nil {
			f9 := &svcapitypes.Matcher{}
			if elem.Matcher.GrpcCode != nil {
				f9.GRPCCode = elem.Matcher.GrpcCode
			}
			if elem.Matcher.HttpCode != nil {
				f9.HTTPCode = elem.Matcher.HttpCode
			}
			ko.Spec.Matcher = f9
		} else {
			ko.Spec.Matcher = nil
		}
		if elem.Port != nil {
			portCopy := int64(*elem.Port)
			ko.Spec.Port = &portCopy
		} else {
			ko.Spec.Port = nil
		}
		if elem.Protocol != "" {
			ko.Spec.Protocol = aws.String(string(elem.Protocol))
		} else {
			ko.Spec.Protocol = nil
		}
		if elem.ProtocolVersion != nil {
			ko.Spec.ProtocolVersion = elem.ProtocolVersion
		} else {
			ko.Spec.ProtocolVersion = nil
		}
		if elem.TargetGroupArn != nil {
			if ko.Status.ACKResourceMetadata == nil {
				ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			}
			tmpARN := ackv1alpha1.AWSResourceName(*elem.TargetGroupArn)
			ko.Status.ACKResourceMetadata.ARN = &tmpARN
		}
		if elem.TargetGroupName != nil {
			ko.Spec.Name = elem.TargetGroupName
		} else {
			ko.Spec.Name = nil
		}
		if elem.TargetType != "" {
			ko.Spec.TargetType = aws.String(string(elem.TargetType))
		} else {
			ko.Spec.TargetType = nil
		}
		if elem.UnhealthyThresholdCount != nil {
			unhealthyThresholdCountCopy := int64(*elem.UnhealthyThresholdCount)
			ko.Spec.UnhealthyThresholdCount = &unhealthyThresholdCountCopy
		} else {
			ko.Spec.UnhealthyThresholdCount = nil
		}
		if elem.VpcId != nil {
			ko.Spec.VPCID = elem.VpcId
		} else {
			ko.Spec.VPCID = nil
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}

	rm.setStatusDefaults(ko)
	err = rm.describeTargets(ctx, &resource{ko})
	if err != nil {
		return nil, err
	}

	rm.setStatusDefaults(ko)
	ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
	if err != nil {
		return nil, err
	}

	return &resource{ko}, nil
}

// requiredFieldsMissingFromReadManyInput returns true if there are any fields
// for the ReadMany Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadManyInput(
	r *resource,
) bool {
	return r.ko.Spec.Name == nil

}

// newListRequestPayload returns SDK-specific struct for the HTTP request
// payload of the List API call for the resource
func (rm *resourceManager) newListRequestPayload(
	r *resource,
) (*svcsdk.DescribeTargetGroupsInput, error) {
	res := &svcsdk.DescribeTargetGroupsInput{}

	if r.ko.Spec.Name != nil {
		f2 := []string{}
		f2 = append(f2, *r.ko.Spec.Name)
		res.Names = f2
	}

	return res, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a copy of the resource with resource fields (in both Spec and
// Status) filled in with values from the CREATE API operation's Output shape.
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	desired *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkCreate")
	defer func() {
		exit(err)
	}()
	input, err := rm.newCreateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.CreateTargetGroupOutput
	_ = resp
	resp, err = rm.sdkapi.CreateTargetGroup(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateTargetGroup", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	found := false
	for _, elem := range resp.TargetGroups {
		if elem.HealthCheckEnabled != nil {
			ko.Spec.HealthCheckEnabled = elem.HealthCheckEnabled
		} else {
			ko.Spec.HealthCheckEnabled = nil
		}
		if elem.HealthCheckIntervalSeconds != nil {
			healthCheckIntervalSecondsCopy := int64(*elem.HealthCheckIntervalSeconds)
			ko.Spec.HealthCheckIntervalSeconds = &healthCheckIntervalSecondsCopy
		} else {
			ko.Spec.HealthCheckIntervalSeconds = nil
		}
		if elem.HealthCheckPath != nil {
			ko.Spec.HealthCheckPath = elem.HealthCheckPath
		} else {
			ko.Spec.HealthCheckPath = nil
		}
		if elem.HealthCheckPort != nil {
			ko.Spec.HealthCheckPort = elem.HealthCheckPort
		} else {
			ko.Spec.HealthCheckPort = nil
		}
		if elem.HealthCheckProtocol != "" {
			ko.Spec.HealthCheckProtocol = aws.String(string(elem.HealthCheckProtocol))
		} else {
			ko.Spec.HealthCheckProtocol = nil
		}
		if elem.HealthCheckTimeoutSeconds != nil {
			healthCheckTimeoutSecondsCopy := int64(*elem.HealthCheckTimeoutSeconds)
			ko.Spec.HealthCheckTimeoutSeconds = &healthCheckTimeoutSecondsCopy
		} else {
			ko.Spec.HealthCheckTimeoutSeconds = nil
		}
		if elem.HealthyThresholdCount != nil {
			healthyThresholdCountCopy := int64(*elem.HealthyThresholdCount)
			ko.Spec.HealthyThresholdCount = &healthyThresholdCountCopy
		} else {
			ko.Spec.HealthyThresholdCount = nil
		}
		if elem.IpAddressType != "" {
			ko.Spec.IPAddressType = aws.String(string(elem.IpAddressType))
		} else {
			ko.Spec.IPAddressType = nil
		}
		if elem.LoadBalancerArns != nil {
			ko.Status.LoadBalancerARNs = aws.StringSlice(elem.LoadBalancerArns)
		} else {
			ko.Status.LoadBalancerARNs = nil
		}
		if elem.Matcher != nil {
			f9 := &svcapitypes.Matcher{}
			if elem.Matcher.GrpcCode != nil {
				f9.GRPCCode = elem.Matcher.GrpcCode
			}
			if elem.Matcher.HttpCode != nil {
				f9.HTTPCode = elem.Matcher.HttpCode
			}
			ko.Spec.Matcher = f9
		} else {
			ko.Spec.Matcher = nil
		}
		if elem.Port != nil {
			portCopy := int64(*elem.Port)
			ko.Spec.Port = &portCopy
		} else {
			ko.Spec.Port = nil
		}
		if elem.Protocol != "" {
			ko.Spec.Protocol = aws.String(string(elem.Protocol))
		} else {
			ko.Spec.Protocol = nil
		}
		if elem.ProtocolVersion != nil {
			ko.Spec.ProtocolVersion = elem.ProtocolVersion
		} else {
			ko.Spec.ProtocolVersion = nil
		}
		if elem.TargetGroupArn != nil {
			if ko.Status.ACKResourceMetadata == nil {
				ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			}
			tmpARN := ackv1alpha1.AWSResourceName(*elem.TargetGroupArn)
			ko.Status.ACKResourceMetadata.ARN = &tmpARN
		}
		if elem.TargetGroupName != nil {
			ko.Spec.Name = elem.TargetGroupName
		} else {
			ko.Spec.Name = nil
		}
		if elem.TargetType != "" {
			ko.Spec.TargetType = aws.String(string(elem.TargetType))
		} else {
			ko.Spec.TargetType = nil
		}
		if elem.UnhealthyThresholdCount != nil {
			unhealthyThresholdCountCopy := int64(*elem.UnhealthyThresholdCount)
			ko.Spec.UnhealthyThresholdCount = &unhealthyThresholdCountCopy
		} else {
			ko.Spec.UnhealthyThresholdCount = nil
		}
		if elem.VpcId != nil {
			ko.Spec.VPCID = elem.VpcId
		} else {
			ko.Spec.VPCID = nil
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}

	rm.setStatusDefaults(ko)
	if ko.Spec.Targets != nil {
		return nil, ackrequeue.NeededAfter(fmt.Errorf("Requing due to register targets in UPDATE"), RequeueAfterUpdateDuration)
	}

	rm.setStatusDefaults(ko)
	if ko.Spec.Tags != nil {
		return nil, ackrequeue.NeededAfter(fmt.Errorf("Requing due to tags in CREATE"), RequeueAfterUpdateDuration)
	}
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateTargetGroupInput, error) {
	res := &svcsdk.CreateTargetGroupInput{}

	if r.ko.Spec.HealthCheckEnabled != nil {
		res.HealthCheckEnabled = r.ko.Spec.HealthCheckEnabled
	}
	if r.ko.Spec.HealthCheckIntervalSeconds != nil {
		healthCheckIntervalSecondsCopy0 := *r.ko.Spec.HealthCheckIntervalSeconds
		if healthCheckIntervalSecondsCopy0 > math.MaxInt32 || healthCheckIntervalSecondsCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field HealthCheckIntervalSeconds is of type int32")
		}
		healthCheckIntervalSecondsCopy := int32(healthCheckIntervalSecondsCopy0)
		res.HealthCheckIntervalSeconds = &healthCheckIntervalSecondsCopy
	}
	if r.ko.Spec.HealthCheckPath != nil {
		res.HealthCheckPath = r.ko.Spec.HealthCheckPath
	}
	if r.ko.Spec.HealthCheckPort != nil {
		res.HealthCheckPort = r.ko.Spec.HealthCheckPort
	}
	if r.ko.Spec.HealthCheckProtocol != nil {
		res.HealthCheckProtocol = svcsdktypes.ProtocolEnum(*r.ko.Spec.HealthCheckProtocol)
	}
	if r.ko.Spec.HealthCheckTimeoutSeconds != nil {
		healthCheckTimeoutSecondsCopy0 := *r.ko.Spec.HealthCheckTimeoutSeconds
		if healthCheckTimeoutSecondsCopy0 > math.MaxInt32 || healthCheckTimeoutSecondsCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field HealthCheckTimeoutSeconds is of type int32")
		}
		healthCheckTimeoutSecondsCopy := int32(healthCheckTimeoutSecondsCopy0)
		res.HealthCheckTimeoutSeconds = &healthCheckTimeoutSecondsCopy
	}
	if r.ko.Spec.HealthyThresholdCount != nil {
		healthyThresholdCountCopy0 := *r.ko.Spec.HealthyThresholdCount
		if healthyThresholdCountCopy0 > math.MaxInt32 || healthyThresholdCountCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field HealthyThresholdCount is of type int32")
		}
		healthyThresholdCountCopy := int32(healthyThresholdCountCopy0)
		res.HealthyThresholdCount = &healthyThresholdCountCopy
	}
	if r.ko.Spec.IPAddressType != nil {
		res.IpAddressType = svcsdktypes.TargetGroupIpAddressTypeEnum(*r.ko.Spec.IPAddressType)
	}
	if r.ko.Spec.Matcher != nil {
		f8 := &svcsdktypes.Matcher{}
		if r.ko.Spec.Matcher.GRPCCode != nil {
			f8.GrpcCode = r.ko.Spec.Matcher.GRPCCode
		}
		if r.ko.Spec.Matcher.HTTPCode != nil {
			f8.HttpCode = r.ko.Spec.Matcher.HTTPCode
		}
		res.Matcher = f8
	}
	if r.ko.Spec.Name != nil {
		res.Name = r.ko.Spec.Name
	}
	if r.ko.Spec.Port != nil {
		portCopy0 := *r.ko.Spec.Port
		if portCopy0 > math.MaxInt32 || portCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field Port is of type int32")
		}
		portCopy := int32(portCopy0)
		res.Port = &portCopy
	}
	if r.ko.Spec.Protocol != nil {
		res.Protocol = svcsdktypes.ProtocolEnum(*r.ko.Spec.Protocol)
	}
	if r.ko.Spec.ProtocolVersion != nil {
		res.ProtocolVersion = r.ko.Spec.ProtocolVersion
	}
	if r.ko.Spec.Tags != nil {
		f13 := []svcsdktypes.Tag{}
		for _, f13iter := range r.ko.Spec.Tags {
			f13elem := &svcsdktypes.Tag{}
			if f13iter.Key != nil {
				f13elem.Key = f13iter.Key
			}
			if f13iter.Value != nil {
				f13elem.Value = f13iter.Value
			}
			f13 = append(f13, *f13elem)
		}
		res.Tags = f13
	}
	if r.ko.Spec.TargetType != nil {
		res.TargetType = svcsdktypes.TargetTypeEnum(*r.ko.Spec.TargetType)
	}
	if r.ko.Spec.UnhealthyThresholdCount != nil {
		unhealthyThresholdCountCopy0 := *r.ko.Spec.UnhealthyThresholdCount
		if unhealthyThresholdCountCopy0 > math.MaxInt32 || unhealthyThresholdCountCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field UnhealthyThresholdCount is of type int32")
		}
		unhealthyThresholdCountCopy := int32(unhealthyThresholdCountCopy0)
		res.UnhealthyThresholdCount = &unhealthyThresholdCountCopy
	}
	if r.ko.Spec.VPCID != nil {
		res.VpcId = r.ko.Spec.VPCID
	}

	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkUpdate")
	defer func() {
		exit(err)
	}()

	if delta.DifferentAt("Spec.Tags") {
		err = rm.updateTags(ctx, desired, latest)
		if err != nil {
			return nil, err
		}
	}
	if !delta.DifferentAt("Spec.Tags") {
		return desired, nil
	}
	
	if delta.DifferentAt("Spec.Targets") {
		added, removed := getTargetsDifference(latest.ko.Spec.Targets, desired.ko.Spec.Targets)
		arn := (string)(*latest.ko.Status.ACKResourceMetadata.ARN)
		if len(removed) > 0 {
			err = rm.deregisterTargets(ctx, arn, removed)
			if err != nil {
				return nil, err
			}
		}
		if len(added) > 0 {
			err = rm.registerTargets(ctx, arn, added)
			if err != nil {
				return nil, err
			}
		}
	}

	if !delta.DifferentExcept("Spec.Targets") {
		return desired, nil
	}
	input, err := rm.newUpdateRequestPayload(ctx, desired, delta)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.ModifyTargetGroupOutput
	_ = resp
	resp, err = rm.sdkapi.ModifyTargetGroup(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ModifyTargetGroup", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	found := false
	for _, elem := range resp.TargetGroups {
		if elem.HealthCheckEnabled != nil {
			ko.Spec.HealthCheckEnabled = elem.HealthCheckEnabled
		} else {
			ko.Spec.HealthCheckEnabled = nil
		}
		if elem.HealthCheckIntervalSeconds != nil {
			healthCheckIntervalSecondsCopy := int64(*elem.HealthCheckIntervalSeconds)
			ko.Spec.HealthCheckIntervalSeconds = &healthCheckIntervalSecondsCopy
		} else {
			ko.Spec.HealthCheckIntervalSeconds = nil
		}
		if elem.HealthCheckPath != nil {
			ko.Spec.HealthCheckPath = elem.HealthCheckPath
		} else {
			ko.Spec.HealthCheckPath = nil
		}
		if elem.HealthCheckPort != nil {
			ko.Spec.HealthCheckPort = elem.HealthCheckPort
		} else {
			ko.Spec.HealthCheckPort = nil
		}
		if elem.HealthCheckProtocol != "" {
			ko.Spec.HealthCheckProtocol = aws.String(string(elem.HealthCheckProtocol))
		} else {
			ko.Spec.HealthCheckProtocol = nil
		}
		if elem.HealthCheckTimeoutSeconds != nil {
			healthCheckTimeoutSecondsCopy := int64(*elem.HealthCheckTimeoutSeconds)
			ko.Spec.HealthCheckTimeoutSeconds = &healthCheckTimeoutSecondsCopy
		} else {
			ko.Spec.HealthCheckTimeoutSeconds = nil
		}
		if elem.HealthyThresholdCount != nil {
			healthyThresholdCountCopy := int64(*elem.HealthyThresholdCount)
			ko.Spec.HealthyThresholdCount = &healthyThresholdCountCopy
		} else {
			ko.Spec.HealthyThresholdCount = nil
		}
		if elem.IpAddressType != "" {
			ko.Spec.IPAddressType = aws.String(string(elem.IpAddressType))
		} else {
			ko.Spec.IPAddressType = nil
		}
		if elem.LoadBalancerArns != nil {
			ko.Status.LoadBalancerARNs = aws.StringSlice(elem.LoadBalancerArns)
		} else {
			ko.Status.LoadBalancerARNs = nil
		}
		if elem.Matcher != nil {
			f9 := &svcapitypes.Matcher{}
			if elem.Matcher.GrpcCode != nil {
				f9.GRPCCode = elem.Matcher.GrpcCode
			}
			if elem.Matcher.HttpCode != nil {
				f9.HTTPCode = elem.Matcher.HttpCode
			}
			ko.Spec.Matcher = f9
		} else {
			ko.Spec.Matcher = nil
		}
		if elem.Port != nil {
			portCopy := int64(*elem.Port)
			ko.Spec.Port = &portCopy
		} else {
			ko.Spec.Port = nil
		}
		if elem.Protocol != "" {
			ko.Spec.Protocol = aws.String(string(elem.Protocol))
		} else {
			ko.Spec.Protocol = nil
		}
		if elem.ProtocolVersion != nil {
			ko.Spec.ProtocolVersion = elem.ProtocolVersion
		} else {
			ko.Spec.ProtocolVersion = nil
		}
		if elem.TargetGroupArn != nil {
			if ko.Status.ACKResourceMetadata == nil {
				ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			}
			tmpARN := ackv1alpha1.AWSResourceName(*elem.TargetGroupArn)
			ko.Status.ACKResourceMetadata.ARN = &tmpARN
		}
		if elem.TargetType != "" {
			ko.Spec.TargetType = aws.String(string(elem.TargetType))
		} else {
			ko.Spec.TargetType = nil
		}
		if elem.UnhealthyThresholdCount != nil {
			unhealthyThresholdCountCopy := int64(*elem.UnhealthyThresholdCount)
			ko.Spec.UnhealthyThresholdCount = &unhealthyThresholdCountCopy
		} else {
			ko.Spec.UnhealthyThresholdCount = nil
		}
		if elem.VpcId != nil {
			ko.Spec.VPCID = elem.VpcId
		} else {
			ko.Spec.VPCID = nil
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	ctx context.Context,
	r *resource,
	delta *ackcompare.Delta,
) (*svcsdk.ModifyTargetGroupInput, error) {
	res := &svcsdk.ModifyTargetGroupInput{}

	if r.ko.Spec.HealthCheckEnabled != nil {
		res.HealthCheckEnabled = r.ko.Spec.HealthCheckEnabled
	}
	if r.ko.Spec.HealthCheckIntervalSeconds != nil {
		healthCheckIntervalSecondsCopy0 := *r.ko.Spec.HealthCheckIntervalSeconds
		if healthCheckIntervalSecondsCopy0 > math.MaxInt32 || healthCheckIntervalSecondsCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field HealthCheckIntervalSeconds is of type int32")
		}
		healthCheckIntervalSecondsCopy := int32(healthCheckIntervalSecondsCopy0)
		res.HealthCheckIntervalSeconds = &healthCheckIntervalSecondsCopy
	}
	if r.ko.Spec.HealthCheckPath != nil {
		res.HealthCheckPath = r.ko.Spec.HealthCheckPath
	}
	if r.ko.Spec.HealthCheckPort != nil {
		res.HealthCheckPort = r.ko.Spec.HealthCheckPort
	}
	if r.ko.Spec.HealthCheckProtocol != nil {
		res.HealthCheckProtocol = svcsdktypes.ProtocolEnum(*r.ko.Spec.HealthCheckProtocol)
	}
	if r.ko.Spec.HealthCheckTimeoutSeconds != nil {
		healthCheckTimeoutSecondsCopy0 := *r.ko.Spec.HealthCheckTimeoutSeconds
		if healthCheckTimeoutSecondsCopy0 > math.MaxInt32 || healthCheckTimeoutSecondsCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field HealthCheckTimeoutSeconds is of type int32")
		}
		healthCheckTimeoutSecondsCopy := int32(healthCheckTimeoutSecondsCopy0)
		res.HealthCheckTimeoutSeconds = &healthCheckTimeoutSecondsCopy
	}
	if r.ko.Spec.HealthyThresholdCount != nil {
		healthyThresholdCountCopy0 := *r.ko.Spec.HealthyThresholdCount
		if healthyThresholdCountCopy0 > math.MaxInt32 || healthyThresholdCountCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field HealthyThresholdCount is of type int32")
		}
		healthyThresholdCountCopy := int32(healthyThresholdCountCopy0)
		res.HealthyThresholdCount = &healthyThresholdCountCopy
	}
	if r.ko.Spec.Matcher != nil {
		f7 := &svcsdktypes.Matcher{}
		if r.ko.Spec.Matcher.GRPCCode != nil {
			f7.GrpcCode = r.ko.Spec.Matcher.GRPCCode
		}
		if r.ko.Spec.Matcher.HTTPCode != nil {
			f7.HttpCode = r.ko.Spec.Matcher.HTTPCode
		}
		res.Matcher = f7
	}
	if r.ko.Status.ACKResourceMetadata != nil && r.ko.Status.ACKResourceMetadata.ARN != nil {
		res.TargetGroupArn = (*string)(r.ko.Status.ACKResourceMetadata.ARN)
	}
	if r.ko.Spec.UnhealthyThresholdCount != nil {
		unhealthyThresholdCountCopy0 := *r.ko.Spec.UnhealthyThresholdCount
		if unhealthyThresholdCountCopy0 > math.MaxInt32 || unhealthyThresholdCountCopy0 < math.MinInt32 {
			return nil, fmt.Errorf("error: field UnhealthyThresholdCount is of type int32")
		}
		unhealthyThresholdCountCopy := int32(unhealthyThresholdCountCopy0)
		res.UnhealthyThresholdCount = &unhealthyThresholdCountCopy
	}

	return res, nil
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer func() {
		exit(err)
	}()
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteTargetGroupOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteTargetGroup(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteTargetGroup", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteTargetGroupInput, error) {
	res := &svcsdk.DeleteTargetGroupInput{}

	if r.ko.Status.ACKResourceMetadata != nil && r.ko.Status.ACKResourceMetadata.ARN != nil {
		res.TargetGroupArn = (*string)(r.ko.Status.ACKResourceMetadata.ARN)
	}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.TargetGroup,
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.Region == nil {
		ko.Status.ACKResourceMetadata.Region = &rm.awsRegion
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions(
	r *resource,
	onSuccess bool,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	var recoverableCondition *ackv1alpha1.Condition = nil
	var syncCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeRecoverable {
			recoverableCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			syncCondition = condition
		}
	}
	var termError *ackerr.TerminalError
	if rm.terminalAWSError(err) || err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound || errors.As(err, &termError) {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type: ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		var errorMessage = ""
		if err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound || errors.As(err, &termError) {
			errorMessage = err.Error()
		} else {
			awsErr, _ := ackerr.AWSError(err)
			errorMessage = awsErr.Error()
		}
		terminalCondition.Status = corev1.ConditionTrue
		terminalCondition.Message = &errorMessage
	} else {
		// Clear the terminal condition if no longer present
		if terminalCondition != nil {
			terminalCondition.Status = corev1.ConditionFalse
			terminalCondition.Message = nil
		}
		// Handling Recoverable Conditions
		if err != nil {
			if recoverableCondition == nil {
				// Add a new Condition containing a non-terminal error
				recoverableCondition = &ackv1alpha1.Condition{
					Type: ackv1alpha1.ConditionTypeRecoverable,
				}
				ko.Status.Conditions = append(ko.Status.Conditions, recoverableCondition)
			}
			recoverableCondition.Status = corev1.ConditionTrue
			awsErr, _ := ackerr.AWSError(err)
			errorMessage := err.Error()
			if awsErr != nil {
				errorMessage = awsErr.Error()
			}
			recoverableCondition.Message = &errorMessage
		} else if recoverableCondition != nil {
			recoverableCondition.Status = corev1.ConditionFalse
			recoverableCondition.Message = nil
		}
	}
	// Required to avoid the "declared but not used" error in the default case
	_ = syncCondition
	if terminalCondition != nil || recoverableCondition != nil || syncCondition != nil {
		return &resource{ko}, true // updated
	}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
	if err == nil {
		return false
	}

	var terminalErr smithy.APIError
	if !errors.As(err, &terminalErr) {
		return false
	}
	switch terminalErr.ErrorCode() {
	case "InvalidConfigurationRequest":
		return true
	default:
		return false
	}
}
