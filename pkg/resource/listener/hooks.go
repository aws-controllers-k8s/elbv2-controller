package listener

import (
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
)

const (
	// AnnotationWeightManagement controls how the controller manages forward
	// action weights for target groups. When set to "ignore", the controller
	// will not reconcile weight differences — allowing an external controller
	// or deployment tool to manage blue/green traffic shifting independently.
	AnnotationWeightManagement = "elbv2.services.k8s.aws/weight-management"
)

// customCheckRequiredFieldsMissingMethod returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status.
func (rm *resourceManager) customCheckRequiredFieldsMissingMethod(
	r *resource,
) bool {
	return r.Identifiers().ARN() == nil
}

// customPreCompare is the delta_pre_compare hook. When weight management is
// ignored, it copies the AWS-side weights into the desired spec so that the
// DeepEqual comparison on DefaultActions does not flag external weight
// changes as drift.
func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if isWeightManagementIgnored(a) {
		mergeLatestWeights(a, b)
	}
}

// isWeightManagementIgnored returns true if the resource has the
// AnnotationWeightManagement annotation set to "ignore", indicating that
// target group weights should be managed by an external controller.
func isWeightManagementIgnored(r *resource) bool {
	if r == nil || r.ko == nil {
		return false
	}
	annotations := r.ko.GetAnnotations()
	if annotations == nil {
		return false
	}
	return annotations[AnnotationWeightManagement] == "ignore"
}

// mergeLatestWeights copies the TargetGroup weights from the latest (AWS)
// state into the desired resource. This prevents external weight changes
// from being detected as drift and from being overwritten during updates.
func mergeLatestWeights(desired, latest *resource) {
	if latest == nil || latest.ko == nil || desired == nil || desired.ko == nil {
		return
	}
	// Build a map from TargetGroupARN to Weight from the latest (AWS) state
	latestWeights := map[string]*int64{}
	for _, action := range latest.ko.Spec.DefaultActions {
		if action != nil && action.ForwardConfig != nil {
			for _, tg := range action.ForwardConfig.TargetGroups {
				if tg.TargetGroupARN != nil {
					latestWeights[*tg.TargetGroupARN] = tg.Weight
				}
			}
		}
	}

	// Overwrite desired weights with latest weights for any target group
	// that exists in both desired and latest.
	for _, action := range desired.ko.Spec.DefaultActions {
		if action != nil && action.ForwardConfig != nil {
			for _, tg := range action.ForwardConfig.TargetGroups {
				if tg.TargetGroupARN != nil {
					if w, ok := latestWeights[*tg.TargetGroupARN]; ok {
						tg.Weight = w
					}
				}
			}
		}
	}
}
