// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

type ActionTypeEnum string

const (
	ActionTypeEnum_authenticate_cognito ActionTypeEnum = "authenticate-cognito"
	ActionTypeEnum_authenticate_oidc    ActionTypeEnum = "authenticate-oidc"
	ActionTypeEnum_fixed_response       ActionTypeEnum = "fixed-response"
	ActionTypeEnum_forward              ActionTypeEnum = "forward"
	ActionTypeEnum_redirect             ActionTypeEnum = "redirect"
)

type AdvertiseTrustStoreCaNamesEnum string

const (
	AdvertiseTrustStoreCaNamesEnum_off AdvertiseTrustStoreCaNamesEnum = "off"
	AdvertiseTrustStoreCaNamesEnum_on  AdvertiseTrustStoreCaNamesEnum = "on"
)

type AnomalyResultEnum string

const (
	AnomalyResultEnum_anomalous AnomalyResultEnum = "anomalous"
	AnomalyResultEnum_normal    AnomalyResultEnum = "normal"
)

type AuthenticateCognitoActionConditionalBehaviorEnum string

const (
	AuthenticateCognitoActionConditionalBehaviorEnum_allow        AuthenticateCognitoActionConditionalBehaviorEnum = "allow"
	AuthenticateCognitoActionConditionalBehaviorEnum_authenticate AuthenticateCognitoActionConditionalBehaviorEnum = "authenticate"
	AuthenticateCognitoActionConditionalBehaviorEnum_deny         AuthenticateCognitoActionConditionalBehaviorEnum = "deny"
)

type AuthenticateOIDCActionConditionalBehaviorEnum string

const (
	AuthenticateOIDCActionConditionalBehaviorEnum_allow        AuthenticateOIDCActionConditionalBehaviorEnum = "allow"
	AuthenticateOIDCActionConditionalBehaviorEnum_authenticate AuthenticateOIDCActionConditionalBehaviorEnum = "authenticate"
	AuthenticateOIDCActionConditionalBehaviorEnum_deny         AuthenticateOIDCActionConditionalBehaviorEnum = "deny"
)

type CapacityReservationStateEnum string

const (
	CapacityReservationStateEnum_failed      CapacityReservationStateEnum = "failed"
	CapacityReservationStateEnum_pending     CapacityReservationStateEnum = "pending"
	CapacityReservationStateEnum_provisioned CapacityReservationStateEnum = "provisioned"
	CapacityReservationStateEnum_rebalancing CapacityReservationStateEnum = "rebalancing"
)

type DescribeTargetHealthInputIncludeEnum string

const (
	DescribeTargetHealthInputIncludeEnum_All              DescribeTargetHealthInputIncludeEnum = "All"
	DescribeTargetHealthInputIncludeEnum_AnomalyDetection DescribeTargetHealthInputIncludeEnum = "AnomalyDetection"
)

type EnablePrefixForIPv6SourceNATEnum string

const (
	EnablePrefixForIPv6SourceNATEnum_off EnablePrefixForIPv6SourceNATEnum = "off"
	EnablePrefixForIPv6SourceNATEnum_on  EnablePrefixForIPv6SourceNATEnum = "on"
)

type EnforceSecurityGroupInboundRulesOnPrivateLinkTrafficEnum string

const (
	EnforceSecurityGroupInboundRulesOnPrivateLinkTrafficEnum_off EnforceSecurityGroupInboundRulesOnPrivateLinkTrafficEnum = "off"
	EnforceSecurityGroupInboundRulesOnPrivateLinkTrafficEnum_on  EnforceSecurityGroupInboundRulesOnPrivateLinkTrafficEnum = "on"
)

type IPAddressType string

const (
	IPAddressType_dualstack                     IPAddressType = "dualstack"
	IPAddressType_dualstack_without_public_ipv4 IPAddressType = "dualstack-without-public-ipv4"
	IPAddressType_ipv4                          IPAddressType = "ipv4"
)

type LoadBalancerSchemeEnum string

const (
	LoadBalancerSchemeEnum_internal        LoadBalancerSchemeEnum = "internal"
	LoadBalancerSchemeEnum_internet_facing LoadBalancerSchemeEnum = "internet-facing"
)

type LoadBalancerStateEnum string

const (
	LoadBalancerStateEnum_active          LoadBalancerStateEnum = "active"
	LoadBalancerStateEnum_active_impaired LoadBalancerStateEnum = "active_impaired"
	LoadBalancerStateEnum_failed          LoadBalancerStateEnum = "failed"
	LoadBalancerStateEnum_provisioning    LoadBalancerStateEnum = "provisioning"
)

type LoadBalancerTypeEnum string

const (
	LoadBalancerTypeEnum_application LoadBalancerTypeEnum = "application"
	LoadBalancerTypeEnum_gateway     LoadBalancerTypeEnum = "gateway"
	LoadBalancerTypeEnum_network     LoadBalancerTypeEnum = "network"
)

type MitigationInEffectEnum string

const (
	MitigationInEffectEnum_no  MitigationInEffectEnum = "no"
	MitigationInEffectEnum_yes MitigationInEffectEnum = "yes"
)

type ProtocolEnum string

const (
	ProtocolEnum_GENEVE  ProtocolEnum = "GENEVE"
	ProtocolEnum_HTTP    ProtocolEnum = "HTTP"
	ProtocolEnum_HTTPS   ProtocolEnum = "HTTPS"
	ProtocolEnum_TCP     ProtocolEnum = "TCP"
	ProtocolEnum_TCP_UDP ProtocolEnum = "TCP_UDP"
	ProtocolEnum_TLS     ProtocolEnum = "TLS"
	ProtocolEnum_UDP     ProtocolEnum = "UDP"
)

