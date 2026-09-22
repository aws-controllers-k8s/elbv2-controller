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
	"fmt"
	"strconv"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"k8s.io/apimachinery/pkg/api/equality"
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

	if res.ko.Status.ACKResourceMetadata == nil || res.ko.Status.ACKResourceMetadata.ARN == nil {
		return fmt.Errorf("rule ARN is not yet available")
	}
	input := &svcsdk.SetRulePrioritiesInput{
		RulePriorities: []svcsdktypes.RulePriorityPair{
			{
				Priority: int32OrNil(res.ko.Spec.Priority),
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
	if sdkPriority == nil {
		return nil
	}
	// The default rule reports a priority of "default", which Atoi maps to 0.
	priority, _ := strconv.Atoi(*sdkPriority)
	priorityInt64 := int64(priority)
	return &priorityInt64
}

func int32OrNil(val *int64) *int32 {
	if val != nil {
		return aws.Int32(int32(*val))
	}
	return nil
}

func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	customCompareConditions(delta, a, b)
}

// customCompareConditions performs custom comparison for Rule conditions.
// AWS ELBv2 API returns both the generic 'values' field and condition-specific
// config fields (e.g., hostHeaderConfig.values) for host-header and path-pattern
// conditions. We only compare the fields that were specified in the desired state.
func customCompareConditions(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if a == nil || b == nil {
		return
	}

	if len(a.ko.Spec.Conditions) != len(b.ko.Spec.Conditions) {
		delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
		return
	}

	matchedObserved := make([]bool, len(b.ko.Spec.Conditions))

	for _, desiredCond := range a.ko.Spec.Conditions {
		i := matchConditionIndex(b.ko.Spec.Conditions, matchedObserved, desiredCond)
		if i < 0 {
			delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
			return
		}
		matchedObserved[i] = true
		observedCond := b.ko.Spec.Conditions[i]

		// For host-header and path-pattern conditions, compare based on what's in desired
		if desiredCond.Field != nil {
			switch *desiredCond.Field {
			case "host-header":
				if desiredCond.HostHeaderConfig != nil {
					if !equality.Semantic.Equalities.DeepEqual(desiredCond.HostHeaderConfig, observedCond.HostHeaderConfig) {
						delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
						return
					}
				}
				if desiredCond.Values != nil {
					if !equality.Semantic.Equalities.DeepEqual(desiredCond.Values, observedCond.Values) {
						delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
						return
					}
				}
			case "path-pattern":
				if desiredCond.PathPatternConfig != nil {
					if !equality.Semantic.Equalities.DeepEqual(desiredCond.PathPatternConfig, observedCond.PathPatternConfig) {
						delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
						return
					}
				}
				if desiredCond.Values != nil {
					if !equality.Semantic.Equalities.DeepEqual(desiredCond.Values, observedCond.Values) {
						delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
						return
					}
				}
			default:
				if !equality.Semantic.Equalities.DeepEqual(desiredCond, observedCond) {
					delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
					return
				}
			}
		}
	}
}

// matchConditionIndex returns the index of an as-yet unmatched observed
// condition that corresponds to desired, or -1. ELBv2 permits several
// http-header conditions on one rule, so conditions are paired on their header
// name where they have one and each observed condition is claimed at most once;
// pairing on field alone collapses them all onto the first observed condition.
func matchConditionIndex(
	observed []*svcapitypes.RuleCondition,
	matched []bool,
	desired *svcapitypes.RuleCondition,
) int {
	if desired == nil || desired.Field == nil {
		return -1
	}
	fallback := -1
	for i, oc := range observed {
		if matched[i] || oc == nil || oc.Field == nil || *oc.Field != *desired.Field {
			continue
		}
		if httpHeaderName(oc) == httpHeaderName(desired) {
			return i
		}
		if fallback < 0 {
			fallback = i
		}
	}
	return fallback
}

func httpHeaderName(c *svcapitypes.RuleCondition) string {
	if c.HTTPHeaderConfig == nil || c.HTTPHeaderConfig.HTTPHeaderName == nil {
		return ""
	}
	return *c.HTTPHeaderConfig.HTTPHeaderName
}
