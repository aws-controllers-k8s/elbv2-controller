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

"""Helper functions for ELB e2e tests
"""

class ELBValidator:
    def __init__(self, elbv2_client):
        self.elbv2_client = elbv2_client

    def get_load_balancer(self, name):
        try:
            response = self.elbv2_client.describe_load_balancers(Names=[name])
            return response['LoadBalancers'][0]
        except Exception:
            return None

    def load_balancer_exists(self, name):
        return self.get_load_balancer(name) is not None
    
    def get_load_balancer_attributes(self, arn):
        response = self.elbv2_client.describe_load_balancer_attributes(LoadBalancerArn=arn)
        return response['Attributes']
    
    def get_listener(self, arn):
        try:
            response = self.elbv2_client.describe_listeners(ListenerArns=[arn])
            return response['Listeners'][0]
        except Exception:
            return None
        
    def listener_exists(self, arn):
        return self.get_listener(arn) is not None
    
    def get_target_group(self, name):
        try:
            response = self.elbv2_client.describe_target_groups(Names=[name])
            return response['TargetGroups'][0]
        except Exception:
            return None
        
    def target_group_exists(self, name):
        return self.get_target_group(name) is not None
    

