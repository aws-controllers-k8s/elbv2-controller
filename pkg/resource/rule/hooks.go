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
	// Since this function is only used in the context of the SDK, we can safely
	// assume that the SDK will never return a nil pointer nor a invalid value.
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
	customCompareActions(delta, a, b)
}

// customCompareActions performs custom comparison for Rule actions.
// Actions is compared manually (compare.is_ignored: true) because the AWS
// ELBv2 API never returns the k8s-only TargetGroupRef fields and assigns an
// Order when the user omits one. Comparing the auto-generated way would always
// report a diff (the ref present in desired but absent in the observed state),
// triggering a redundant ModifyRule on every reconcile. We strip the ref fields
// and normalize the server-assigned Order before comparing; the resolved
// TargetGroupARN is still compared so genuine drift is detected.
func customCompareActions(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if a == nil || b == nil {
		return
	}

	desired := a.ko.Spec.Actions
	observed := b.ko.Spec.Actions
	if len(desired) != len(observed) {
		delta.Add("Spec.Actions", desired, observed)
		return
	}

	for i := range desired {
		d := normalizeActionForCompare(desired[i])
		o := normalizeActionForCompare(observed[i])
		// AWS assigns an Order when the user does not specify one, so only
		// compare Order when it was set in the desired state.
		if desired[i] != nil && desired[i].Order == nil {
			o.Order = nil
		}
		if !equality.Semantic.Equalities.DeepEqual(d, o) {
			delta.Add("Spec.Actions", desired, observed)
			return
		}
	}
}

// normalizeActionForCompare returns a deep copy of the action with the k8s-only
// TargetGroupRef fields removed, both at the action level and within
// ForwardConfig.TargetGroups, so they do not produce spurious diffs.
func normalizeActionForCompare(action *svcapitypes.Action) *svcapitypes.Action {
	if action == nil {
		return nil
	}
	a := action.DeepCopy()
	a.TargetGroupRef = nil
	if a.ForwardConfig != nil {
		for j := range a.ForwardConfig.TargetGroups {
			if a.ForwardConfig.TargetGroups[j] != nil {
				a.ForwardConfig.TargetGroups[j].TargetGroupRef = nil
			}
		}
	}
	return a
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

	for _, desiredCond := range a.ko.Spec.Conditions {
		var observedCond *svcapitypes.RuleCondition
		if desiredCond.Field != nil {
			for _, oc := range b.ko.Spec.Conditions {
				if oc.Field != nil && *oc.Field == *desiredCond.Field {
					observedCond = oc
					break
				}
			}
		}

		if observedCond == nil {
			delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
			return
		}

		if (desiredCond.Field == nil && observedCond.Field != nil) ||
			(desiredCond.Field != nil && observedCond.Field == nil) ||
			(desiredCond.Field != nil && observedCond.Field != nil && *desiredCond.Field != *observedCond.Field) {
			delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
			return
		}

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
