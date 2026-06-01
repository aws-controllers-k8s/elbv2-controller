	if delta.DifferentAt("Spec.Targets") {
		added, removed := getTargetsDifference(latest.ko.Spec.Targets, desired.ko.Spec.Targets)
		if latest.ko.Status.ACKResourceMetadata == nil || latest.ko.Status.ACKResourceMetadata.ARN == nil {
			return nil, fmt.Errorf("target group ARN is not yet available")
		}
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

	if delta.DifferentAt("Spec.Attributes") {
		if err := rm.updateTargetGroupAttributes(ctx, desired, latest); err != nil {
			return nil, err
		}
	}

	if !delta.DifferentExcept("Spec.Targets", "Spec.Attributes") {
		return desired, nil
	}
