	if ko.Spec.Targets != nil || len(ko.Spec.Attributes) > 0 {
		return nil, ackrequeue.NeededAfter(fmt.Errorf("Requeuing for post-create updates (targets or attributes)"), RequeueAfterUpdateDuration)
	}
