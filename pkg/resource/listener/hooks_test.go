package listener

import (
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
)

func ptr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestIsWeightManagementIgnored(t *testing.T) {
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
				"elbv2.services.k8s.aws/weight-management": "ignore",
			},
			expected: true,
		},
		{
			name: "annotation set to other value",
			annotations: map[string]string{
				"elbv2.services.k8s.aws/weight-management": "managed",
			},
			expected: false,
		},
		{
			name: "other annotations present but not weight-management",
			annotations: map[string]string{
				"some.other.annotation": "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &svcapitypes.Listener{}
			l.SetAnnotations(tt.annotations)
			r := &resource{ko: l}
			result := isWeightManagementIgnored(r)
			if result != tt.expected {
				t.Errorf("isWeightManagementIgnored() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsWeightManagementIgnoredNilResource(t *testing.T) {
	if isWeightManagementIgnored(nil) {
		t.Error("expected false for nil resource")
	}

	r := &resource{ko: nil}
	if isWeightManagementIgnored(r) {
		t.Error("expected false for resource with nil ko")
	}
}

func TestMergeLatestWeights(t *testing.T) {
	t.Run("merges weights for matching TGs", func(t *testing.T) {
		desired := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(100)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(0)},
							},
						},
					},
				},
			},
		}}
		latest := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(70)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(30)},
							},
						},
					},
				},
			},
		}}

		mergeLatestWeights(desired, latest)

		blueWeight := *desired.ko.Spec.DefaultActions[0].ForwardConfig.TargetGroups[0].Weight
		greenWeight := *desired.ko.Spec.DefaultActions[0].ForwardConfig.TargetGroups[1].Weight

		if blueWeight != 70 {
			t.Errorf("expected blue weight 70, got %d", blueWeight)
		}
		if greenWeight != 30 {
			t.Errorf("expected green weight 30, got %d", greenWeight)
		}
	})

	t.Run("does not change weights for TGs not in latest", func(t *testing.T) {
		desired := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(100)},
								{TargetGroupARN: ptr("arn:aws:tg:new-green"), Weight: int64Ptr(0)},
							},
						},
					},
				},
			},
		}}
		latest := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(70)},
							},
						},
					},
				},
			},
		}}

		mergeLatestWeights(desired, latest)

		blueWeight := *desired.ko.Spec.DefaultActions[0].ForwardConfig.TargetGroups[0].Weight
		newGreenWeight := *desired.ko.Spec.DefaultActions[0].ForwardConfig.TargetGroups[1].Weight

		if blueWeight != 70 {
			t.Errorf("expected blue weight 70 (merged from latest), got %d", blueWeight)
		}
		if newGreenWeight != 0 {
			t.Errorf("expected new-green weight 0 (unchanged, not in latest), got %d", newGreenWeight)
		}
	})

	t.Run("handles nil latest gracefully", func(t *testing.T) {
		desired := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(100)},
							},
						},
					},
				},
			},
		}}

		// Should not panic
		mergeLatestWeights(desired, nil)
		mergeLatestWeights(nil, nil)

		weight := *desired.ko.Spec.DefaultActions[0].ForwardConfig.TargetGroups[0].Weight
		if weight != 100 {
			t.Errorf("expected weight unchanged (100), got %d", weight)
		}
	})
}

func TestCustomPreCompareWeightManagement(t *testing.T) {
	t.Run("with annotation: copies weights, no delta on weight-only change", func(t *testing.T) {
		desired := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(100)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(0)},
							},
						},
					},
				},
			},
		}}
		desired.ko.SetAnnotations(map[string]string{
			"elbv2.services.k8s.aws/weight-management": "ignore",
		})

		latest := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(70)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(30)},
							},
						},
					},
				},
			},
		}}

		delta := newResourceDelta(desired, latest)

		// Weights should have been normalized — no delta on DefaultActions
		if delta.DifferentAt("Spec.DefaultActions") {
			t.Error("expected no delta on DefaultActions when weight-management is ignored")
		}
	})

	t.Run("without annotation: detects weight differences as delta", func(t *testing.T) {
		desired := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(100)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(0)},
							},
						},
					},
				},
			},
		}}

		latest := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(70)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(30)},
							},
						},
					},
				},
			},
		}}

		delta := newResourceDelta(desired, latest)

		// Without annotation, weight differences should be detected
		if !delta.DifferentAt("Spec.DefaultActions") {
			t.Error("expected delta on DefaultActions when weight-management is NOT set")
		}
	})

	t.Run("with annotation: adding a new TG still creates delta", func(t *testing.T) {
		desired := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(100)},
								{TargetGroupARN: ptr("arn:aws:tg:green"), Weight: int64Ptr(0)},
							},
						},
					},
				},
			},
		}}
		desired.ko.SetAnnotations(map[string]string{
			"elbv2.services.k8s.aws/weight-management": "ignore",
		})

		latest := &resource{ko: &svcapitypes.Listener{
			Spec: svcapitypes.ListenerSpec{
				DefaultActions: []*svcapitypes.Action{
					{
						Type: ptr("forward"),
						ForwardConfig: &svcapitypes.ForwardActionConfig{
							TargetGroups: []*svcapitypes.TargetGroupTuple{
								{TargetGroupARN: ptr("arn:aws:tg:blue"), Weight: int64Ptr(70)},
								// green is missing from latest
							},
						},
					},
				},
			},
		}}

		delta := newResourceDelta(desired, latest)

		// Different number of TGs should still create a delta
		if !delta.DifferentAt("Spec.DefaultActions") {
			t.Error("expected delta when desired has an extra TG (even with weight-management ignored)")
		}
	})
}
