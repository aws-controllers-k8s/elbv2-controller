	if ko.Spec.Targets != nil {
		return nil, ackrequeue.NeededAfter(fmt.Errorf("Requing due to register targets in UPDATE"), RequeueAfterUpdateDuration)
	}
	
	if desired.ko.Spec.Tags != nil {
        return nil, ackrequeue.NeededAfter(fmt.Errorf("Requing due to tags in UPDATE"), RequeueAfterUpdateDuration)
    } 