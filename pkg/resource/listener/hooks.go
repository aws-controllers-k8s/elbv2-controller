package listener

import (
	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"k8s.io/apimachinery/pkg/api/equality"
)

// customCheckRequiredFieldsMissingMethod returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status.
func (rm *resourceManager) customCheckRequiredFieldsMissingMethod(
	r *resource,
) bool {
	return r.Identifiers().ARN() == nil
}

func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	customCompareDefaultActions(delta, a, b)
}

// customCompareDefaultActions compares DefaultActions on copies without the
// k8s-only TargetGroupRef, the write-only OIDC client secret and the members
// DescribeListeners fills in when the spec leaves them unset.
func customCompareDefaultActions(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if a == nil || b == nil {
		return
	}
	desired := a.ko.Spec.DefaultActions
	observed := b.ko.Spec.DefaultActions
	if len(desired) != len(observed) {
		delta.Add("Spec.DefaultActions", desired, observed)
		return
	}
	for i := range desired {
		if !defaultActionEqual(desired[i], observed[i]) {
			delta.Add("Spec.DefaultActions", desired, observed)
			return
		}
	}
}

func defaultActionEqual(desired, observed *svcapitypes.Action) bool {
	if desired == nil || observed == nil {
		return desired == observed
	}
	d := desired.DeepCopy()
	o := observed.DeepCopy()
	d.TargetGroupRef, o.TargetGroupRef = nil, nil
	if d.Order == nil {
		o.Order = nil
	}
	if d.TargetGroupARN == nil {
		o.TargetGroupARN = nil
	}
	if d.ForwardConfig == nil {
		o.ForwardConfig = nil
	} else if o.ForwardConfig != nil {
		normalizeForwardConfig(d.ForwardConfig, o.ForwardConfig)
	}
	if d.RedirectConfig != nil && o.RedirectConfig != nil {
		normalizeRedirectConfig(d.RedirectConfig, o.RedirectConfig)
	}
	if d.AuthenticateOIDCConfig != nil && o.AuthenticateOIDCConfig != nil {
		normalizeOIDCConfig(d.AuthenticateOIDCConfig, o.AuthenticateOIDCConfig)
	}
	if d.AuthenticateCognitoConfig != nil && o.AuthenticateCognitoConfig != nil {
		normalizeCognitoConfig(d.AuthenticateCognitoConfig, o.AuthenticateCognitoConfig)
	}
	return equality.Semantic.Equalities.DeepEqual(d, o)
}

func normalizeForwardConfig(d, o *svcapitypes.ForwardActionConfig) {
	if d.TargetGroupStickinessConfig == nil {
		o.TargetGroupStickinessConfig = nil
	}
	for i, tg := range d.TargetGroups {
		if tg == nil || i >= len(o.TargetGroups) || o.TargetGroups[i] == nil {
			continue
		}
		tg.TargetGroupRef, o.TargetGroups[i].TargetGroupRef = nil, nil
		if tg.Weight == nil {
			o.TargetGroups[i].Weight = nil
		}
	}
}

func normalizeRedirectConfig(d, o *svcapitypes.RedirectActionConfig) {
	if d.Host == nil {
		o.Host = nil
	}
	if d.Path == nil {
		o.Path = nil
	}
	if d.Port == nil {
		o.Port = nil
	}
	if d.Protocol == nil {
		o.Protocol = nil
	}
	if d.Query == nil {
		o.Query = nil
	}
}

func normalizeOIDCConfig(d, o *svcapitypes.AuthenticateOIDCActionConfig) {
	d.ClientSecret, o.ClientSecret = nil, nil
	d.UseExistingClientSecret, o.UseExistingClientSecret = nil, nil
	if d.OnUnauthenticatedRequest == nil {
		o.OnUnauthenticatedRequest = nil
	}
	if d.Scope == nil {
		o.Scope = nil
	}
	if d.SessionCookieName == nil {
		o.SessionCookieName = nil
	}
	if d.SessionTimeout == nil {
		o.SessionTimeout = nil
	}
}

func normalizeCognitoConfig(d, o *svcapitypes.AuthenticateCognitoActionConfig) {
	if d.OnUnauthenticatedRequest == nil {
		o.OnUnauthenticatedRequest = nil
	}
	if d.Scope == nil {
		o.Scope = nil
	}
	if d.SessionCookieName == nil {
		o.SessionCookieName = nil
	}
	if d.SessionTimeout == nil {
		o.SessionTimeout = nil
	}
}
