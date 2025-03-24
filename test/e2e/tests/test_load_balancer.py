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

"""Integration tests for the ELB LoadBalancer API.
"""

import logging
import time

import pytest
from acktest.k8s import resource as k8s
from acktest import tags
from acktest.resources import random_suffix_name
from e2e import CRD_GROUP, CRD_VERSION, load_elbv2_resource, service_marker
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import ELBValidator

RESOURCE_PLURAL = "loadbalancers"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def simple_load_balancer(elbv2_client):

    resource_name = random_suffix_name("lb", 16)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["LOAD_BALANCER_NAME"] = resource_name

    resource_data = load_elbv2_resource(
        "load_balancer",
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
    assert not validator.load_balancer_exists(resource_name)

@service_marker
@pytest.mark.canary
class TestLoadBalancer:
    def test_create_delete(self, elbv2_client, simple_load_balancer):
        (ref, cr, lb_name) = simple_load_balancer
        assert lb_name is not None

        validator = ELBValidator(elbv2_client)
        assert validator.load_balancer_exists(lb_name)

    
        cr = k8s.get_resource(ref)
        
        assert cr is not None
        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']

        # Update attributes
        updates = {
            "spec": {
                "attributes": [
                    {
                        "key": "client_keep_alive.seconds",
                        "value": "4200"
                    }
                ]
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        lbAttributes = validator.get_load_balancer_attributes(cr["status"]["ackResourceMetadata"]["arn"])
        assert lbAttributes is not None

        # find the attribute we just updated
        for attribute in lbAttributes:
            if attribute["Key"] == "client_keep_alive.seconds":
                assert attribute["Value"] == "4200"
                break
        else:
            assert False, "Attribute not found"


    def test_tags(self, elbv2_client, simple_load_balancer):
        (ref, cr, lb_name) = simple_load_balancer
        assert lb_name is not None

        validator = ELBValidator(elbv2_client)
        assert validator.load_balancer_exists(lb_name)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']

        assert 'tags' in cr['spec']
        user_tags = cr['spec']['tags']

        response_tags = validator.get_tags(arn)

        assert 'Tags' in response_tags
        actual_tags = response_tags["Tags"]

        tags.assert_ack_system_tags(
            tags=actual_tags,
        )

        user_tags = [{"Key": d["key"], "Value": d["value"]} for d in user_tags]
        tags.assert_equal_without_ack_tags(
            expected=user_tags,
            actual=actual_tags,
        )

        # Update attributes
        updates = {
            "spec": {
                "tags": [
                    {
                        "key": "updated_tagKey",
                        "value": "updated_tagValue"
                    },
                    {
                        "key": "new_tagKey",
                        "value": "new_tagValue"
                    }
                ]
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']
        
        assert 'tags' in cr['spec']
        user_tags = cr['spec']['tags']

        response_tags = validator.get_tags(arn)

        assert 'Tags' in response_tags
        actual_tags = response_tags["Tags"]

        tags.assert_ack_system_tags(
            tags=actual_tags,
        )

        user_tags = [{"Key": d["key"], "Value": d["value"]} for d in user_tags]
        tags.assert_equal_without_ack_tags(
            expected=user_tags,
            actual=actual_tags,
        )