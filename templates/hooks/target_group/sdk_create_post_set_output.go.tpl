	if ko.Spec.Targets != nil {
		return nil, ackrequeue.NeededAfter(fmt.Errorf("Requing due to register targets in UPDATE"), RequeueAfterUpdateDuration)
	}
