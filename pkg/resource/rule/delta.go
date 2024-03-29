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

// Code generated by ack-generate. DO NOT EDIT.

package rule

import (
	"bytes"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	acktags "github.com/aws-controllers-k8s/runtime/pkg/tags"
)

// Hack to avoid import errors during build...
var (
	_ = &bytes.Buffer{}
	_ = &reflect.Method{}
	_ = &acktags.Tags{}
)

// newResourceDelta returns a new `ackcompare.Delta` used to compare two
// resources
func newResourceDelta(
	a *resource,
	b *resource,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if (a == nil && b != nil) ||
		(a != nil && b == nil) {
		delta.Add("", a, b)
		return delta
	}

	if len(a.ko.Spec.Actions) != len(b.ko.Spec.Actions) {
		delta.Add("Spec.Actions", a.ko.Spec.Actions, b.ko.Spec.Actions)
	} else if len(a.ko.Spec.Actions) > 0 {
		if !reflect.DeepEqual(a.ko.Spec.Actions, b.ko.Spec.Actions) {
			delta.Add("Spec.Actions", a.ko.Spec.Actions, b.ko.Spec.Actions)
		}
	}
	if len(a.ko.Spec.Conditions) != len(b.ko.Spec.Conditions) {
		delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
	} else if len(a.ko.Spec.Conditions) > 0 {
		if !reflect.DeepEqual(a.ko.Spec.Conditions, b.ko.Spec.Conditions) {
			delta.Add("Spec.Conditions", a.ko.Spec.Conditions, b.ko.Spec.Conditions)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ListenerARN, b.ko.Spec.ListenerARN) {
		delta.Add("Spec.ListenerARN", a.ko.Spec.ListenerARN, b.ko.Spec.ListenerARN)
	} else if a.ko.Spec.ListenerARN != nil && b.ko.Spec.ListenerARN != nil {
		if *a.ko.Spec.ListenerARN != *b.ko.Spec.ListenerARN {
			delta.Add("Spec.ListenerARN", a.ko.Spec.ListenerARN, b.ko.Spec.ListenerARN)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.ListenerRef, b.ko.Spec.ListenerRef) {
		delta.Add("Spec.ListenerRef", a.ko.Spec.ListenerRef, b.ko.Spec.ListenerRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Priority, b.ko.Spec.Priority) {
		delta.Add("Spec.Priority", a.ko.Spec.Priority, b.ko.Spec.Priority)
	} else if a.ko.Spec.Priority != nil && b.ko.Spec.Priority != nil {
		if *a.ko.Spec.Priority != *b.ko.Spec.Priority {
			delta.Add("Spec.Priority", a.ko.Spec.Priority, b.ko.Spec.Priority)
		}
	}
	if len(a.ko.Spec.Tags) != len(b.ko.Spec.Tags) {
		delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
	} else if len(a.ko.Spec.Tags) > 0 {
		if !reflect.DeepEqual(a.ko.Spec.Tags, b.ko.Spec.Tags) {
			delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
		}
	}

	return delta
}
