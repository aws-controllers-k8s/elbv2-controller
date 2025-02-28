	err = rm.describeTargets(ctx, &resource{ko})
	if err != nil {
		return nil, err
	}

	rm.setStatusDefaults(ko)

	Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))