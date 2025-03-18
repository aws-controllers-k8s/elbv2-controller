    if delta.DifferentAt("Spec.Tags") {
         err = rm.updateTags(ctx, desired, latest)
         if err != nil {
             return nil, err
         }
     }
     if !delta.DifferentExcept("Spec.Tags") {
         return desired, nil
     }