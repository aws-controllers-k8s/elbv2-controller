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
# the deployed controller. The gate must exist in the runtime the controller was
# built against -- enabling an unknown gate is fatal to the controller, which
# surfaces here as the rollout below never completing. It landed in runtime
# v0.62.0.
#
# Enabling the gate restarts the shared controller Deployment, which is a
# cluster-wide side effect in a suite pytest-xdist spreads across 16 workers: a
# restart mid-run can time out any other worker waiting on ACK.ResourceSynced.
# The gate is therefore turned on exactly once per run and never turned back
# off -- see ignore_field_drift_enabled for why that is both safe and necessary.

# Controller deployment coordinates in the kind test cluster (see
# test-infra/scripts/controller-setup.sh and the controller Helm chart, which
# wires FEATURE_GATES into the --feature-gates flag).
CONTROLLER_NAMESPACE = "ack-system"
CONTROLLER_DEPLOYMENT = "ack-elbv2-controller"
CONTROLLER_CONTAINER = "controller"
FEATURE_GATE = "IgnoreFieldDrift"
# Generous window for the new pod to roll out and take over reconciliation.
ROLLOUT_WAIT_SECONDS = 120

# How long to wait for a resource to reach ACK.ResourceSynced=True (120s).
SYNC_WAIT_PERIODS = 12
SYNC_PERIOD_LENGTH = 10

# An inert annotation patched onto the CR purely to trigger a reconcile. An
# out-of-band AWS change produces no watch event, and this controller's resync
# period is the runtime default of 10 hours -- config/controller/deployment.yaml
# (what the e2e job deploys via kustomize) passes no --reconcile-*-resync-seconds
# override and Listener's RequeueOnSuccessSeconds() is 0, so getResyncPeriod
# falls through to defaultResyncPeriod. Without an explicit nudge the controller
# would not look at the resource again for the rest of the run, and any
# assertion about what it did with the drift would be vacuous.
#
# Touching an annotation is enough because the runtime adds
# AnnotationChangedPredicate to the event filter whenever the IgnoreFieldDrift
# gate is on (runtime reconciler.go, SetupWithManager); the default filter is
# GenerationChangedPredicate alone, which an annotation edit would not satisfy.
# This keeps the probe off the spec, so the only delta the reconcile sees is the
# external drift itself.
RECONCILE_PROBE_ANNOTATION = "e2e.test.ack.aws.dev/reconcile-probe"

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


def _gate_enabled_in_spec() -> bool:
    """Returns True if the controller Deployment's FEATURE_GATES env var already
    has FEATURE_GATE turned on. Read from the Deployment spec (not the running
    pod), so it reflects a patch that has been accepted but is still rolling."""
    pairs = _parse_gates(_get_feature_gates_env())
    return pairs.get(FEATURE_GATE) == "true"


def _parse_gates(existing: str) -> dict:
    """Parses a FEATURE_GATES string ("A=true,B=false") into a dict."""
    pairs = {}
    for part in filter(None, (p.strip() for p in existing.split(","))):
        if "=" in part:
            k, v = part.split("=", 1)
            pairs[k.strip()] = v.strip()
    return pairs


def _merge_gate(existing: str, gate: str, enabled: bool) -> str:
    """Returns a FEATURE_GATES string with `gate` set to `enabled`, preserving
    any other gates already present."""
    pairs = _parse_gates(existing)
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


