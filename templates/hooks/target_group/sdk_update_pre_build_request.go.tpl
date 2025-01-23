	if delta.DifferentAt("Spec.Targets") {
		added, removed := getTargetsDifference(latest.ko.Spec.Targets, desired.ko.Spec.Targets)
		arn := (string)(*latest.ko.Status.ACKResourceMetadata.ARN)
		if len(removed) > 0 {
			err = rm.deregisterTargets(ctx, arn, removed)
			if err != nil {
				return nil, err
			}
		}
		if len(added) > 0 {
			err = rm.registerTargets(ctx, arn, added)
			if err != nil {
				return nil, err
			}
		}
	}

	if !delta.DifferentExcept("Spec.Targets") {
		return desired, nil
	}