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
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
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
	_, err = rm.sdkapi.RegisterTargetsWithContext(ctx, input)
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
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deregisterTargets")
	defer func() { exit(err) }()

	input := &svcsdk.DeregisterTargetsInput{
		TargetGroupArn: &arn,
		Targets:        apifyTargetDescription(targets),
	}
	_, err = rm.sdkapi.DeregisterTargetsWithContext(ctx, input)
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
	resp, err := rm.sdkapi.DescribeTargetHealthWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "DescribeTargetHealth", err)
	if err != nil {
		return err
	}

	res.ko.Spec.Targets = extractTargetDescription(resp.TargetHealthDescriptions)
	return nil
}

func apifyTargetDescription(target []*svcapitypes.TargetDescription) []*svcsdk.TargetDescription {
	convertedTarget := make([]*svcsdk.TargetDescription, 0, len(target))
	for _, t := range target {
		convertedTarget = append(convertedTarget, &svcsdk.TargetDescription{
			Id:               (*string)(t.ID),
			AvailabilityZone: (*string)(t.AvailabilityZone),
			Port:             (*int64)(t.Port),
		})
	}
	return convertedTarget
}

func extractTargetDescription(targetHealth []*svcsdk.TargetHealthDescription) []*svcapitypes.TargetDescription {
	convertedTarget := make([]*svcapitypes.TargetDescription, 0, len(targetHealth))
	for _, t := range targetHealth {
		convertedTarget = append(convertedTarget, &svcapitypes.TargetDescription{
			ID:               (*string)(t.Target.Id),
			AvailabilityZone: (*string)(t.Target.AvailabilityZone),
			Port:             (*int64)(t.Target.Port),
		})
	}
	return convertedTarget
}