@pytest.fixture(scope="session")
def ignore_field_drift_enabled():
    """Turns the IgnoreFieldDrift feature gate on for the run, exactly once, and
    deliberately never turns it back off.

    Session scope is per WORKER, not per run: pytest-xdist distributes individual
    tests across 16 worker processes with LoadScheduling, so a fixture at any
    scope is instantiated once in every worker that happens to pick up a test
    from this file. An enable/restore pair therefore rolls the shared controller
    Deployment twice for every such worker, and each restart could time out an
    unrelated listener / load balancer / target group test waiting on
    ACK.ResourceSynced.

    Two properties keep this to a single restart:

      - Check-then-set. A worker that finds the gate already on in the
        Deployment spec skips the patch. Concurrent workers that both observe it
        off compute the same FEATURE_GATES string from the same starting value,
        so the second patch leaves the pod template byte-identical, does not bump
        the Deployment generation, and does not trigger a second rollout. That
        makes an explicit cross-process lock unnecessary.

      - No restore. Restoring would cost a second rollout, and a teardown in one
        worker would disable the gate underneath a drift test still running in
        another. Leaving it on is safe because the gate is inert unless a
        resource carries the ignore-field-drift annotation, which only this
        section's resources do; every other test in the suite behaves identically
        with it on. The kind cluster is torn down at the end of the run, so
        nothing outlives it.

    One rollout mid-run is still a shared-cluster hiccup. Removing it entirely
    means enabling the gate at controller setup time (FEATURE_GATES in
    test-infra's controller-setup.sh, as IAMRoleSelector already does), after
    which this fixture becomes a no-op check.
    """
    if not _gate_enabled_in_spec():
        _set_feature_gates_env(_merge_gate(_get_feature_gates_env(), FEATURE_GATE, True))
    # Whether we patched or another worker did, do not hand out the fixture
    # until the controller serving the gate is actually up.
    _wait_for_rollout()
    yield


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
        # wait_resource_consumed_by_controller returns as soon as the resource
        # has any .status at all, which is the first status write and predates
        # the ARN landing. The listener below resolves these target groups by
        # reference, so wait until each is actually reconciled rather than
        # merely observed -- otherwise the listener's own reference resolution
        # is still pending when its fixture reads status.
        assert k8s.wait_on_condition(
            ref, "ACK.ResourceSynced", "True",
            wait_periods=SYNC_WAIT_PERIODS, period_length=SYNC_PERIOD_LENGTH,
        ), f"target group {ref.name} never reached ACK.ResourceSynced=True"

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

    # Same trap as in two_target_groups, and worse here: the test reads the ARN
    # out of the CR this fixture yields, so a snapshot taken at the first status
    # write (before status.ackResourceMetadata exists) gives it a KeyError it
    # cannot recover from. Wait for a real reconcile, then re-read.
    assert k8s.wait_on_condition(
        ref, "ACK.ResourceSynced", "True",
        wait_periods=SYNC_WAIT_PERIODS, period_length=SYNC_PERIOD_LENGTH,
    ), f"listener {resource_name} never reached ACK.ResourceSynced=True"
    cr = k8s.get_resource(ref)
    assert cr["status"]["ackResourceMetadata"]["arn"]

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

        # Capture the synced timestamp before touching anything. ACK rewrites
        # ACK.ResourceSynced.lastTransitionTime on every reconcile, so this is
        # what lets the assertions below distinguish "the controller reconciled
        # and left the drift alone" from "the controller never looked".
        synced_before = condition.get_synced_last_transition_time(ref)
        assert synced_before is not None

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

        expected_external = {
            tgs[0]["TargetGroupArn"]: EXTERNAL_WEIGHT_1,
            tgs[1]["TargetGroupArn"]: EXTERNAL_WEIGHT_2,
        }

        # Precondition: AWS really reports the shifted distribution before the
        # controller is asked to look at it. If the reconcile forced below raced
        # ahead of ModifyListener taking effect, sdkFind would read the original
        # 90/10, find no drift, and the survival assertion would pass for a
        # reason that has nothing to do with the feature.
        shifted = _weights_by_tg_arn(validator.get_listener(listener_arn))
        assert shifted == expected_external, (
            f"out-of-band ModifyListener did not take effect: {shifted}"
        )

        # Force a reconcile. Nothing else will: the AWS-side change produced no
        # watch event and the next resync is 10 hours out (see
        # RECONCILE_PROBE_ANNOTATION). Then require that a reconcile which
        # started AFTER the drift completed with ACK.ResourceSynced=True. This
        # single assertion carries both claims -- that the controller looked, and
        # that it still considers the resource in sync despite the live weights
        # (50/50) differing from the declared spec (90/10). A plain
        # wait_on_condition here would return on its first poll off the condition
        # written back at create time and prove neither.
        k8s.patch_custom_resource(
            ref, {"metadata": {"annotations": {RECONCILE_PROBE_ANNOTATION: "1"}}},
        )
        assert k8s.wait_on_condition_after(
            ref, "ACK.ResourceSynced", "True",
            last_transition_after=synced_before,
            wait_periods=SYNC_WAIT_PERIODS, period_length=SYNC_PERIOD_LENGTH,
        ), (
            "no reconcile completed with ACK.ResourceSynced=True after the "
            "out-of-band weight shift"
        )

        # Now that a fresh reconcile is known to have run, this is meaningful:
        # the controller examined the resource and declined to call
        # ModifyListener to revert the ignored spec.defaultActions path.
        after = _weights_by_tg_arn(validator.get_listener(listener_arn))
        assert after == expected_external, (
            "controller reverted externally-shifted listener weights despite "
            f"ignore-field-drift on spec.defaultActions: {after}"
        )

        # Editing the ignored field in the spec is retained but NOT pushed to
        # AWS: patch the declared weights to a third value and confirm the live
        # weights are unchanged (still the external 50/50).
        synced_before_edit = condition.get_synced_last_transition_time(ref)
        assert synced_before_edit is not None

        latest = k8s.get_resource(ref)
        new_actions = latest["spec"]["defaultActions"]
        new_actions[0]["forwardConfig"]["targetGroups"][0]["weight"] = 70
        new_actions[0]["forwardConfig"]["targetGroups"][1]["weight"] = 30
        k8s.patch_custom_resource(ref, {"spec": {"defaultActions": new_actions}})

        # This patch bumps metadata.generation, so a reconcile is guaranteed to
        # be queued -- but not that it has finished. Wait for one that started
        # after the edit rather than assuming a fixed sleep covers it.
        assert k8s.wait_on_condition_after(
            ref, "ACK.ResourceSynced", "True",
            last_transition_after=synced_before_edit,
            wait_periods=SYNC_WAIT_PERIODS, period_length=SYNC_PERIOD_LENGTH,
        ), (
            "no reconcile completed with ACK.ResourceSynced=True after the "
            "spec edit to the ignored field"
        )

        after_edit = _weights_by_tg_arn(validator.get_listener(listener_arn))
        assert after_edit == expected_external, (
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
