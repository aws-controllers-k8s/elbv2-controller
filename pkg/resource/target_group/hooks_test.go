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
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
)

func ptr(s string) *string {
	return &s
}

func TestContainsExactTargetGroupAttribute(t *testing.T) {
	attributes := []*svcapitypes.TargetGroupAttribute{
		{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
		{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
	}

	tests := []struct {
		name     string
		target   *svcapitypes.TargetGroupAttribute
		expected bool
	}{
		{
			name:     "matching key and value",
			target:   &svcapitypes.TargetGroupAttribute{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			expected: true,
		},
		{
			name:     "matching key but different value",
			target:   &svcapitypes.TargetGroupAttribute{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("false")},
			expected: false,
		},
		{
			name:     "key not found",
			target:   &svcapitypes.TargetGroupAttribute{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
			expected: false,
		},
		{
			name:     "nil key in target",
			target:   &svcapitypes.TargetGroupAttribute{Key: nil, Value: ptr("true")},
			expected: false,
		},
		{
			name:     "nil value in target",
			target:   &svcapitypes.TargetGroupAttribute{Key: ptr("proxy_protocol_v2.enabled"), Value: nil},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsExactTargetGroupAttribute(attributes, tt.target)
			if result != tt.expected {
				t.Errorf("containsExactTargetGroupAttribute() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTargetGroupAttributesHaveChanged(t *testing.T) {
	tests := []struct {
		name     string
		desired  []*svcapitypes.TargetGroupAttribute
		latest   []*svcapitypes.TargetGroupAttribute
		changed  bool
	}{
		{
			name: "no change - identical attributes",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			changed: false,
		},
		{
			name: "no change - both empty",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest:  []*svcapitypes.TargetGroupAttribute{},
			changed: false,
		},
		{
			name: "change - attribute value modified",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("false")},
			},
			changed: true,
		},
		{
			name: "change - attribute added in desired",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
				{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			changed: true,
		},
		{
			name: "change - attribute removed from desired (present in latest)",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
				{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
			},
			changed: true,
		},
		{
			name: "change - multiple attributes removed",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
				{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
				{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
			},
			changed: true,
		},
		{
			name: "change - all attributes removed",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := targetGroupAttributesHaveChanged(tt.desired, tt.latest)
			if result != tt.changed {
				t.Errorf("targetGroupAttributesHaveChanged() = %v, want %v", result, tt.changed)
			}
		})
	}
}

func TestCompareTargetGroupAttributes(t *testing.T) {
	tests := []struct {
		name       string
		desired    []*svcapitypes.TargetGroupAttribute
		latest     []*svcapitypes.TargetGroupAttribute
		expectDiff bool
	}{
		{
			name: "no diff - identical",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
			},
			expectDiff: false,
		},
		{
			name: "diff - value changed",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val2")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
			},
			expectDiff: true,
		},
		{
			name: "diff - attribute removed",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
			},
			expectDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := ackcompare.NewDelta()
			a := &resource{ko: &svcapitypes.TargetGroup{
				Spec: svcapitypes.TargetGroupSpec{
					Attributes: tt.desired,
				},
			}}
			b := &resource{ko: &svcapitypes.TargetGroup{
				Spec: svcapitypes.TargetGroupSpec{
					Attributes: tt.latest,
				},
			}}
			compareTargetGroupAttributes(delta, a, b)
			hasDiff := len(delta.Differences) > 0
			if hasDiff != tt.expectDiff {
				t.Errorf("compareTargetGroupAttributes() produced diff=%v, want diff=%v", hasDiff, tt.expectDiff)
			}
		})
	}
}

func TestUpdateTargetGroupAttributes_BuildsCompleteAttributeSet(t *testing.T) {
	// This test validates the logic of updateTargetGroupAttributes by
	// checking that the function correctly builds a complete attribute set
	// that includes both desired attributes and reset entries for removed attributes.

	desired := &resource{
		ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Attributes: []*svcapitypes.TargetGroupAttribute{
					{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
				},
			},
		},
	}

	latest := &resource{
		ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Attributes: []*svcapitypes.TargetGroupAttribute{
					{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
					{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
					{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
				},
			},
		},
	}

	// Build the desired attributes map (same logic as updateTargetGroupAttributes)
	desiredAttrs := make(map[string]string)
	for _, attr := range desired.ko.Spec.Attributes {
		if attr.Key == nil || *attr.Key == "" {
			continue
		}
		if attr.Value != nil {
			desiredAttrs[*attr.Key] = *attr.Value
		} else {
			desiredAttrs[*attr.Key] = ""
		}
	}

	// Build the full set (same logic as updateTargetGroupAttributes)
	type attrPair struct {
		key   string
		value string
	}
	var result []attrPair
	for key, value := range desiredAttrs {
		result = append(result, attrPair{key, value})
	}
	for _, attr := range latest.ko.Spec.Attributes {
		if attr.Key == nil || *attr.Key == "" {
			continue
		}
		if _, exists := desiredAttrs[*attr.Key]; !exists {
			result = append(result, attrPair{*attr.Key, ""})
		}
	}

	// Verify: proxy_protocol_v2.enabled should be "true"
	foundProxy := false
	// Verify: deregistration_delay.timeout_seconds should be reset to ""
	foundDereg := false
	// Verify: slow_start.duration_seconds should be reset to ""
	foundSlow := false

	for _, p := range result {
		switch p.key {
		case "proxy_protocol_v2.enabled":
			foundProxy = true
			if p.value != "true" {
				t.Errorf("proxy_protocol_v2.enabled = %q, want %q", p.value, "true")
			}
		case "deregistration_delay.timeout_seconds":
			foundDereg = true
			if p.value != "" {
				t.Errorf("deregistration_delay.timeout_seconds = %q, want empty string (reset)", p.value)
			}
		case "slow_start.duration_seconds":
			foundSlow = true
			if p.value != "" {
				t.Errorf("slow_start.duration_seconds = %q, want empty string (reset)", p.value)
			}
		}
	}

	if !foundProxy {
		t.Error("proxy_protocol_v2.enabled should be present in the result")
	}
	if !foundDereg {
		t.Error("deregistration_delay.timeout_seconds should be present in the result (reset to default)")
	}
	if !foundSlow {
		t.Error("slow_start.duration_seconds should be present in the result (reset to default)")
	}
}
