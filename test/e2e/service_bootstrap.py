# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Bootstraps the resources required to run the EFS integration tests.
"""
import logging

from acktest.bootstrapping import BootstrapFailureException, Resources
from acktest.bootstrapping.vpc import VPC
from acktest.bootstrapping.function import Function
from acktest.aws.identity import get_region, get_account_id
from e2e import bootstrap_directory
from e2e.bootstrap_resources import BootstrapResources



def service_bootstrap() -> Resources:
    logging.getLogger().setLevel(logging.INFO)
    aws_region = get_region()
    account_id = get_account_id()
    code_uri1=f"{account_id}.dkr.ecr.{aws_region}.amazonaws.com/ack-e2e-testing-elbv2-controller:v1"
    code_uri2=f"{account_id}.dkr.ecr.{aws_region}.amazonaws.com/ack-e2e-testing-elbv2-controller:v2"

    resources = BootstrapResources(
        ACKVPC=VPC(name_prefix="ack-elb-vpc", num_public_subnet=2, num_private_subnet=2),
        Function1=Function(name_prefix="ack-elb-function-1", code_uri=code_uri1, service="elasticloadbalancing"),
        Function2=Function(name_prefix="ack-elb-function-2", code_uri=code_uri2, service="elasticloadbalancing")
    )

    try:
        resources.bootstrap()
    except BootstrapFailureException as ex:
        exit(254)

    return resources

if __name__ == "__main__":
    config = service_bootstrap()
    # Write config to current directory by default
    config.serialize(bootstrap_directory)
