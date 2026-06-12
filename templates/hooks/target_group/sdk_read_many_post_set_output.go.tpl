	// When target management is ignored, skip reading targets from AWS so that
	// externally registered targets are not treated as drift.
	if !isTargetManagementIgnored(r) {
		err = rm.describeTargets(ctx, &resource{ko})
		if err != nil {
			return nil, err
		}
	}

	rm.setStatusDefaults(ko)

	// Set target group attributes
	ko.Spec.Attributes, err = rm.getTargetGroupAttributes(ctx, ko)
	if err != nil {
		return nil, err
	}
