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
	"errors"
	"fmt"
	"reflect"
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
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

func i64(v int64) *int64 {
	return &v
}

func target(id string, port *int64, az *string) *svcapitypes.TargetDescription {
	return &svcapitypes.TargetDescription{ID: ptr(id), Port: port, AvailabilityZone: az}
}

func targetIDs(targets []*svcapitypes.TargetDescription) []string {
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		switch {
		case t == nil:
			out = append(out, "<nil>")
		case t.ID == nil:
			out = append(out, "<nil-id>")
		case t.Port == nil:
			out = append(out, *t.ID)
		default:
			out = append(out, fmt.Sprintf("%s:%d", *t.ID, *t.Port))
		}
	}
	return out
}

func TestGetTargetsDifference(t *testing.T) {
	tests := []struct {
		name    string
		latest  []*svcapitypes.TargetDescription
		desired []*svcapitypes.TargetDescription
		added   []string
		removed []string
	}{
		{
			name:    "same instance on two ports is unchanged",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
			added:   []string{},
			removed: []string{},
		},
		{
			name:    "same instance on two ports in a different order is unchanged",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(443), nil), target("i-1", i64(80), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
			added:   []string{},
			removed: []string{},
		},
		{
			name:    "same ip on three ports with aws supplied availability zone is unchanged",
			latest:  []*svcapitypes.TargetDescription{target("10.0.0.1", i64(80), ptr("us-west-2a")), target("10.0.0.1", i64(443), ptr("us-west-2a")), target("10.0.0.1", i64(8080), ptr("us-west-2a"))},
			desired: []*svcapitypes.TargetDescription{target("10.0.0.1", i64(80), nil), target("10.0.0.1", i64(443), nil), target("10.0.0.1", i64(8080), nil)},
			added:   []string{},
			removed: []string{},
		},
		{
			name:    "omitted port matches the port aws reports",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), ptr("us-west-2a"))},
			desired: []*svcapitypes.TargetDescription{target("i-1", nil, nil)},
			added:   []string{},
			removed: []string{},
		},
		{
			name:    "one of two ports removed",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil)},
			added:   []string{},
			removed: []string{"i-1:443"},
		},
		{
			name:    "one more port added to an existing instance",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
			added:   []string{"i-1:443"},
			removed: []string{},
		},
		{
			name:    "port changed on the only target",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(8080), nil)},
			added:   []string{"i-1:8080"},
			removed: []string{"i-1:80"},
		},
		{
			name:    "availability zone changed",
			latest:  []*svcapitypes.TargetDescription{target("10.0.0.1", i64(80), ptr("us-west-2a"))},
			desired: []*svcapitypes.TargetDescription{target("10.0.0.1", i64(80), ptr("us-west-2b"))},
			added:   []string{"10.0.0.1:80"},
			removed: []string{"10.0.0.1:80"},
		},
		{
			name:    "different instance replaces the old one",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-2", i64(80), nil)},
			added:   []string{"i-2:80"},
			removed: []string{"i-1:80"},
		},
		{
			name:    "exact identity is paired before an omitted port claims it",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", nil, nil), target("i-1", i64(443), nil)},
			added:   []string{},
			removed: []string{},
		},
		{
			name:    "nil id in desired does not panic",
			latest:  []*svcapitypes.TargetDescription{target("i-1", i64(80), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), {Port: i64(443)}},
			added:   []string{"<nil-id>"},
			removed: []string{},
		},
		{
			name:    "nil id in latest does not panic",
			latest:  []*svcapitypes.TargetDescription{{Port: i64(443)}, target("i-1", i64(80), nil)},
			desired: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil)},
			added:   []string{},
			removed: []string{"<nil-id>"},
		},
		{
			name:    "nil entries in both lists do not panic",
			latest:  []*svcapitypes.TargetDescription{nil},
			desired: []*svcapitypes.TargetDescription{nil},
			added:   []string{"<nil>"},
			removed: []string{"<nil>"},
		},
		{
			name:    "empty on both sides",
			latest:  []*svcapitypes.TargetDescription{},
			desired: []*svcapitypes.TargetDescription{},
			added:   []string{},
			removed: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed := getTargetsDifference(tt.latest, tt.desired)
			if got := targetIDs(added); !reflect.DeepEqual(got, tt.added) {
				t.Errorf("added = %v, want %v", got, tt.added)
			}
			if got := targetIDs(removed); !reflect.DeepEqual(got, tt.removed) {
				t.Errorf("removed = %v, want %v", got, tt.removed)
			}
		})
	}
}