type RedirectActionStatusCodeEnum string

const (
	RedirectActionStatusCodeEnum_HTTP_301 RedirectActionStatusCodeEnum = "HTTP_301"
	RedirectActionStatusCodeEnum_HTTP_302 RedirectActionStatusCodeEnum = "HTTP_302"
)

type RevocationType string

const (
	RevocationType_CRL RevocationType = "CRL"
)

type TargetAdministrativeOverrideReasonEnum string

const (
	TargetAdministrativeOverrideReasonEnum_AdministrativeOverride_NoOverride               TargetAdministrativeOverrideReasonEnum = "AdministrativeOverride.NoOverride"
	TargetAdministrativeOverrideReasonEnum_AdministrativeOverride_Unknown                  TargetAdministrativeOverrideReasonEnum = "AdministrativeOverride.Unknown"
	TargetAdministrativeOverrideReasonEnum_AdministrativeOverride_ZonalShiftActive         TargetAdministrativeOverrideReasonEnum = "AdministrativeOverride.ZonalShiftActive"
	TargetAdministrativeOverrideReasonEnum_AdministrativeOverride_ZonalShiftDelegatedToDns TargetAdministrativeOverrideReasonEnum = "AdministrativeOverride.ZonalShiftDelegatedToDns"
)

type TargetAdministrativeOverrideStateEnum string

const (
	TargetAdministrativeOverrideStateEnum_no_override                  TargetAdministrativeOverrideStateEnum = "no_override"
	TargetAdministrativeOverrideStateEnum_unknown                      TargetAdministrativeOverrideStateEnum = "unknown"
	TargetAdministrativeOverrideStateEnum_zonal_shift_active           TargetAdministrativeOverrideStateEnum = "zonal_shift_active"
	TargetAdministrativeOverrideStateEnum_zonal_shift_delegated_to_dns TargetAdministrativeOverrideStateEnum = "zonal_shift_delegated_to_dns"
)

type TargetGroupIPAddressTypeEnum string

const (
	TargetGroupIPAddressTypeEnum_ipv4 TargetGroupIPAddressTypeEnum = "ipv4"
	TargetGroupIPAddressTypeEnum_ipv6 TargetGroupIPAddressTypeEnum = "ipv6"
)

type TargetHealthReasonEnum string

const (
	TargetHealthReasonEnum_Elb_InitialHealthChecking       TargetHealthReasonEnum = "Elb.InitialHealthChecking"
	TargetHealthReasonEnum_Elb_InternalError               TargetHealthReasonEnum = "Elb.InternalError"
	TargetHealthReasonEnum_Elb_RegistrationInProgress      TargetHealthReasonEnum = "Elb.RegistrationInProgress"
	TargetHealthReasonEnum_Target_DeregistrationInProgress TargetHealthReasonEnum = "Target.DeregistrationInProgress"
	TargetHealthReasonEnum_Target_FailedHealthChecks       TargetHealthReasonEnum = "Target.FailedHealthChecks"
	TargetHealthReasonEnum_Target_HealthCheckDisabled      TargetHealthReasonEnum = "Target.HealthCheckDisabled"
	TargetHealthReasonEnum_Target_InvalidState             TargetHealthReasonEnum = "Target.InvalidState"
	TargetHealthReasonEnum_Target_IpUnusable               TargetHealthReasonEnum = "Target.IpUnusable"
	TargetHealthReasonEnum_Target_NotInUse                 TargetHealthReasonEnum = "Target.NotInUse"
	TargetHealthReasonEnum_Target_NotRegistered            TargetHealthReasonEnum = "Target.NotRegistered"
	TargetHealthReasonEnum_Target_ResponseCodeMismatch     TargetHealthReasonEnum = "Target.ResponseCodeMismatch"
	TargetHealthReasonEnum_Target_Timeout                  TargetHealthReasonEnum = "Target.Timeout"
)

type TargetHealthStateEnum string

const (
	TargetHealthStateEnum_draining           TargetHealthStateEnum = "draining"
	TargetHealthStateEnum_healthy            TargetHealthStateEnum = "healthy"
	TargetHealthStateEnum_initial            TargetHealthStateEnum = "initial"
	TargetHealthStateEnum_unavailable        TargetHealthStateEnum = "unavailable"
	TargetHealthStateEnum_unhealthy          TargetHealthStateEnum = "unhealthy"
	TargetHealthStateEnum_unhealthy_draining TargetHealthStateEnum = "unhealthy.draining"
	TargetHealthStateEnum_unused             TargetHealthStateEnum = "unused"
)

type TargetTypeEnum string

const (
	TargetTypeEnum_alb      TargetTypeEnum = "alb"
	TargetTypeEnum_instance TargetTypeEnum = "instance"
	TargetTypeEnum_ip       TargetTypeEnum = "ip"
	TargetTypeEnum_lambda   TargetTypeEnum = "lambda"
)

type TrustStoreAssociationStatusEnum string

const (
	TrustStoreAssociationStatusEnum_active  TrustStoreAssociationStatusEnum = "active"
	TrustStoreAssociationStatusEnum_removed TrustStoreAssociationStatusEnum = "removed"
)

type TrustStoreStatus string

const (
	TrustStoreStatus_ACTIVE   TrustStoreStatus = "ACTIVE"
	TrustStoreStatus_CREATING TrustStoreStatus = "CREATING"
)
