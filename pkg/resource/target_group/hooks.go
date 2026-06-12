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

package target_group

import (
	"context"
	"fmt"
	"time"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

const (
	// AnnotationTargetManagement controls how the controller manages targets
	// registered with the target group. When set to "ignore", the controller
	// will not read, register, or deregister targets — allowing an external
	// controller to manage target registration independently.
	AnnotationTargetManagement = "elbv2.services.k8s.aws/target-management"
)

var (
	RequeueAfterUpdateDuration = 5 * time.Second
)

// isTargetManagementIgnored returns true if the resource has the
// AnnotationTargetManagement annotation set to "ignore", indicating that
// target registration should be managed by an external controller.
func isTargetManagementIgnored(r *resource) bool {
	if r == nil || r.ko == nil {
		return false
	}
	annotations := r.ko.GetAnnotations()
	if annotations == nil {
		return false
	}
	return annotations[AnnotationTargetManagement] == "ignore"
}

func customCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	compareTargetDescription(delta, a, b)
	compareTargetGroupAttributes(delta, a, b)
}

// compareTargetGroupAttributes adds a delta entry when one of the attributes
// declared in the desired spec is missing or has a different value in the
// latest (AWS) state. Attributes is marked compare.is_ignored in
// generator.yaml, so ACK does not compare it automatically.
//
// Only attributes the user explicitly declares are managed. Attributes the
// user omits are left as-is on AWS, mirroring the LoadBalancer resource. This
// avoids the infinite reconcile that resetting omitted attributes to their
// server-side defaults would cause.
func compareTargetGroupAttributes(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if targetGroupAttributesHaveDrifted(a.ko.Spec.Attributes, b.ko.Spec.Attributes) {
		delta.Add("Spec.Attributes", a.ko.Spec.Attributes, b.ko.Spec.Attributes)
	}
}

// targetGroupAttributesHaveDrifted returns true if any attribute declared in
// the desired spec is missing or has a different value in the latest state.
func targetGroupAttributesHaveDrifted(desired, latest []*svcapitypes.TargetGroupAttribute) bool {
	for _, attr := range desired {
		if !containsExactTargetGroupAttribute(latest, attr) {
			return true
		}
	}
	return false
}

// containsExactTargetGroupAttribute returns true if the key is in the attributes slice
// and has the same value. Nil-value attributes with the same key are considered equal.
func containsExactTargetGroupAttribute(attributes []*svcapitypes.TargetGroupAttribute, targetAttribute *svcapitypes.TargetGroupAttribute) bool {
	for _, attribute := range attributes {
		if attribute.Key != nil && targetAttribute.Key != nil &&
			*attribute.Key == *targetAttribute.Key &&
			((attribute.Value == nil && targetAttribute.Value == nil) ||
				(attribute.Value != nil && targetAttribute.Value != nil &&
					*attribute.Value == *targetAttribute.Value)) {
			return true
		}
	}
	return false
}

// getTargetGroupAttributes returns the attributes of the target group from AWS.
func (rm *resourceManager) getTargetGroupAttributes(
	ctx context.Context,
	ko *svcapitypes.TargetGroup,
) ([]*svcapitypes.TargetGroupAttribute, error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getTargetGroupAttributes")
	var err error
	defer func() {
		exit(err)
	}()

	attributes := []*svcapitypes.TargetGroupAttribute{}
	var resp *svcsdk.DescribeTargetGroupAttributesOutput

	if ko.Status.ACKResourceMetadata == nil || ko.Status.ACKResourceMetadata.ARN == nil {
		return nil, fmt.Errorf("target group ARN is not yet available")
	}
	resp, err = rm.sdkapi.DescribeTargetGroupAttributes(ctx, &svcsdk.DescribeTargetGroupAttributesInput{
		TargetGroupArn: (*string)(ko.Status.ACKResourceMetadata.ARN),
	})
	rm.metrics.RecordAPICall("READ_ONE", "DescribeTargetGroupAttributes", err)
	if err != nil {
		return nil, err
	}

	// Convert the attributes SDK type to the k8s API type
	for _, attr := range resp.Attributes {
		attribute := &svcapitypes.TargetGroupAttribute{
			Key:   attr.Key,
			Value: attr.Value,
		}
		attributes = append(attributes, attribute)
	}
	return attributes, nil
}

