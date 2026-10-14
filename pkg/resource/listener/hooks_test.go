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

package listener

import (
	"testing"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
)

func tgRef(name string) *ackv1alpha1.AWSResourceReferenceWrapper {
	return &ackv1alpha1.AWSResourceReferenceWrapper{
		From: &ackv1alpha1.AWSResourceReference{Name: aws.String(name)},
	}
}

// forwardAsReturned is the shape DescribeListeners returns for a forward
// action that was created with a single target group.
func forwardAsReturned(order int64, arn string) *svcapitypes.Action {
	return &svcapitypes.Action{
		Type:           aws.String("forward"),
		Order:          aws.Int64(order),
		TargetGroupARN: aws.String(arn),
		ForwardConfig: &svcapitypes.ForwardActionConfig{
			TargetGroups: []*svcapitypes.TargetGroupTuple{
				{TargetGroupARN: aws.String(arn), Weight: aws.Int64(1)},
			},
			TargetGroupStickinessConfig: &svcapitypes.TargetGroupStickinessConfig{
				Enabled: aws.Bool(false),
			},
		},
	}
}

func weightedForward(order *int64, weightA, weightB int64) *svcapitypes.Action {
	return &svcapitypes.Action{
		Type:  aws.String("forward"),
		Order: order,
		ForwardConfig: &svcapitypes.ForwardActionConfig{
			TargetGroups: []*svcapitypes.TargetGroupTuple{
				{TargetGroupARN: aws.String("arn:tg/a"), Weight: aws.Int64(weightA)},
				{TargetGroupARN: aws.String("arn:tg/b"), Weight: aws.Int64(weightB)},
			},
			TargetGroupStickinessConfig: &svcapitypes.TargetGroupStickinessConfig{
				Enabled:         aws.Bool(true),
				DurationSeconds: aws.Int64(3600),
			},
		},
	}
}

func redirect(path *string) *svcapitypes.RedirectActionConfig {
	return &svcapitypes.RedirectActionConfig{
		Protocol:   aws.String("HTTPS"),
		Port:       aws.String("#{port}"),
		Host:       aws.String("#{host}"),
		Path:       path,
		Query:      aws.String("#{query}"),
		StatusCode: aws.String("HTTP_301"),
	}
}

func oidc(secret *string, defaults bool) *svcapitypes.AuthenticateOIDCActionConfig {
	cfg := &svcapitypes.AuthenticateOIDCActionConfig{
		Issuer:                aws.String("https://idp.example.com"),
		AuthorizationEndpoint: aws.String("https://idp.example.com/auth"),
		TokenEndpoint:         aws.String("https://idp.example.com/token"),
		UserInfoEndpoint:      aws.String("https://idp.example.com/userinfo"),
		ClientID:              aws.String("client-id"),
		ClientSecret:          secret,
	}
	if defaults {
		cfg.OnUnauthenticatedRequest = aws.String("authenticate")
		cfg.Scope = aws.String("openid")
		cfg.SessionCookieName = aws.String("AWSELBAuthSessionCookie")
		cfg.SessionTimeout = aws.Int64(604800)
	}
	return cfg
}

