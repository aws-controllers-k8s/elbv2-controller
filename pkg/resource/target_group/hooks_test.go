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
		{Key: ptr("stickiness.enabled"), Value: nil},
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
			name:     "nil value in target (base has non-nil value)",
			target:   &svcapitypes.TargetGroupAttribute{Key: ptr("proxy_protocol_v2.enabled"), Value: nil},
			expected: false,
		},
		{
			name:     "both values nil with matching key",
			target:   &svcapitypes.TargetGroupAttribute{Key: ptr("stickiness.enabled"), Value: nil},
			expected: true,
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
		name    string
		desired []*svcapitypes.TargetGroupAttribute
		latest  []*svcapitypes.TargetGroupAttribute
		changed bool
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
			name:    "no change - both empty",
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
			name:    "no change - empty desired attributes",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			changed: false,
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
			name:    "no diff - empty desired attributes",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
			},
			expectDiff: false,
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

func TestCustomBuildAttributesForUpdate(t *testing.T) {
	desired := []*svcapitypes.TargetGroupAttribute{
		{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
	}
	latest := []*svcapitypes.TargetGroupAttribute{
		{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
		{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
		{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
	}

	result := customBuildAttributesForUpdate(desired, latest)

	foundProxy := false
	foundDereg := false
	foundSlow := false

	for _, attr := range result {
		if attr.Key == nil {
			continue
		}
		switch *attr.Key {
		case "proxy_protocol_v2.enabled":
			foundProxy = true
			if attr.Value == nil || *attr.Value != "true" {
				t.Errorf("proxy_protocol_v2.enabled = %v, want %q", attr.Value, "true")
			}
		case "deregistration_delay.timeout_seconds":
			foundDereg = true
			if attr.Value == nil || *attr.Value != "" {
				t.Errorf("deregistration_delay.timeout_seconds = %v, want empty string (reset)", attr.Value)
			}
		case "slow_start.duration_seconds":
			foundSlow = true
			if attr.Value == nil || *attr.Value != "" {
				t.Errorf("slow_start.duration_seconds = %v, want empty string (reset)", attr.Value)
			}
		}
	}

	if !foundProxy {
		t.Error("proxy_protocol_v2.enabled should be present")
	}
	if !foundDereg {
		t.Error("deregistration_delay.timeout_seconds should be present (reset to default)")
	}
	if !foundSlow {
		t.Error("slow_start.duration_seconds should be present (reset to default)")
	}
}

// TestDeregistrationDelayTimeoutScenario simulates the full lifecycle of
// deregistration_delay.timeout_seconds attribute:
//  1. Initial: no timeout set (default 300s on AWS side)
//  2. User sets deregistration_delay.timeout_seconds = "60"
//  3. User modifies it to "120"
//  4. User removes it (should reset to default)
func TestDeregistrationDelayTimeoutScenario(t *testing.T) {
	t.Run("step1_initial_no_timeout_set", func(t *testing.T) {
		// User spec has no attributes
		desired := []*svcapitypes.TargetGroupAttribute{}
		// AWS returns default value
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
		}

		// No change - user hasn't specified any attributes, so AWS defaults
		// should be left untouched
		changed := targetGroupAttributesHaveChanged(desired, latest)
		if changed {
			t.Error("expected no change: user has no desired attributes, AWS defaults should be preserved")
		}

		// Verify no delta is produced
		delta := ackcompare.NewDelta()
		a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: desired}}}
		b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: latest}}}
		compareTargetGroupAttributes(delta, a, b)
		if len(delta.Differences) != 0 {
			t.Error("expected no delta when desired attributes is empty")
		}
	})

	t.Run("step2_set_timeout_to_60", func(t *testing.T) {
		// User sets deregistration_delay.timeout_seconds = "60"
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
		}
		// AWS still has default 300s
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
		}

		changed := targetGroupAttributesHaveChanged(desired, latest)
		if !changed {
			t.Error("expected change: desired=60, latest=300")
		}

		// Verify the update would send "60"
		result := customBuildAttributesForUpdate(desired, latest)
		found := false
		for _, attr := range result {
			if attr.Key != nil && *attr.Key == "deregistration_delay.timeout_seconds" {
				found = true
				if attr.Value == nil || *attr.Value != "60" {
					t.Errorf("expected deregistration_delay.timeout_seconds=60, got=%v", attr.Value)
				}
			}
		}
		if !found {
			t.Error("deregistration_delay.timeout_seconds not found in result")
		}
	})

	t.Run("step3_modify_timeout_to_120", func(t *testing.T) {
		// User modifies deregistration_delay.timeout_seconds from "60" to "120"
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
		}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
		}

		changed := targetGroupAttributesHaveChanged(desired, latest)
		if !changed {
			t.Error("expected change: desired=120, latest=60")
		}
	})

	t.Run("step4_remove_all_attributes_no_reset", func(t *testing.T) {
		// User removes all attributes from spec
		desired := []*svcapitypes.TargetGroupAttribute{}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
		}

		// When the user clears all desired attributes, no change is triggered.
		// To reset an attribute, the user should explicitly set it to its
		// desired value rather than removing it.
		changed := targetGroupAttributesHaveChanged(desired, latest)
		if changed {
			t.Error("expected no change: empty desired attributes means user hasn't opted into attribute management")
		}
	})
}

