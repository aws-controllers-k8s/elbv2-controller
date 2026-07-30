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

"""Integration tests for the ELB Listener API.
"""

import logging
import time

import pytest
from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import CRD_GROUP, CRD_VERSION, load_elbv2_resource, service_marker
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import ELBValidator

from .test_load_balancer import simple_load_balancer

RESOURCE_PLURAL = "listeners"
TARGET_GROUP_PLURAL = "targetgroups"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 20
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def simple_listener(elbv2_client, simple_load_balancer):
    (lb_ref, lb_cr, _) = simple_load_balancer

    resource_name = random_suffix_name("listener", 16)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["LISTENER_NAME"] = resource_name
    replacements["LOAD_BALANCER_ARN"] = lb_cr["status"]["ackResourceMetadata"]["arn"]

    resource_data = load_elbv2_resource(
        "listener",
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
    assert not validator.listener_exists(cr["status"]["ackResourceMetadata"]["arn"])

@service_marker
@pytest.mark.canary
class TestListener:
    def test_create_delete(self, elbv2_client, simple_listener):
        (ref, cr) = simple_listener
        assert cr is not None
        listener_arn = cr["status"]["ackResourceMetadata"]["arn"]

        validator = ELBValidator(elbv2_client)
        assert validator.listener_exists(listener_arn)

        # Update settings
        updates = {
            "spec": {
                "port": 9000,
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        listener = validator.get_listener(listener_arn)
        assert listener is not None
        assert listener["Port"] == 9000


# ---------------------------------------------------------------------------
# services.k8s.aws/ignore-field-drift coverage
# ---------------------------------------------------------------------------
#
# The IgnoreFieldDrift runtime feature (aws-controllers-k8s/runtime#256) lets a
# resource opt specific spec paths out of drift reconciliation via the
# services.k8s.aws/ignore-field-drift annotation. This mirrors the motivating
# use case in aws-controllers-k8s/elbv2-controller#85 "Scenario 2": a blue/green
# deploy tool shifts traffic weights across target groups on the live listener,
# and without ignore-field-drift the controller reconciles the weights back to
# the declared spec, breaking the deployment.
#
# The feature gate is Alpha and disabled by default, so the test enables it on
# the deployed controller for the duration of the module and restores the prior
# value afterwards. This requires a controller built from a runtime that carries
# the gate; against a runtime without it, enabling an unknown gate is fatal, so
# this coverage runs green only once the controller's runtime dependency
# includes the feature.

# Controller deployment coordinates in the kind test cluster (see
# test-infra/scripts/controller-setup.sh and the controller Helm chart, which
# wires FEATURE_GATES into the --feature-gates flag).
CONTROLLER_NAMESPACE = "ack-system"
CONTROLLER_DEPLOYMENT = "ack-elbv2-controller"
CONTROLLER_CONTAINER = "controller"
FEATURE_GATE = "IgnoreFieldDrift"
# Generous window for the new pod to roll out and take over reconciliation.
ROLLOUT_WAIT_SECONDS = 120

# The declared (spec) weights and the externally-shifted weights. The two must
# differ so a revert would be observable.
DECLARED_WEIGHT_1 = 90
DECLARED_WEIGHT_2 = 10
EXTERNAL_WEIGHT_1 = 50
EXTERNAL_WEIGHT_2 = 50


def _apps_client():
    # Build the AppsV1Api against acktest's configured ApiClient (which points
    # at the kind cluster). A bare AppsV1Api() would default to localhost:80.
    from kubernetes import client as k8s_client
    return k8s_client.AppsV1Api(k8s._get_k8s_api_client())


def _get_feature_gates_env() -> str:
    """Returns the current value of the FEATURE_GATES env var on the controller
    container, or "" if it is unset."""
    dep = _apps_client().read_namespaced_deployment(
        CONTROLLER_DEPLOYMENT, CONTROLLER_NAMESPACE,
    )
    for c in dep.spec.template.spec.containers:
        if c.name != CONTROLLER_CONTAINER:
            continue
        for e in (c.env or []):
            if e.name == "FEATURE_GATES":
                return e.value or ""
    return ""


def _set_feature_gates_env(value: str):
    """Patches the FEATURE_GATES env var on the controller container and waits
    for the rollout to complete. The controller wires this env var into its
    --feature-gates flag (see the controller Helm chart)."""
    body = {
        "spec": {
            "template": {
                "spec": {
                    "containers": [
                        {"name": CONTROLLER_CONTAINER,
                         "env": [{"name": "FEATURE_GATES", "value": value}]},
                    ]
                }
            }
        }
    }
    _apps_client().patch_namespaced_deployment(
        CONTROLLER_DEPLOYMENT, CONTROLLER_NAMESPACE, body,
    )
    _wait_for_rollout()


def _merge_gate(existing: str, gate: str, enabled: bool) -> str:
    """Returns a FEATURE_GATES string with `gate` set to `enabled`, preserving
    any other gates already present."""
    pairs = {}
    for part in filter(None, (p.strip() for p in existing.split(","))):
        if "=" in part:
            k, v = part.split("=", 1)
            pairs[k.strip()] = v.strip()
    pairs[gate] = "true" if enabled else "false"
    return ",".join(f"{k}={v}" for k, v in pairs.items())


def _wait_for_rollout():
    """Blocks until the controller deployment reports all replicas updated and
    available for the current generation."""
    client = _apps_client()
    deadline = time.time() + ROLLOUT_WAIT_SECONDS
    while time.time() < deadline:
        dep = client.read_namespaced_deployment(
            CONTROLLER_DEPLOYMENT, CONTROLLER_NAMESPACE,
        )
        spec_replicas = dep.spec.replicas or 1
        status = dep.status
        if (status.observed_generation is not None
                and status.observed_generation >= dep.metadata.generation
                and (status.updated_replicas or 0) >= spec_replicas
                and (status.available_replicas or 0) >= spec_replicas
                and (status.unavailable_replicas or 0) == 0):
            # Give the fresh pod a moment to acquire leadership / start reconciling.
            time.sleep(5)
            return
        time.sleep(3)
    raise AssertionError(
        f"controller deployment {CONTROLLER_DEPLOYMENT} did not roll out within "
        f"{ROLLOUT_WAIT_SECONDS}s after toggling the {FEATURE_GATE} feature gate"
    )


def _weights_by_tg_arn(listener: dict) -> dict:
    """Returns {target_group_arn: weight} from a describe_listeners entry's
    first default forward action."""
    actions = listener.get("DefaultActions", [])
    forward = next(
        (a for a in actions if a.get("Type") == "forward"), None,
    )
    assert forward is not None, "listener has no forward default action"
    tgs = forward["ForwardConfig"]["TargetGroups"]
    return {tg["TargetGroupArn"]: tg["Weight"] for tg in tgs}


@pytest.fixture(scope="module")
def ignore_field_drift_enabled():
    """Enables the IgnoreFieldDrift feature gate on the controller for the
    duration of the module, then restores the prior FEATURE_GATES value."""
    original = _get_feature_gates_env()
    _set_feature_gates_env(_merge_gate(original, FEATURE_GATE, True))
    yield
    # Restore exactly what was there before (which may be "").
    _set_feature_gates_env(original)


@pytest.fixture(scope="module")
def two_target_groups(elbv2_client):
    """Creates two ip-type target groups (in the bootstrapped VPC) for the
    weighted forward action. ip-type is used instead of lambda so the fixture
    does not depend on registered targets -- the weight-drift scenario needs
    only the target groups themselves."""
    refs = []
    names = []
    for i in range(2):
        name = random_suffix_name(f"tg-ifd-{i+1}", 24)
        replacements = REPLACEMENT_VALUES.copy()
        replacements["TARGET_GROUP_NAME"] = name
        data = load_elbv2_resource(
            "target_group_ip", additional_replacements=replacements,
        )
        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, TARGET_GROUP_PLURAL, name, namespace="default",
        )
        k8s.create_custom_resource(ref, data)
        refs.append(ref)
        names.append(name)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    for ref in refs:
        cr = k8s.wait_resource_consumed_by_controller(ref)
        assert cr is not None

    yield names

    for ref in refs:
        try:
            _, deleted = k8s.delete_custom_resource(ref, 3, DELETE_WAIT_AFTER_SECONDS)
            assert deleted
        except Exception:
            pass


@pytest.fixture
def ignore_field_drift_listener(request, elbv2_client, simple_load_balancer, two_target_groups):
    """A Listener with a weighted forward action across two target groups,
    annotated to ignore drift on spec.defaultActions.

    Parametrize the ignored paths via an indirect fixture param, e.g.:

        @pytest.mark.parametrize(
            "ignore_field_drift_listener",
            [{"ignore_paths": "spec.defaultActions"}],
            indirect=True,
        )

    Defaults to ignoring spec.defaultActions so callers that don't parametrize
    keep the weight-drift behaviour."""
    (lb_ref, lb_cr, _) = simple_load_balancer
    param = getattr(request, "param", None) or {}
    ignore_paths = param.get("ignore_paths", "spec.defaultActions")

    resource_name = random_suffix_name("listener-ifd", 24)
    replacements = REPLACEMENT_VALUES.copy()
    replacements["LISTENER_NAME"] = resource_name
    replacements["LOAD_BALANCER_ARN"] = lb_cr["status"]["ackResourceMetadata"]["arn"]
    replacements["TARGET_GROUP_NAME_1"] = two_target_groups[0]
    replacements["TARGET_GROUP_NAME_2"] = two_target_groups[1]
    replacements["WEIGHT_1"] = str(DECLARED_WEIGHT_1)
    replacements["WEIGHT_2"] = str(DECLARED_WEIGHT_2)
    replacements["IGNORE_PATHS"] = ignore_paths

    resource_data = load_elbv2_resource(
        "listener_ignore_field_drift",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

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

    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, DELETE_WAIT_AFTER_SECONDS)
        assert deleted
    except Exception:
        pass


@service_marker
class TestListenerIgnoreFieldDrift:
    """Verifies the services.k8s.aws/ignore-field-drift annotation on an ELBv2
    Listener's forward-action target-group weights (elbv2#85 Scenario 2).

    The controller still applies the declared weights at create but stops
    reconciling drift on the ignored spec.defaultActions path: an externally
    shifted weight distribution survives, the resource stays Synced, and an edit
    to the ignored field is retained in the spec but not pushed to AWS."""

    def test_weight_drift_ignored(
        self, elbv2_client, ignore_field_drift_enabled, ignore_field_drift_listener,
    ):
        (ref, cr) = ignore_field_drift_listener
        listener_arn = cr["status"]["ackResourceMetadata"]["arn"]
        validator = ELBValidator(elbv2_client)

        # Baseline: the declared weights were applied at create, and the
        # resource is Synced.
        listener = validator.get_listener(listener_arn)
        assert listener is not None
        baseline = _weights_by_tg_arn(listener)
        assert sorted(baseline.values()) == sorted(
            [DECLARED_WEIGHT_1, DECLARED_WEIGHT_2]
        ), f"unexpected baseline weights: {baseline}"
        condition.assert_synced(ref)

        # Snapshot the live forward action, then flip the weights out-of-band
        # (the blue/green deploy tool shifting traffic).
        forward = next(
            a for a in listener["DefaultActions"] if a.get("Type") == "forward"
        )
        tgs = forward["ForwardConfig"]["TargetGroups"]
        assert len(tgs) == 2
        shifted_tgs = [
            {"TargetGroupArn": tgs[0]["TargetGroupArn"], "Weight": EXTERNAL_WEIGHT_1},
            {"TargetGroupArn": tgs[1]["TargetGroupArn"], "Weight": EXTERNAL_WEIGHT_2},
        ]
        elbv2_client.modify_listener(
            ListenerArn=listener_arn,
            DefaultActions=[
                {
                    "Type": "forward",
                    "ForwardConfig": {"TargetGroups": shifted_tgs},
                }
            ],
        )
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # The externally-shifted weights must survive: the controller does not
        # reconcile drift on spec.defaultActions, so it does not call
        # ModifyListener to revert them.
        after = _weights_by_tg_arn(validator.get_listener(listener_arn))
        assert after == {
            tgs[0]["TargetGroupArn"]: EXTERNAL_WEIGHT_1,
            tgs[1]["TargetGroupArn"]: EXTERNAL_WEIGHT_2,
        }, (
            "controller reverted externally-shifted listener weights despite "
            f"ignore-field-drift on spec.defaultActions: {after}"
        )

        # The resource stays Synced even though the live weights (50/50) differ
        # from the declared spec (90/10).
        assert k8s.wait_on_condition(
            ref, "ACK.ResourceSynced", "True",
            wait_periods=6, period_length=10,
        )

        # Editing the ignored field in the spec is retained but NOT pushed to
        # AWS: patch the declared weights to a third value and confirm the live
        # weights are unchanged (still the external 50/50).
        latest = k8s.get_resource(ref)
        new_actions = latest["spec"]["defaultActions"]
        new_actions[0]["forwardConfig"]["targetGroups"][0]["weight"] = 70
        new_actions[0]["forwardConfig"]["targetGroups"][1]["weight"] = 30
        k8s.patch_custom_resource(ref, {"spec": {"defaultActions": new_actions}})
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        after_edit = _weights_by_tg_arn(validator.get_listener(listener_arn))
        assert after_edit == {
            tgs[0]["TargetGroupArn"]: EXTERNAL_WEIGHT_1,
            tgs[1]["TargetGroupArn"]: EXTERNAL_WEIGHT_2,
        }, (
            "controller pushed a spec edit on an ignored field to AWS: "
            f"{after_edit}"
        )

        # The declared edit is retained in the CR spec (retain semantics).
        latest = k8s.get_resource(ref)
        spec_weights = sorted(
            tg["weight"]
            for tg in latest["spec"]["defaultActions"][0]["forwardConfig"]["targetGroups"]
        )
        assert spec_weights == [30, 70], f"spec did not retain the edit: {spec_weights}"
