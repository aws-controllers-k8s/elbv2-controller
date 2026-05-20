	err = rm.describeTargets(ctx, &resource{ko})
	if err != nil {
		return nil, err
	}

	rm.setStatusDefaults(ko)

	// Set target group attributes
	ko.Spec.Attributes, err = rm.getTargetGroupAttributes(ctx, ko)
	if err != nil {
		return nil, err
	}
