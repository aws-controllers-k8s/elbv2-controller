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
			name: "nil conditions should not panic",
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

func TestNormalizeConditions(t *testing.T) {
	tests := []struct {
		name     string
		desired  *resource
		latest   *resource
		expected *resource
	}{
		{
			name: "host-header: desired has HostHeaderConfig only, latest has both - clear Values",
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
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "host-header: desired has Values only, latest has both - clear HostHeaderConfig",
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
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:            aws.String("host-header"),
								HostHeaderConfig: nil,
								Values:           []*string{aws.String("example.com")},
							},
						},
					},
				},
			},
		},
		{
			name: "host-header: desired has both, latest has both - keep both",
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
			latest: &resource{
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
			expected: &resource{
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
		},
		{
			name: "path-pattern: desired has PathPatternConfig only, latest has both - clear Values",
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
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("path-pattern"),
								PathPatternConfig: &svcapitypes.PathPatternConditionConfig{
									Values: []*string{aws.String("/api/*")},
								},
								Values: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "path-pattern: desired has Values only, latest has both - clear PathPatternConfig",
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
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:             aws.String("path-pattern"),
								PathPatternConfig: nil,
								Values:            []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
		},
		{
			name: "path-pattern: desired has both, latest has both - keep both",
			desired: &resource{
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
			latest: &resource{
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
			expected: &resource{
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
		},
		{
			name: "multiple conditions: mixed host-header and path-pattern",
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
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
								HostHeaderConfig: &svcapitypes.HostHeaderConditionConfig{
									Values: []*string{aws.String("example.com")},
								},
								Values: nil,
							},
							{
								Field:             aws.String("path-pattern"),
								PathPatternConfig: nil,
								Values:            []*string{aws.String("/api/*")},
							},
						},
					},
				},
			},
		},
		{
			name: "http-header condition: should not be modified",
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
			latest: &resource{
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
			expected: &resource{
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
		},
		{
			name: "mixed condition types: host-header and http-header",
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
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:            aws.String("host-header"),
								HostHeaderConfig: nil,
								Values:           []*string{aws.String("example.com")},
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
		},
		{
			name:     "nil desired resource - should not panic",
			desired:  nil,
			latest:   &resource{ko: &svcapitypes.Rule{Spec: svcapitypes.RuleSpec{Conditions: []*svcapitypes.RuleCondition{}}}},
			expected: &resource{ko: &svcapitypes.Rule{Spec: svcapitypes.RuleSpec{Conditions: []*svcapitypes.RuleCondition{}}}},
		},
		{
			name:     "nil latest resource - should not panic",
			desired:  &resource{ko: &svcapitypes.Rule{Spec: svcapitypes.RuleSpec{Conditions: []*svcapitypes.RuleCondition{}}}},
			latest:   nil,
			expected: nil,
		},
		{
			name: "nil conditions in desired - should not panic",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
			latest: &resource{
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
			expected: &resource{
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
		},
		{
			name: "nil conditions in latest - should not panic",
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
			latest: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: nil,
					},
				},
			},
		},
		{
			name: "condition not found in latest - should not panic",
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
			latest: &resource{
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
			expected: &resource{
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
		},
		{
			name: "desired has neither HostHeaderConfig nor Values - clear both in latest",
			desired: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field: aws.String("host-header"),
							},
						},
					},
				},
			},
			latest: &resource{
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
			expected: &resource{
				ko: &svcapitypes.Rule{
					Spec: svcapitypes.RuleSpec{
						Conditions: []*svcapitypes.RuleCondition{
							{
								Field:            aws.String("host-header"),
								HostHeaderConfig: nil,
								Values:           nil,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a deep copy of latest to avoid modifying the test input
			var latestCopy *resource
			if tt.latest != nil {
				latestCopy = &resource{ko: tt.latest.ko.DeepCopy()}
			}

			// Call normalizeConditions
			normalizeConditions(tt.desired, latestCopy)

			// Compare the result with expected
			if tt.expected == nil {
				assert.Nil(t, latestCopy)
			} else {
				assert.NotNil(t, latestCopy)
				assert.Equal(t, len(tt.expected.ko.Spec.Conditions), len(latestCopy.ko.Spec.Conditions))

				for i, expectedCond := range tt.expected.ko.Spec.Conditions {
					actualCond := latestCopy.ko.Spec.Conditions[i]

					// Check Field
					if expectedCond.Field == nil {
						assert.Nil(t, actualCond.Field)
					} else {
						assert.NotNil(t, actualCond.Field)
						assert.Equal(t, *expectedCond.Field, *actualCond.Field)
					}

					// Check HostHeaderConfig
					if expectedCond.HostHeaderConfig == nil {
						assert.Nil(t, actualCond.HostHeaderConfig, "HostHeaderConfig should be nil")
					} else {
						assert.NotNil(t, actualCond.HostHeaderConfig, "HostHeaderConfig should not be nil")
						assert.Equal(t, expectedCond.HostHeaderConfig, actualCond.HostHeaderConfig)
					}

					// Check PathPatternConfig
					if expectedCond.PathPatternConfig == nil {
						assert.Nil(t, actualCond.PathPatternConfig, "PathPatternConfig should be nil")
					} else {
						assert.NotNil(t, actualCond.PathPatternConfig, "PathPatternConfig should not be nil")
						assert.Equal(t, expectedCond.PathPatternConfig, actualCond.PathPatternConfig)
					}

					// Check Values
					if expectedCond.Values == nil {
						assert.Nil(t, actualCond.Values, "Values should be nil")
					} else {
						assert.NotNil(t, actualCond.Values, "Values should not be nil")
						assert.Equal(t, expectedCond.Values, actualCond.Values)
					}

					// Check HTTPHeaderConfig (should not be modified)
					if expectedCond.HTTPHeaderConfig != nil {
						assert.Equal(t, expectedCond.HTTPHeaderConfig, actualCond.HTTPHeaderConfig)
					}
				}
			}
		})
	}
}
