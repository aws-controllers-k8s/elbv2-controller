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
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

var (
	RequeueAfterUpdateDuration = 5 * time.Second
)

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

// validateTargets rejects a spec that RegisterTargets can never accept. An ID
// is required on every target, so a missing one is a terminal error the user
// must fix rather than something to retry.
func validateTargets(targets []*svcapitypes.TargetDescription) error {
	for i, t := range targets {
		if t == nil || t.ID == nil || *t.ID == "" {
			return ackerr.NewTerminalError(
				fmt.Errorf("spec.targets[%d]: id is required", i),
			)
		}
	}
	return nil
}

// areDifferentTarget reports whether the two targets differ in any part of the
// identity RegisterTargets and DeregisterTargets operate on: ID, Port and
// AvailabilityZone. It does not modify either argument.
func areDifferentTarget(latest, desired *svcapitypes.TargetDescription) bool {
	if latest == nil || desired == nil {
		return latest != nil || desired != nil
	}
	return !equalPtr(latest.ID, desired.ID) ||
		!equalPtr(latest.Port, desired.Port) ||
		!equalPtr(latest.AvailabilityZone, desired.AvailabilityZone)
}

// targetMatchesIgnoringUnset reports whether latest is the target that desired
// identifies, treating a Port or AvailabilityZone the user omitted as matching
// whatever AWS reports: AWS fills both in on read-back.
func targetMatchesIgnoringUnset(latest, desired *svcapitypes.TargetDescription) bool {
	if latest == nil || desired == nil {
		return false
	}
	if !equalPtr(latest.ID, desired.ID) {
		return false
	}
	if desired.Port != nil && !equalPtr(latest.Port, desired.Port) {
		return false
	}
	if desired.AvailabilityZone != nil && !equalPtr(latest.AvailabilityZone, desired.AvailabilityZone) {
		return false
	}
	return true
}

func equalPtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// findTarget returns the index of an unmatched latest target that desired
// identifies, or -1. candidates are the indices of the latest targets sharing
// desired's ID.
func findTarget(
	latest []*svcapitypes.TargetDescription,
	candidates []int,
	matched []bool,
	desired *svcapitypes.TargetDescription,
	exact bool,
) int {
	for _, i := range candidates {
		if matched[i] {
			continue
		}
		if exact {
			if !areDifferentTarget(latest[i], desired) {
				return i
			}
			continue
		}
		if targetMatchesIgnoringUnset(latest[i], desired) {
			return i
		}
	}
	return -1
}

// getTargetsDifference pairs each desired target with the latest target of the
// same identity -- (ID, Port, AvailabilityZone), not ID alone, since the same
// instance ID or IP may be registered on several ports -- and returns the
// targets to register and to deregister. Exact identities are paired first, so
// a target whose Port the user omitted cannot claim a fully specified one's.
func getTargetsDifference(
	latest []*svcapitypes.TargetDescription,
	desired []*svcapitypes.TargetDescription,
) (added []*svcapitypes.TargetDescription, removed []*svcapitypes.TargetDescription) {

	added = make([]*svcapitypes.TargetDescription, 0, len(desired))
	removed = make([]*svcapitypes.TargetDescription, 0, len(latest))

	candidatesByID := make(map[string][]int, len(latest))
	for i, t := range latest {
		if t == nil || t.ID == nil {
			continue
		}
		candidatesByID[*t.ID] = append(candidatesByID[*t.ID], i)
	}

	matched := make([]bool, len(latest))
	unpaired := make([]*svcapitypes.TargetDescription, 0, len(desired))

	for _, d := range desired {
		if d == nil || d.ID == nil {
			// Invalid spec; reported as an addition so validateTargets gets to
			// turn it into a terminal error on the update path.
			added = append(added, d)
			continue
		}
		if i := findTarget(latest, candidatesByID[*d.ID], matched, d, true); i >= 0 {
			matched[i] = true
			continue
		}
		unpaired = append(unpaired, d)
	}

	for _, d := range unpaired {
		if i := findTarget(latest, candidatesByID[*d.ID], matched, d, false); i >= 0 {
			matched[i] = true
			continue
		}
		added = append(added, d)
	}

	for i, t := range latest {
		if !matched[i] {
			removed = append(removed, t)
		}
	}

	return added, removed
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
