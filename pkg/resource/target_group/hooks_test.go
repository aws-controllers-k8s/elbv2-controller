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

func int64Ptr(i int64) *int64 {
	return &i
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

func TestTargetGroupAttributesHaveDrifted(t *testing.T) {
	tests := []struct {
		name    string
		desired []*svcapitypes.TargetGroupAttribute
		latest  []*svcapitypes.TargetGroupAttribute
		drifted bool
	}{
		{
			name: "no drift - identical attributes",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			drifted: false,
		},
		{
			name:    "no drift - both empty",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest:  []*svcapitypes.TargetGroupAttribute{},
			drifted: false,
		},
		{
			name: "drift - attribute value modified",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("false")},
			},
			drifted: true,
		},
		{
			name: "drift - desired attribute missing from latest",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
				{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			drifted: true,
		},
		{
			name: "no drift - latest has extra undeclared attributes",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
				{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
				{Key: ptr("slow_start.duration_seconds"), Value: ptr("0")},
			},
			drifted: false,
		},
		{
			name:    "no drift - empty desired leaves AWS state untouched",
			desired: []*svcapitypes.TargetGroupAttribute{},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			},
			drifted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := targetGroupAttributesHaveDrifted(tt.desired, tt.latest)
			if result != tt.drifted {
				t.Errorf("targetGroupAttributesHaveDrifted() = %v, want %v", result, tt.drifted)
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
		{
			name: "no diff - latest has extra undeclared attribute",
			desired: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
			},
			latest: []*svcapitypes.TargetGroupAttribute{
				{Key: ptr("key1"), Value: ptr("val1")},
				{Key: ptr("key2"), Value: ptr("val2")},
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

func TestDeregistrationDelayTimeoutScenario(t *testing.T) {
	t.Run("step1_initial_no_timeout_set", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
		}

		if targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected no drift: user has no desired attributes, AWS defaults should be preserved")
		}

		delta := ackcompare.NewDelta()
		a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: desired}}}
		b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: latest}}}
		compareTargetGroupAttributes(delta, a, b)
		if len(delta.Differences) != 0 {
			t.Error("expected no delta when desired attributes is empty")
		}
	})

	t.Run("step2_set_timeout_to_60", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
		}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
		}

		if !targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected drift: desired=60, latest=300")
		}
	})

	t.Run("step3_modify_timeout_to_120", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
		}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
		}

		if !targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected drift: desired=120, latest=60")
		}
	})

	t.Run("step4_remove_attribute_leaves_aws_untouched", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
		}

		if targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected no drift: removing an attribute from spec leaves the AWS value untouched")
		}
	})
}

func TestMultiAttributeDriftScenario(t *testing.T) {
	t.Run("step1_set_multiple_attributes", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
		}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("false")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("300")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("false")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("round_robin")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("0")},
		}

		if !targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected drift: all 5 attributes differ from AWS defaults")
		}

		delta := ackcompare.NewDelta()
		a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: desired}}}
		b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Attributes: latest}}}
		compareTargetGroupAttributes(delta, a, b)
		if len(delta.Differences) == 0 {
			t.Error("expected delta differences for 5 attribute changes")
		}
	})

	t.Run("step2_modify_subset_of_attributes", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")},
		}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("60")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("30")},
		}

		if !targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected drift: 2 attributes modified (timeout 60->120, slow_start 30->60)")
		}
	})

	t.Run("step3_remove_some_attributes_leaves_aws_untouched", func(t *testing.T) {
		desired := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")},
		}
		latest := []*svcapitypes.TargetGroupAttribute{
			{Key: ptr("proxy_protocol_v2.enabled"), Value: ptr("true")},
			{Key: ptr("deregistration_delay.timeout_seconds"), Value: ptr("120")},
			{Key: ptr("preserve_client_ip.enabled"), Value: ptr("true")},
			{Key: ptr("load_balancing.algorithm.type"), Value: ptr("least_outstanding_requests")},
			{Key: ptr("slow_start.duration_seconds"), Value: ptr("60")},
		}

		if targetGroupAttributesHaveDrifted(desired, latest) {
			t.Error("expected no drift: all declared attributes match; undeclared attributes are left untouched")
		}
	})
}

func TestIsTargetManagementIgnored(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name:        "nil annotations",
			annotations: nil,
			expected:    false,
		},
		{
			name:        "empty annotations",
			annotations: map[string]string{},
			expected:    false,
		},
		{
			name: "annotation set to ignore",
			annotations: map[string]string{
				"elbv2.services.k8s.aws/target-management": "ignore",
			},
			expected: true,
		},
		{
			name: "annotation set to other value",
			annotations: map[string]string{
				"elbv2.services.k8s.aws/target-management": "managed",
			},
			expected: false,
		},
		{
			name: "other annotations present but not target-management",
			annotations: map[string]string{
				"some.other.annotation": "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &svcapitypes.TargetGroup{}
			tg.SetAnnotations(tt.annotations)
			r := &resource{ko: tg}
			result := isTargetManagementIgnored(r)
			if result != tt.expected {
				t.Errorf("isTargetManagementIgnored() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsTargetManagementIgnoredNilResource(t *testing.T) {
	if isTargetManagementIgnored(nil) {
		t.Error("expected false for nil resource")
	}

	r := &resource{ko: nil}
	if isTargetManagementIgnored(r) {
		t.Error("expected false for resource with nil ko")
	}
}

// TestCompareTargetDescriptionWithIgnoredAnnotation verifies that when target
// management is ignored and sdkFind skips describeTargets, both desired and
// latest have the same targets (from the k8s spec), resulting in no delta.
func TestCompareTargetDescriptionWithIgnoredAnnotation(t *testing.T) {
	t.Run("both nil targets - no delta", func(t *testing.T) {
		delta := ackcompare.NewDelta()
		desired := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: nil,
			},
		}}
		latest := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: nil,
			},
		}}
		compareTargetDescription(delta, desired, latest)
		if len(delta.Differences) > 0 {
			t.Error("expected no delta when both desired and latest have nil targets")
		}
	})

	t.Run("same non-empty targets - no delta", func(t *testing.T) {
		delta := ackcompare.NewDelta()
		targets := []*svcapitypes.TargetDescription{
			{ID: ptr("i-12345"), Port: int64Ptr(80)},
		}
		desired := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: targets,
			},
		}}
		latest := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: targets,
			},
		}}
		compareTargetDescription(delta, desired, latest)
		if len(delta.Differences) > 0 {
			t.Error("expected no delta when desired and latest have identical targets")
		}
	})

	t.Run("desired empty, latest has targets - delta (simulates annotation NOT set)", func(t *testing.T) {
		delta := ackcompare.NewDelta()
		desired := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: nil,
			},
		}}
		latest := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: []*svcapitypes.TargetDescription{
					{ID: ptr("i-external")},
				},
			},
		}}
		compareTargetDescription(delta, desired, latest)
		if len(delta.Differences) == 0 {
			t.Error("expected delta when desired is empty and latest has targets (annotation not set)")
		}
	})

	t.Run("desired has targets, latest empty - delta", func(t *testing.T) {
		delta := ackcompare.NewDelta()
		desired := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: []*svcapitypes.TargetDescription{
					{ID: ptr("i-new-target")},
				},
			},
		}}
		latest := &resource{ko: &svcapitypes.TargetGroup{
			Spec: svcapitypes.TargetGroupSpec{
				Targets: nil,
			},
		}}
		compareTargetDescription(delta, desired, latest)
		if len(delta.Differences) == 0 {
			t.Error("expected delta when desired has targets and latest is empty")
		}
	})
}
