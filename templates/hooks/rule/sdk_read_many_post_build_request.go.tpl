    // The ARN is guaranteed to be checked by the sdkFind method. We can safely cast it here.
    input.RuleArns = []string{(string)(*r.ko.Status.ACKResourceMetadata.ARN)}
    // Unset the LoadBalancerArn field since we can't set both ListenerArn and LoadBalancerArn
    // Probably needs to be done in the code-generator. @a-hilaly.
    input.ListenerArn = nil