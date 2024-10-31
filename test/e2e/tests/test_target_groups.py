# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the ELB TargetGroups.
"""

import logging
import time

import pytest
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import CRD_GROUP, CRD_VERSION, load_elbv2_resource, service_marker
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import ELBValidator

RESOURCE_PLURAL = "targetgroups"

CREATE_WAIT_AFTER_SECONDS = 30
UPDATE_WAIT_AFTER_SECONDS = 20
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def simple_target_group(elbv2_client):

    resource_name = random_suffix_name("tg", 16)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["TARGET_GROUP_NAME"] = resource_name

    resource_data = load_elbv2_resource(
        "target_group",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        resource_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr, resource_name)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    time.sleep(DELETE_WAIT_AFTER_SECONDS)

    validator = ELBValidator(elbv2_client)
    assert not validator.target_group_exists(resource_name)

@service_marker
@pytest.mark.canary
class TestTargetGroups:
    def test_create_delete(self, elbv2_client, simple_target_group):
        (ref, cr, tg_name) = simple_target_group

        validator = ELBValidator(elbv2_client)
        cr = k8s.get_resource(ref)
        assert validator.target_group_exists(tg_name)
        assert cr is not None
        assert "spec" in cr
        assert "targets" in cr["spec"]
        assert len(cr['spec']["targets"]) == 1
        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        resource_arn = cr['status']['ackResourceMetadata']['arn']
        targets = validator.get_registered_targets(resource_arn)
        assert len(targets) == 1
        assert targets[0]["Target"]["Id"] == REPLACEMENT_VALUES["FUNCTION_ARN_1"]
        # Update healthyThresholdCount
        updates = {
            "spec": {
                "healthyThresholdCount": 10,
                "targets": [
                    {
                        "id": REPLACEMENT_VALUES["FUNCTION_ARN_2"],
                    }
                ]
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        tg_healthy_threshold_count = validator.get_target_group(tg_name)["HealthyThresholdCount"]
        assert tg_healthy_threshold_count == 10
        targets = validator.get_registered_targets(resource_arn)
        assert len(targets) == 1
        assert targets[0]["Target"]["Id"] == REPLACEMENT_VALUES["FUNCTION_ARN_2"]

        updates = {
            "spec": {
                "targets": []
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        targets = validator.get_registered_targets(resource_arn)
        assert len(targets) == 0
