	err = rm.describeTargets(ctx, &resource{ko})
	if err != nil {
		return nil, err
	}

	rm.setStatusDefaults(ko)
