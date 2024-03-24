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

"""Integration tests for the ELB Rule API.
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

from .test_listener import simple_listener
from .test_load_balancer import simple_load_balancer
from .test_target_groups import simple_target_group

RESOURCE_PLURAL = "rules"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def simple_rule(elbv2_client, simple_listener, simple_target_group, simple_load_balancer):
    (listener_ref, listener_cr) = simple_listener
    (target_group_ref, target_group_cr, _) = simple_target_group

    resource_name = random_suffix_name("rule", 16)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["RULE_NAME"] = resource_name
    replacements["LISTENER_ARN"] = listener_cr["status"]["ackResourceMetadata"]["arn"]
    replacements["TARGET_GROUP_ARN"] = target_group_cr["status"]["ackResourceMetadata"]["arn"]

    resource_data = load_elbv2_resource(
        "rule",
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

    yield (ref, cr)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    time.sleep(DELETE_WAIT_AFTER_SECONDS)

    validator = ELBValidator(elbv2_client)
    assert not validator.rule_exists(cr["status"]["ackResourceMetadata"]["arn"])

@service_marker
@pytest.mark.canary
class TestRule:
    def test_create_delete(self, elbv2_client, simple_rule):
        (ref, cr) = simple_rule
        assert cr is not None
        rule_arn = cr["status"]["ackResourceMetadata"]["arn"]

        validator = ELBValidator(elbv2_client)
        rule = validator.get_rule(rule_arn)
        assert rule is not None

        # Update settings
        updates = {
            "spec": {
                "priority": 500,
                "conditions": [{
                    "field": "http-request-method",
                    "httpRequestMethodConfig": {
                        "values": ["GET"]
                    }
                }]
            },
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        rule = validator.get_rule(rule_arn)
        assert rule is not None
        assert rule["Priority"] == "500"
        assert rule["Conditions"][0]["Field"] == "http-request-method"
        assert rule["Conditions"][0]["HttpRequestMethodConfig"]["Values"] == ["GET"]
