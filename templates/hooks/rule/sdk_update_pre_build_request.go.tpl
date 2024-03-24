	if delta.DifferentAt("Spec.Priority") {
		if err = rm.setRulePriority(ctx, desired); err != nil {
			return nil, err
		}
	} else if !delta.DifferentExcept("Spec.Priority") {
		return desired, nil
	}
