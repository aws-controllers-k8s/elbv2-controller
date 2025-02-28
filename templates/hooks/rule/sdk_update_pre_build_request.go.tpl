	if delta.DifferentAt("Spec.Priority") {
		if err = rm.setRulePriority(ctx, desired); err != nil {
			return nil, err
		}
	} else if !delta.DifferentExcept("Spec.Priority") {
		return desired, nil
	}

    if delta.DifferentAt("Spec.Tags") {
        err = rm.updateTags(ctx, desired, latest)
        if err != nil {
            return nil, err
        }
    }
    if !delta.DifferentAt("Spec.Tags") {
        return desired, nil
    }