	ko.Spec.Priority = priorityFromSDK(resp.Rules[0].Priority)
	ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