func TestNewResourceDelta_DefaultActions(t *testing.T) {
	tests := []struct {
		name        string
		desired     []*svcapitypes.Action
		observed    []*svcapitypes.Action
		expectDelta bool
	}{
		{
			name: "targetGroupARN only, service adds order and forward config",
			desired: []*svcapitypes.Action{
				{Type: aws.String("forward"), TargetGroupARN: aws.String("arn:tg/a")},
			},
			observed:    []*svcapitypes.Action{forwardAsReturned(1, "arn:tg/a")},
			expectDelta: false,
		},
		{
			name: "targetGroupRef only, resolved ARN matches",
			desired: []*svcapitypes.Action{
				{
					Type:           aws.String("forward"),
					TargetGroupRef: tgRef("my-tg"),
					TargetGroupARN: aws.String("arn:tg/a"),
				},
			},
			observed:    []*svcapitypes.Action{forwardAsReturned(1, "arn:tg/a")},
			expectDelta: false,
		},
		{
			name: "forwardConfig with ref and no weight, service adds targetGroupARN weight and stickiness",
			desired: []*svcapitypes.Action{
				{
					Type: aws.String("forward"),
					ForwardConfig: &svcapitypes.ForwardActionConfig{
						TargetGroups: []*svcapitypes.TargetGroupTuple{
							{TargetGroupRef: tgRef("my-tg"), TargetGroupARN: aws.String("arn:tg/a")},
						},
					},
				},
			},
			observed:    []*svcapitypes.Action{forwardAsReturned(1, "arn:tg/a")},
			expectDelta: false,
		},
		{
			name:        "weighted forward config unchanged",
			desired:     []*svcapitypes.Action{weightedForward(nil, 90, 10)},
			observed:    []*svcapitypes.Action{weightedForward(aws.Int64(1), 90, 10)},
			expectDelta: false,
		},
		{
			name:        "weights changed on the load balancer",
			desired:     []*svcapitypes.Action{weightedForward(nil, 90, 10)},
			observed:    []*svcapitypes.Action{weightedForward(aws.Int64(1), 50, 50)},
			expectDelta: true,
		},
		{
			name: "target group changed",
			desired: []*svcapitypes.Action{
				{Type: aws.String("forward"), TargetGroupARN: aws.String("arn:tg/b")},
			},
			observed:    []*svcapitypes.Action{forwardAsReturned(1, "arn:tg/a")},
			expectDelta: true,
		},
		{
			name: "explicit order differs",
			desired: []*svcapitypes.Action{
				{Type: aws.String("forward"), Order: aws.Int64(2), TargetGroupARN: aws.String("arn:tg/a")},
			},
			observed:    []*svcapitypes.Action{forwardAsReturned(1, "arn:tg/a")},
			expectDelta: true,
		},
		{
			name: "redirect with service keyword defaults",
			desired: []*svcapitypes.Action{
				{
					Type: aws.String("redirect"),
					RedirectConfig: &svcapitypes.RedirectActionConfig{
						Protocol:   aws.String("HTTPS"),
						StatusCode: aws.String("HTTP_301"),
					},
				},
			},
			observed: []*svcapitypes.Action{
				{Type: aws.String("redirect"), Order: aws.Int64(1), RedirectConfig: redirect(aws.String("/#{path}"))},
			},
			expectDelta: false,
		},
		{
			name: "redirect path changed",
			desired: []*svcapitypes.Action{
				{Type: aws.String("redirect"), RedirectConfig: redirect(aws.String("/new/#{path}"))},
			},
			observed: []*svcapitypes.Action{
				{Type: aws.String("redirect"), Order: aws.Int64(1), RedirectConfig: redirect(aws.String("/#{path}"))},
			},
			expectDelta: true,
		},
		{
			name: "authenticate-oidc then forward, secret not returned and defaults filled",
			desired: []*svcapitypes.Action{
				{Type: aws.String("authenticate-oidc"), AuthenticateOIDCConfig: oidc(aws.String("s3cret"), false)},
				{Type: aws.String("forward"), TargetGroupARN: aws.String("arn:tg/a")},
			},
			observed: []*svcapitypes.Action{
				{Type: aws.String("authenticate-oidc"), Order: aws.Int64(1), AuthenticateOIDCConfig: oidc(nil, true)},
				forwardAsReturned(2, "arn:tg/a"),
			},
			expectDelta: false,
		},
		{
			name: "action type changed",
			desired: []*svcapitypes.Action{
				{
					Type:                aws.String("fixed-response"),
					FixedResponseConfig: &svcapitypes.FixedResponseActionConfig{StatusCode: aws.String("503")},
				},
			},
			observed:    []*svcapitypes.Action{forwardAsReturned(1, "arn:tg/a")},
			expectDelta: true,
		},
		{
			name: "action count differs",
			desired: []*svcapitypes.Action{
				{Type: aws.String("forward"), TargetGroupARN: aws.String("arn:tg/a")},
			},
			observed: []*svcapitypes.Action{
				{Type: aws.String("authenticate-oidc"), Order: aws.Int64(1), AuthenticateOIDCConfig: oidc(nil, true)},
				forwardAsReturned(2, "arn:tg/a"),
			},
			expectDelta: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			desired := &resource{
				ko: &svcapitypes.Listener{
					Spec: svcapitypes.ListenerSpec{DefaultActions: tc.desired},
				},
			}
			observed := &resource{
				ko: &svcapitypes.Listener{
					Spec: svcapitypes.ListenerSpec{DefaultActions: tc.observed},
				},
			}

			delta := newResourceDelta(desired, observed)
			assert.Equal(t, tc.expectDelta, delta.DifferentAt("Spec.DefaultActions"),
				"Spec.DefaultActions delta mismatch")
			assert.Equal(t, tc.expectDelta, delta.DifferentAt("Spec"),
				"unexpected delta outside Spec.DefaultActions: %v", delta.Differences)
		})
	}
}
