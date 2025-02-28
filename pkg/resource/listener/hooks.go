package listener

import (
	"context"
	"time"

	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws-controllers-k8s/elbv2-controller/pkg/resource/tags"
	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
)

var (
	RequeueAfterUpdateDuration = 5 * time.Second
)

// customCheckRequiredFieldsMissingMethod returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status.
func (rm *resourceManager) customCheckRequiredFieldsMissingMethod(
	r *resource,
) bool {
	return r.Identifiers().ARN() == nil
}

func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN  string,
) ([]*svcapitypes.Tag, error) {
	return tags.GetResourceTags(ctx, rm.sdkapi, rm.metrics, resourceARN )
}

func (rm *resourceManager) updateTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.describeTargets")
	defer func() { exit(err) }()
	return tags.SyncRecourseTags(ctx, rm.sdkapi, rm.metrics, string(*desired.ko.Status.ACKResourceMetadata.ARN), desired.ko.Spec.Tags, latest.ko.Spec.Tags)
}