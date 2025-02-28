	if err := rm.setResourceAdditionalFields(ctx, ko); err != nil {
		return nil, err
	}
	
	Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))