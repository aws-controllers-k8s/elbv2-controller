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
	"testing"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
)

func TestCustomCompareConditions(t *testing.T) {
	tests := []struct {
		name            string
		desired         *resource
		observed        *resource
		expectDelta     bool
		deltaFieldCount int
	}{
		{
			name: "user specified hostHeaderConfig, observed has both - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")}, // AWS returns this too
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "user specified values, observed has both - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("host-header"),
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "user specified pathPatternConfig, observed has both - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/api/*")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/api/*")},
								},
								Values: []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "multiple conditions with mixed formats - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
							},
							{
								Field:  aws.String("path-pattern"),
								Values: []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")},
							},
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/api/*")},
								},
								Values: []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "user specified hostHeaderConfig with different values - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("different.com")},
								},
								Values: []*string{aws.String("different.com")},
							},
						},
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
		{
			name: "user specified values with different values - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("path-pattern"),
								Values: []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/different/*")},
								},
								Values: []*string{aws.String("/different/*")},
							},
						},
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
		{
			name: "user specified both formats (workaround) - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name:    "nil desired resource should not panic",
			desired: nil,
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("host-header"),
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "nil observed resource should not panic",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("host-header"),
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			observed:        nil,
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "nil desired conditions and non-nil observed - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("host-header"),
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
		{
			name: "non-nil desired conditions and nil observed - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("host-header"),
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
		{
			name: "Both desired and  observed conditions are nil- no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "http-header condition - matching values - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("http-header"),
								HTTPHeaderConfig: &svcapitypes.HTTPHeaderConditionConfig{
									HTTPHeaderName: aws.String("X-Custom-Header"),
									Values:         []*string{aws.String("value1"), aws.String("value2")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("http-header"),
								HTTPHeaderConfig: &svcapitypes.HTTPHeaderConditionConfig{
									HTTPHeaderName: aws.String("X-Custom-Header"),
									Values:         []*string{aws.String("value1"), aws.String("value2")},
								},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "http-header condition - different values - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("http-header"),
								HTTPHeaderConfig: &svcapitypes.HTTPHeaderConditionConfig{
									HTTPHeaderName: aws.String("X-Custom-Header"),
									Values:         []*string{aws.String("value1")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("http-header"),
								HTTPHeaderConfig: &svcapitypes.HTTPHeaderConditionConfig{
									HTTPHeaderName: aws.String("X-Custom-Header"),
									Values:         []*string{aws.String("value2")},
								},
							},
						},
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
		{
			name: "http-request-method condition - matching - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("http-request-method"),
								HTTPRequestMethodConfig: &svcapitypes.HTTPRequestMethodConditionConfig{
									Values: []*string{aws.String("GET"), aws.String("POST")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("http-request-method"),
								HTTPRequestMethodConfig: &svcapitypes.HTTPRequestMethodConditionConfig{
									Values: []*string{aws.String("GET"), aws.String("POST")},
								},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "source-ip condition - matching - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("source-ip"),
								SourceIPConfig: &svcapitypes.SourceIPConditionConfig{
									Values: []*string{aws.String("192.168.1.0/24")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("source-ip"),
								SourceIPConfig: &svcapitypes.SourceIPConditionConfig{
									Values: []*string{aws.String("192.168.1.0/24")},
								},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "query-string condition - matching - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("query-string"),
								QueryStringConfig: &svcapitypes.QueryStringConditionConfig{
									Values: []*svcapitypes.QueryStringKeyValuePair{
										{
											Key:   aws.String("version"),
											Value: aws.String("v1"),
										},
									},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("query-string"),
								QueryStringConfig: &svcapitypes.QueryStringConditionConfig{
									Values: []*svcapitypes.QueryStringKeyValuePair{
										{
											Key:   aws.String("version"),
											Value: aws.String("v1"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "mixed condition types - host-header and http-header - no delta",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:  aws.String("host-header"),
								Values: []*string{aws.String("example.com")},
							},
							{
								Field: aws.String("http-header"),
								HTTPHeaderConfig: &svcapitypes.HTTPHeaderConditionConfig{
									HTTPHeaderName: aws.String("X-Custom-Header"),
									Values:         []*string{aws.String("value1")},
								},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")},
							},
							{
								Field: aws.String("http-header"),
								HTTPHeaderConfig: &svcapitypes.HTTPHeaderConditionConfig{
									HTTPHeaderName: aws.String("X-Custom-Header"),
									Values:         []*string{aws.String("value1")},
								},
							},
						},
					},
				},
			},
			expectDelta:     false,
			deltaFieldCount: 0,
		},
		{
			name: "user specified both hostHeaderConfig and values with DIFFERENT values - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("different.com")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
		{
			name: "user specified both pathPatternConfig and values with DIFFERENT values - delta expected",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/api/*")},
								},
								Values: []*string{aws.String("/different/*")},
							},
						},
					},
				},
			},
			observed: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/api/*")},
								},
								Values: []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
			expectDelta:     true,
			deltaFieldCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := ackcompare.NewDelta()
			customCompareConditions(delta, tt.desired, tt.observed)

			if tt.expectDelta {
				assert.True(t, len(delta.Differences) > 0, "Expected delta but got none")
				if tt.deltaFieldCount > 0 {
					assert.Equal(t, tt.deltaFieldCount, len(delta.Differences), "Delta field count mismatch")
				}
			} else {
				assert.Equal(t, 0, len(delta.Differences), "Expected no delta but got: %v", delta.Differences)
			}
		})
	}
}
