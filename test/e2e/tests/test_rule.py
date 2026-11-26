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
    replacements["TARGET_GROUP_NAME"] = target_group_cr["spec"]["name"]

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


@pytest.fixture(scope="module")
def rule_with_transforms(elbv2_client, simple_listener, simple_target_group, simple_load_balancer):
    (_, listener_cr) = simple_listener
    (_, target_group_cr, _) = simple_target_group

    resource_name = random_suffix_name("rule-tf", 16)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["RULE_NAME"] = resource_name
    replacements["LISTENER_ARN"] = listener_cr["status"]["ackResourceMetadata"]["arn"]
    replacements["TARGET_GROUP_NAME"] = target_group_cr["spec"]["name"]

    resource_data = load_elbv2_resource(
        "rule_transforms",
        additional_replacements=replacements,
    )

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


def _transforms_by_type(rule):
    return {t["Type"]: t for t in rule.get("Transforms", [])}


@service_marker
class TestRuleTransforms:
    def test_transforms(self, elbv2_client, rule_with_transforms):
        (ref, cr) = rule_with_transforms
        rule_arn = cr["status"]["ackResourceMetadata"]["arn"]
        validator = ELBValidator(elbv2_client)

        # Create: both transforms reach the listener rule
        transforms = _transforms_by_type(validator.get_rule(rule_arn))
        assert transforms["url-rewrite"]["UrlRewriteConfig"]["Rewrites"] == [
            {"Regex": "^/old/(.*)$", "Replace": "/new/$1"}
        ]
        assert transforms["host-header-rewrite"]["HostHeaderRewriteConfig"]["Rewrites"] == [
            {"Regex": "^.*$", "Replace": "example.com"}
        ]
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)
        # The spec keeps the order it was written in, whatever order ELBv2 returns.
        assert [t["type"] for t in k8s.get_resource(ref)["spec"]["transforms"]] == [
            "url-rewrite", "host-header-rewrite"
        ]

        # Update: change the url-rewrite, drop the host-header-rewrite
        updates = {
            "spec": {
                "transforms": [{
                    "type": "url-rewrite",
                    "urlRewriteConfig": {
                        "rewrites": [{"regex": "^/old2/(.*)$", "replace": "/new2/$1"}]
                    },
                }]
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        transforms = _transforms_by_type(validator.get_rule(rule_arn))
        assert list(transforms) == ["url-rewrite"]
        assert transforms["url-rewrite"]["UrlRewriteConfig"]["Rewrites"] == [
            {"Regex": "^/old2/(.*)$", "Replace": "/new2/$1"}
        ]
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        # Remove every transform: needs ResetTransforms on ModifyRule
        k8s.patch_custom_resource(ref, {"spec": {"transforms": None}})
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        assert not validator.get_rule(rule_arn).get("Transforms")
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)
