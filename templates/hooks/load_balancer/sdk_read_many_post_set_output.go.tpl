	if err := rm.setResourceAdditionalFields(ctx, ko); err != nil {
		return nil, err
	}

	ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
	if err != nil {
		return nil, err
	}