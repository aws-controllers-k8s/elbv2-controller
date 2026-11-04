package listener

import (
	"sort"

	svcapitypes "github.com/aws-controllers-k8s/elbv2-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws/aws-sdk-go-v2/aws"
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

// customCompareDefaultActions compares DefaultActions in the form
// DescribeListeners reports them: both sides get the members the service
// fills in when the spec leaves them unset, target groups are sorted by ARN,
// and the members the service never returns (the k8s-only references and the
// OIDC client secret) are left out. A member removed from the spec then
// differs from the value still on the load balancer and reaches
// ModifyListener.
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
	if !equality.Semantic.DeepEqual(normalizeActions(desired), normalizeActions(observed)) {
		delta.Add("Spec.DefaultActions", desired, observed)
	}
}

// normalizeActions returns copies of actions in the form DescribeListeners
// returns them.
func normalizeActions(actions []*svcapitypes.Action) []*svcapitypes.Action {
	out := make([]*svcapitypes.Action, len(actions))
	for i, action := range actions {
		if action == nil {
			continue
		}
		c := action.DeepCopy()
		c.TargetGroupRef = nil
		if c.Order == nil {
			c.Order = aws.Int64(int64(i + 1))
		}
		normalizeForward(c)
		normalizeRedirect(c.RedirectConfig)
		normalizeFixedResponse(c.FixedResponseConfig)
		normalizeOIDC(c.AuthenticateOIDCConfig)
		normalizeCognito(c.AuthenticateCognitoConfig)
		out[i] = c
	}
	return out
}

// normalizeForward fills the forward members the service returns: a forward
// with only targetGroupARN comes back with a one-entry forwardConfig, a
// one-entry forwardConfig comes back with the action-level targetGroupARN,
// an unset weight is 1 and unset stickiness is disabled. Target groups are
// returned in ARN order, not declared order.
func normalizeForward(c *svcapitypes.Action) {
	if c.ForwardConfig == nil && c.TargetGroupARN != nil {
		c.ForwardConfig = &svcapitypes.ForwardActionConfig{
			TargetGroups: []*svcapitypes.TargetGroupTuple{{TargetGroupARN: c.TargetGroupARN}},
		}
	}
	fc := c.ForwardConfig
	if fc == nil {
		return
	}
	if fc.TargetGroupStickinessConfig == nil {
		fc.TargetGroupStickinessConfig = &svcapitypes.TargetGroupStickinessConfig{}
	}
	if fc.TargetGroupStickinessConfig.Enabled == nil {
		fc.TargetGroupStickinessConfig.Enabled = aws.Bool(false)
	}
	if !*fc.TargetGroupStickinessConfig.Enabled {
		fc.TargetGroupStickinessConfig.DurationSeconds = nil
	}
	for _, tg := range fc.TargetGroups {
		if tg == nil {
			continue
		}
		tg.TargetGroupRef = nil
		if tg.Weight == nil {
			tg.Weight = aws.Int64(1)
		}
	}
	sort.SliceStable(fc.TargetGroups, func(i, j int) bool {
		return targetGroupARN(fc.TargetGroups[i]) < targetGroupARN(fc.TargetGroups[j])
	})
	if c.TargetGroupARN == nil && len(fc.TargetGroups) == 1 && fc.TargetGroups[0] != nil {
		c.TargetGroupARN = fc.TargetGroups[0].TargetGroupARN
	}
}

func targetGroupARN(tg *svcapitypes.TargetGroupTuple) string {
	if tg == nil {
		return ""
	}
	return aws.ToString(tg.TargetGroupARN)
}

// normalizeRedirect fills the keyword the service returns for each unset
// redirect component.
func normalizeRedirect(r *svcapitypes.RedirectActionConfig) {
	if r == nil {
		return
	}
	if r.Protocol == nil {
		r.Protocol = aws.String("#{protocol}")
	}
	if r.Port == nil {
		r.Port = aws.String("#{port}")
	}
	if r.Host == nil {
		r.Host = aws.String("#{host}")
	}
	if r.Path == nil {
		r.Path = aws.String("/#{path}")
	}
	if r.Query == nil {
		r.Query = aws.String("#{query}")
	}
}

func normalizeFixedResponse(f *svcapitypes.FixedResponseActionConfig) {
	if f == nil {
		return
	}
	if f.ContentType == nil {
		f.ContentType = aws.String("text/plain")
	}
}

// normalizeOIDC fills the session defaults and drops the client secret, which
// the service never returns: a spec that changes only the secret produces no
// delta and the new secret only reaches AWS with another change.
func normalizeOIDC(o *svcapitypes.AuthenticateOIDCActionConfig) {
	if o == nil {
		return
	}
	o.ClientSecret = nil
	o.UseExistingClientSecret = nil
	if len(o.AuthenticationRequestExtraParams) == 0 {
		o.AuthenticationRequestExtraParams = nil
	}
	o.OnUnauthenticatedRequest, o.Scope, o.SessionCookieName, o.SessionTimeout = authDefaults(
		o.OnUnauthenticatedRequest, o.Scope, o.SessionCookieName, o.SessionTimeout)
}

func normalizeCognito(c *svcapitypes.AuthenticateCognitoActionConfig) {
	if c == nil {
		return
	}
	if len(c.AuthenticationRequestExtraParams) == 0 {
		c.AuthenticationRequestExtraParams = nil
	}
	c.OnUnauthenticatedRequest, c.Scope, c.SessionCookieName, c.SessionTimeout = authDefaults(
		c.OnUnauthenticatedRequest, c.Scope, c.SessionCookieName, c.SessionTimeout)
}

// authDefaults returns the authenticate-oidc and authenticate-cognito session
// members with the service defaults filled in.
func authDefaults(onUnauthenticated, scope, cookie *string, timeout *int64) (*string, *string, *string, *int64) {
	if onUnauthenticated == nil {
		onUnauthenticated = aws.String("authenticate")
	}
	if scope == nil {
		scope = aws.String("openid")
	}
	if cookie == nil {
		cookie = aws.String("AWSELBAuthSessionCookie")
	}
	if timeout == nil {
		timeout = aws.Int64(604800)
	}
	return onUnauthenticated, scope, cookie, timeout
}