// TestMultiAttributeLifecycleScenario simulates the full lifecycle of
// multiple TargetGroup attributes:
//
//	proxy_protocol_v2.enabled - boolean flag
//	deregistration_delay.timeout_seconds - timeout value
//	preserve_client_ip.enabled - boolean flag
//	load_balancing.algorithm.type - algorithm selection
//	slow_start.duration_seconds - duration value
//
// This test verifies that multiple attributes can be set, modified,
// and removed together in a realistic controller reconciliation scenario.
func TestMultiAttributeLifecycleScenario(t *testing.T) {
	t.Run("step1_set_multiple_attributes", func(t *testing.T) {
		// User configures multiple attributes
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
		}
		// AWS currently has different values (or defaults)
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("false")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("false")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("round_robin")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("0")},
		}

		// All 5 attributes differ → should detect change
		changed := targetGroupAttributesHaveChanged(desired, latest)
		if !changed {
			t.Error("expected change: all 5 attributes differ from AWS defaults")
		}

		// Verify delta is produced
		delta := ackcompare.NewDelta()
		a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: desired}}}
		b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: latest}}}
		compareTargetGroupAttributes(delta, a, b)
		if len(delta.Differences) == 0 {
			t.Error("expected delta differences for 5 attribute changes")
		}
	})

	t.Run("step2_modify_subset_of_attributes", func(t *testing.T) {
		// User modifies only 2 of the 5 attributes
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")}, // changed from 60 → 120
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")}, // changed from 30 → 60
		}
		// AWS reflects previous desired state
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
		}

		changed := targetGroupAttributesHaveChanged(desired, latest)
		if !changed {
			t.Error("expected change: 2 attributes modified (timeout 60→120, slow_start 30→60)")
		}

		// Verify only the 2 changed attributes are detected
		delta := ackcompare.NewDelta()
		a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: desired}}}
		b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: latest}}}
		compareTargetGroupAttributes(delta, a, b)
		if len(delta.Differences) == 0 {
			t.Error("expected delta differences for 2 attribute modifications")
		}
	})

	t.Run("step3_remove_some_attributes_keep_others", func(t *testing.T) {
		// User removes 2 attributes, keeps 3
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")},
		}
		// AWS still has all 5 from previous state
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")},
		}

		changed := targetGroupAttributesHaveChanged(desired, latest)
		if !changed {
			t.Error("expected change: 2 attributes removed (preserve_client_ip, load_balancing.algorithm)")
		}

		result := customBuildAttributesForUpdate(desired, latest)

		// Verify removed attributes get reset entries
		foundPreserveClientIP := false
		foundLBAlgorithm := false
		for _, attr := range result {
			if attr.Key == nil {
				continue
			}
			switch *attr.Key {
			case "preserve_client_ip.enabled":
				foundPreserveClientIP = true
				if attr.Value == nil || *attr.Value != "" {
					t.Errorf("preserve_client_ip.enabled should be empty string for reset, got=%v", attr.Value)
				}
			case "load_balancing.algorithm.type":
				foundLBAlgorithm = true
				if attr.Value == nil || *attr.Value != "" {
					t.Errorf("load_balancing.algorithm.type should be empty string for reset, got=%v", attr.Value)
				}
			}
		}
		if !foundPreserveClientIP {
			t.Error("preserve_client_ip.enabled should be present with empty value for reset")
		}
		if !foundLBAlgorithm {
			t.Error("load_balancing.algorithm.type should be present with empty value for reset")
		}

		// Verify kept attributes still have their values
		for _, attr := range result {
			if attr.Key == nil {
				continue
			}
			switch *attr.Key {
			case "proxy_protocol_v2.enabled":
				if attr.Value == nil || *attr.Value != "true" {
					t.Errorf("proxy_protocol_v2.enabled should be 'true', got=%v", attr.Value)
				}
			case "deregistration_delay.timeout_seconds":
				if attr.Value == nil || *attr.Value != "120" {
					t.Errorf("deregistration_delay.timeout_seconds should be '120', got=%v", attr.Value)
				}
			case "slow_start.duration_seconds":
				if attr.Value == nil || *attr.Value != "60" {
					t.Errorf("slow_start.duration_seconds should be '60', got=%v", attr.Value)
				}
			}
		}
	})

	t.Run("step4_remove_all_attributes", func(t *testing.T) {
		// User removes ALL attributes from spec
		desired := []*svcapitypes.TargetGroupAttribute{}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")},
		}

		changed := targetGroupAttributesHaveChanged(desired, latest)
		if changed {
			t.Error("expected no change: empty desired means attributes are not managed")
		}

	})
}