// updateTargetGroupAttributes pushes the attributes declared in the desired
// spec to AWS via ModifyTargetGroupAttributes. Only attributes with non-empty
// key and value are sent; omitted attributes are left untouched on AWS.
func (rm *resourceManager) updateTargetGroupAttributes(
	ctx context.Context,
	desired *resource,
	latest *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateTargetGroupAttributes")
	var err error
	defer func() {
		exit(err)
	}()

	if latest.ko.Status.ACKResourceMetadata == nil || latest.ko.Status.ACKResourceMetadata.ARN == nil {
		return fmt.Errorf("target group ARN is not yet available")
	}

	input := &svcsdk.ModifyTargetGroupAttributesInput{
		TargetGroupArn: (*string)(latest.ko.Status.ACKResourceMetadata.ARN),
		Attributes:     []svcsdktypes.TargetGroupAttribute{},
	}
	for _, attr := range desired.ko.Spec.Attributes {
		if attr.Key == nil || attr.Value == nil || *attr.Key == "" || *attr.Value == "" {
			continue
		}
		input.Attributes = append(input.Attributes, svcsdktypes.TargetGroupAttribute{
			Key:   attr.Key,
			Value: attr.Value,
		})
	}
	if len(input.Attributes) == 0 {
		return nil
	}

	_, err = rm.sdkapi.ModifyTargetGroupAttributes(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ModifyTargetGroupAttributes", err)
	if err != nil {
		return err
	}

	return nil
}

func compareTargetDescription(
	delta *ackcompare.Delta,
	desired *resource,
	latest *resource,
) {
	if len(desired.ko.Spec.Targets) != len(latest.ko.Spec.Targets) {
		delta.Add("Spec.Targets", desired.ko.Spec.Targets, latest.ko.Spec.Targets)
	} else if len(desired.ko.Spec.Targets) > 0 {
		added, removed := getTargetsDifference(latest.ko.Spec.Targets, desired.ko.Spec.Targets)

		if len(added) > 0 || len(removed) > 0 {
			delta.Add("Spec.Targets", added, removed)
		}
		return

	}
}

func areDifferentTarget(latest, desired *svcapitypes.TargetDescription) bool {
	if (latest == nil && desired != nil) || (latest != nil && desired == nil) {
		return true
	}

	if latest.Port != nil && desired.Port == nil {
		desired.Port = latest.Port
	}
	if latest.AvailabilityZone != nil && desired.AvailabilityZone == nil {
		desired.AvailabilityZone = latest.AvailabilityZone
	}
	if (latest.ID == nil && desired.ID != nil) || (latest.ID != nil && desired.ID == nil) ||
		(latest.ID != nil && desired.ID != nil && *latest.ID != *desired.ID) ||
		(latest.Port == nil && desired.Port != nil) || (latest.Port != nil && desired.Port != nil && *latest.Port != *desired.Port) ||
		(latest.AvailabilityZone == nil && desired.AvailabilityZone != nil) ||
		(latest.AvailabilityZone != nil && desired.AvailabilityZone != nil && *latest.AvailabilityZone != *desired.AvailabilityZone) {
		return true
	}

	return false
}

func getTargetsDifference(
	latest []*svcapitypes.TargetDescription,
	desired []*svcapitypes.TargetDescription,
) (added []*svcapitypes.TargetDescription, removed []*svcapitypes.TargetDescription) {

	toAdd := make([]*svcapitypes.TargetDescription, 0, min(len(latest), len(desired)))
	toDelete := make([]*svcapitypes.TargetDescription, 0, min(len(latest), len(desired)))

	am := make(map[string]*svcapitypes.TargetDescription)

	for _, v := range latest {
		am[*v.ID] = v
	}

	for _, v := range desired {
		if t, ok := am[*v.ID]; !ok || areDifferentTarget(t, v) {
			toAdd = append(toAdd, v)
		}
	}

	bm := make(map[string]*svcapitypes.TargetDescription)
	for _, v := range desired {
		bm[*v.ID] = v
	}

	for _, v := range latest {
		if _, ok := bm[*v.ID]; !ok {
			toDelete = append(toDelete, v)
		}
	}

	return toAdd, toDelete
}

func (rm *resourceManager) registerTargets(
	ctx context.Context,
	arn string,
	targets []*svcapitypes.TargetDescription,
) (err error) {

	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.registerTargets")
	defer func() { exit(err) }()

	input := &svcsdk.RegisterTargetsInput{
		TargetGroupArn: &arn,
		Targets:        apifyTargetDescription(targets),
	}
	_, err = rm.sdkapi.RegisterTargets(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "RegisterTargets", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deregisterTargets(
	ctx context.Context,
	arn string,
	targets []*svcapitypes.TargetDescription,
) (err error) {
	if len(targets) == 0 {
		return nil
	}

	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deregisterTargets")
	defer func() { exit(err) }()

	input := &svcsdk.DeregisterTargetsInput{
		TargetGroupArn: &arn,
		Targets:        apifyTargetDescription(targets),
	}
	_, err = rm.sdkapi.DeregisterTargets(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "DeregisterTargets", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) describeTargets(
	ctx context.Context,
	res *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.describeTargets")
	defer func() { exit(err) }()

	input := &svcsdk.DescribeTargetHealthInput{
		TargetGroupArn: (*string)(res.ko.Status.ACKResourceMetadata.ARN),
	}
	resp, err := rm.sdkapi.DescribeTargetHealth(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "DescribeTargetHealth", err)
	if err != nil {
		return err
	}

	targetHealthPtrs := make([]*svcsdktypes.TargetHealthDescription, len(resp.TargetHealthDescriptions))
	for i := range resp.TargetHealthDescriptions {
		targetHealthPtrs[i] = &resp.TargetHealthDescriptions[i]
	}
	res.ko.Spec.Targets = extractTargetDescription(targetHealthPtrs)
	return nil
}

func apifyTargetDescription(target []*svcapitypes.TargetDescription) []svcsdktypes.TargetDescription {
	convertedTarget := make([]svcsdktypes.TargetDescription, len(target))
	for i, t := range target {
		td := svcsdktypes.TargetDescription{
			Id:               t.ID,
			AvailabilityZone: t.AvailabilityZone,
		}
		if t.Port != nil {
			td.Port = aws.Int32(int32(*t.Port))
		}
		convertedTarget[i] = td
	}
	return convertedTarget
}

func extractTargetDescription(targetHealth []*svcsdktypes.TargetHealthDescription) []*svcapitypes.TargetDescription {
	convertedTarget := make([]*svcapitypes.TargetDescription, 0, len(targetHealth))
	for _, t := range targetHealth {
		if t.Target == nil {
			continue
		}
		td := &svcapitypes.TargetDescription{
			ID:               t.Target.Id,
			AvailabilityZone: t.Target.AvailabilityZone,
		}
		if t.Target.Port != nil {
			td.Port = aws.Int64(int64(*t.Target.Port))
		}
		convertedTarget = append(convertedTarget, td)
	}
	return convertedTarget
}
