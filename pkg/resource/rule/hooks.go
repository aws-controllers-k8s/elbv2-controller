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

package rule

import (
	"context"
	"errors"
	"strconv"

	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

var (
	// ErrInvalidPriority is an error that is returned when the priority value is invalid.
	ErrInvalidPriority = errors.New("invalid priority value")
)

// setRulePriority sets the priority of the rule.
func (rm *resourceManager) setRulePriority(
	ctx context.Context,
	res *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateLoadBalancerAttributes")
	defer func() { exit(err) }()

	input := &svcsdk.SetRulePrioritiesInput{
		RulePriorities: []svcsdktypes.RulePriorityPair{
			{
				Priority: aws.Int32(int32(*res.ko.Spec.Priority)),
				RuleArn:  (*string)(res.ko.Status.ACKResourceMetadata.ARN),
			},
		},
	}
	_, err = rm.sdkapi.SetRulePriorities(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateRule", err)
	if err != nil {
		return err
	}

	return nil
}

// customCheckRequiredFieldsMissingMethod returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status.
func (rm *resourceManager) customCheckRequiredFieldsMissingMethod(
	r *resource,
) bool {
	return r.Identifiers().ARN() == nil
}

// priorityFromSDK converts the priority from the SDK type to API type.
//
// Yes, the API takes a pointer to int64, but the SDK returns a pointer to string...
func priorityFromSDK(sdkPriority *string) *int64 {
	// Since this function is only used in the context of the SDK, we can safely
	// assume that the SDK will never return a nil pointer nor a invalid value.
	priority, _ := strconv.Atoi(*sdkPriority)
	priorityInt64 := int64(priority)
	return &priorityInt64
}
