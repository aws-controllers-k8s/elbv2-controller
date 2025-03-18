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

package tags

import (
	"context"

	acktags "github.com/aws-controllers-k8s/runtime/pkg/tags"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

var (
	_             = svcapitypes.Tag{}
	_             = acktags.NewTags()
	ACKSystemTags = []string{"services.k8s.aws/namespace", "services.k8s.aws/controller-version"}
)

type metricsRecorder interface {
	RecordAPICall(opType string, opID string, err error)
}

type tagsClient interface {
	DescribeTags(ctx context.Context, params *svcsdk.DescribeTagsInput, optFuncs ...func(*svcsdk.Options)) (*svcsdk.DescribeTagsOutput, error)
	AddTags(ctx context.Context, params *svcsdk.AddTagsInput, optFuncs ...func(*svcsdk.Options)) (*svcsdk.AddTagsOutput, error)
	RemoveTags(ctx context.Context, params *svcsdk.RemoveTagsInput, optFuncs ...func(*svcsdk.Options)) (*svcsdk.RemoveTagsOutput, error)
}

// GetRecourceTags uses DescribeTags API Call to get the tags of a resource.
func GetResourceTags(
	ctx context.Context,
	client tagsClient,
	mr metricsRecorder,
	resourceARN string,
) ([]*svcapitypes.Tag, error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("GetRecourceTags")
	defer func() { exit(nil) }()

	describeTagsResponse, err := client.DescribeTags(
		ctx,
		&svcsdk.DescribeTagsInput{
			ResourceArns: []string{resourceARN},
		},
	)
	mr.RecordAPICall("GET", "DescribeTags", err)
	if err != nil {
		return nil, err
	}

	tags := make([]*svcapitypes.Tag, 0, len(describeTagsResponse.TagDescriptions))
	for _, tagDescription := range describeTagsResponse.TagDescriptions {
		for _, tagStruct := range tagDescription.Tags {
			tags = append(tags, &svcapitypes.Tag{
				Key:   tagStruct.Key,
				Value: tagStruct.Value,
			})
		}
	}
	return tags, nil
}

// SyncResourceTags uses TagResource and UntagResource API Calls to add, remove
// and update resource tags.
func SyncRecourseTags(
	ctx context.Context,
	client tagsClient,
	mr metricsRecorder,
	resourceARN string,
	currentTags []*svcapitypes.Tag,
	desiredTags []*svcapitypes.Tag,
	toACKTags func(tags []*svcapitypes.Tag) (acktags.Tags, []string),
	FromACKTags func(tags acktags.Tags) []*svcapitypes.Tag,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("SyncRecourseTags")
	defer func() { exit(err) }()

	from, _ := ToACKTags(currentTags)
	to, _ := ToACKTags(desiredTags)

	toAdd, _, toRemove := ackcompare.GetTagsDifference(from, to)

	addedOrUpdated := FromACKTags(toAdd)
	removeList := FromACKTags(toRemove)

	var removed []string
	for _, tag := range removeList {
		if tag.Key != nil {
			removed = append(removed, *tag.Key)
		}
	}

	if len(removed) > 0 {
		_, err = client.RemoveTags(ctx, &svcsdk.RemoveTagsInput{
			ResourceArns: []string{resourceARN},
			TagKeys:      removed,
		})
		mr.RecordAPICall("UPDATE", "RemoveTags", err)
		if err != nil {
			return err
		}
	}

	if len(addedOrUpdated) > 0 {
		_, err = client.AddTags(ctx, &svcsdk.AddTagsInput{
			ResourceArns: []string{resourceARN},
			Tags:         sdkTagsFromResourceTags(addedOrUpdated),
		})
		mr.RecordAPICall("UPDATE", "AddTags", err)
		if err != nil {
			return err
		}
	}
	return nil
}

// svcTagsFromResourceTags transforms a *svcapitypes.Tag array to a *svcsdk.Tag array.
func sdkTagsFromResourceTags(rTags []*svcapitypes.Tag) []svcsdktypes.Tag {
	tags := make([]svcsdktypes.Tag, len(rTags))
	for i := range rTags {
		tags[i] = svcsdktypes.Tag{
			Key:   rTags[i].Key,
			Value: rTags[i].Value,
		}
	}
	return tags
}