func TestCompareTargetDescriptionMultiPortNoDelta(t *testing.T) {
	desired := []*svcapitypes.TargetDescription{
		target("i-1", i64(80), nil),
		target("i-1", i64(443), nil),
		target("i-2", i64(80), nil),
	}
	latest := []*svcapitypes.TargetDescription{
		target("i-1", i64(80), ptr("us-west-2a")),
		target("i-1", i64(443), ptr("us-west-2a")),
		target("i-2", i64(80), ptr("us-west-2b")),
	}

	delta := ackcompare.NewDelta()
	a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Targets: desired}}}
	b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Targets: latest}}}
	compareTargetDescription(delta, a, b)

	if len(delta.Differences) != 0 {
		t.Errorf("expected no delta for the same instance registered on several ports, got %d differences", len(delta.Differences))
	}
}

func TestCompareTargetDescriptionDoesNotMutateDesired(t *testing.T) {
	desired := []*svcapitypes.TargetDescription{
		target("i-1", nil, nil),
		target("i-2", i64(80), nil),
	}
	latest := []*svcapitypes.TargetDescription{
		target("i-1", i64(80), ptr("us-west-2a")),
		target("i-2", i64(80), ptr("us-west-2b")),
	}
	desiredBefore := deepCopyTargets(desired)
	latestBefore := deepCopyTargets(latest)

	delta := ackcompare.NewDelta()
	a := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Targets: desired}}}
	b := &resource{ko: &svcapitypes.TargetGroup{Spec: svcapitypes.TargetGroupSpec{Targets: latest}}}
	compareTargetDescription(delta, a, b)

	if !reflect.DeepEqual(desired, desiredBefore) {
		t.Errorf("compareTargetDescription mutated desired: got %+v, want %+v", targetIDs(desired), targetIDs(desiredBefore))
	}
	if !reflect.DeepEqual(latest, latestBefore) {
		t.Errorf("compareTargetDescription mutated latest: got %+v, want %+v", targetIDs(latest), targetIDs(latestBefore))
	}
}

func TestAreDifferentTargetDoesNotMutate(t *testing.T) {
	desired := target("i-1", nil, nil)
	latest := target("i-1", i64(80), ptr("us-west-2a"))

	if !areDifferentTarget(latest, desired) {
		t.Error("expected targets with different ports to be reported as different")
	}
	if desired.Port != nil {
		t.Errorf("areDifferentTarget wrote Port %d back into desired", *desired.Port)
	}
	if desired.AvailabilityZone != nil {
		t.Errorf("areDifferentTarget wrote AvailabilityZone %q back into desired", *desired.AvailabilityZone)
	}
}

func TestAreDifferentTargetNilHandling(t *testing.T) {
	if areDifferentTarget(nil, nil) {
		t.Error("expected two nil targets to be equal")
	}
	if !areDifferentTarget(nil, target("i-1", nil, nil)) {
		t.Error("expected a nil latest to differ from a non-nil desired")
	}
	if !areDifferentTarget(target("i-1", nil, nil), nil) {
		t.Error("expected a non-nil latest to differ from a nil desired")
	}
}

func TestValidateTargets(t *testing.T) {
	tests := []struct {
		name     string
		targets  []*svcapitypes.TargetDescription
		wantErr  bool
		terminal bool
	}{
		{
			name:    "all targets have an id",
			targets: []*svcapitypes.TargetDescription{target("i-1", i64(80), nil), target("i-1", i64(443), nil)},
		},
		{
			name:    "no targets",
			targets: []*svcapitypes.TargetDescription{},
		},
		{
			name:     "nil id is terminal",
			targets:  []*svcapitypes.TargetDescription{{Port: i64(80)}},
			wantErr:  true,
			terminal: true,
		},
		{
			name:     "empty id is terminal",
			targets:  []*svcapitypes.TargetDescription{target("", i64(80), nil)},
			wantErr:  true,
			terminal: true,
		},
		{
			name:     "nil target is terminal",
			targets:  []*svcapitypes.TargetDescription{nil},
			wantErr:  true,
			terminal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTargets(tt.targets)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTargets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.terminal {
				var terminal *ackerr.TerminalError
				if !errors.As(err, &terminal) {
					t.Errorf("validateTargets() error = %v, want a terminal error", err)
				}
			}
		})
	}
}

func deepCopyTargets(targets []*svcapitypes.TargetDescription) []*svcapitypes.TargetDescription {
	out := make([]*svcapitypes.TargetDescription, 0, len(targets))
	for _, t := range targets {
		if t == nil {
			out = append(out, nil)
			continue
		}
		out = append(out, t.DeepCopy())
	}
	return out
}
