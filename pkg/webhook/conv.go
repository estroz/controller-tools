/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	admissionreg "k8s.io/api/admissionregistration/v1"
	admissionreglegacy "k8s.io/api/admissionregistration/v1beta1"
)

// The functions in this file are required to convert v1 (superset) to v1beta1 webhook configs
// since the parent package of both versions does not contain conversion funcs, like for CRDs.

func v1Tov1beta1MutatingWebhooks(v1webhooks ...admissionreg.MutatingWebhook) (v1beta1webhooks []admissionreglegacy.MutatingWebhook) {
	for _, v1webhook := range v1webhooks {
		v1beta1webhooks = append(v1beta1webhooks, admissionreglegacy.MutatingWebhook{
			Name:          v1webhook.Name,
			ClientConfig:  v1Tov1beta1ClientConfig(v1webhook.ClientConfig),
			FailurePolicy: v1Tov1beta1FailurePolicy(v1webhook.FailurePolicy),
			MatchPolicy:   v1Tov1beta1MatchPolicy(v1webhook.MatchPolicy),
			SideEffects:   v1Tov1beta1SideEffects(v1webhook.SideEffects),
			Rules:         v1Tov1beta1Rules(v1webhook.Rules),
		})
	}
	return v1beta1webhooks
}

func v1Tov1beta1ValidatingWebhooks(v1webhooks ...admissionreg.ValidatingWebhook) (v1beta1webhooks []admissionreglegacy.ValidatingWebhook) {
	for _, v1webhook := range v1webhooks {
		v1beta1webhooks = append(v1beta1webhooks, admissionreglegacy.ValidatingWebhook{
			Name:          v1webhook.Name,
			ClientConfig:  v1Tov1beta1ClientConfig(v1webhook.ClientConfig),
			FailurePolicy: v1Tov1beta1FailurePolicy(v1webhook.FailurePolicy),
			MatchPolicy:   v1Tov1beta1MatchPolicy(v1webhook.MatchPolicy),
			SideEffects:   v1Tov1beta1SideEffects(v1webhook.SideEffects),
			Rules:         v1Tov1beta1Rules(v1webhook.Rules),
		})
	}
	return v1beta1webhooks
}

func v1Tov1beta1ClientConfig(v1ClientConfig admissionreg.WebhookClientConfig) (v1beta1ClientConfig admissionreglegacy.WebhookClientConfig) {
	v1beta1ClientConfig.Service = &admissionreglegacy.ServiceReference{
		Name:      v1ClientConfig.Service.Name,
		Namespace: v1ClientConfig.Service.Namespace,
		Path:      v1ClientConfig.Service.Path,
	}
	v1beta1ClientConfig.CABundle = v1ClientConfig.CABundle
	return v1beta1ClientConfig
}

func v1Tov1beta1FailurePolicy(v1FailurePolicy *admissionreg.FailurePolicyType) (v1beta1FailurePolicy *admissionreglegacy.FailurePolicyType) {
	if v1FailurePolicy == nil {
		return nil
	}
	fp := admissionreglegacy.FailurePolicyType(*v1FailurePolicy)
	return &fp
}

func v1Tov1beta1MatchPolicy(v1MatchPolicy *admissionreg.MatchPolicyType) (v1beta1MatchPolicy *admissionreglegacy.MatchPolicyType) {
	if v1MatchPolicy == nil {
		return nil
	}
	mp := admissionreglegacy.MatchPolicyType(*v1MatchPolicy)
	return &mp
}

func v1Tov1beta1SideEffects(v1SideEffects *admissionreg.SideEffectClass) (v1beta1SideEffects *admissionreglegacy.SideEffectClass) {
	if v1SideEffects == nil {
		return nil
	}
	se := admissionreglegacy.SideEffectClass(*v1SideEffects)
	return &se
}

func v1Tov1beta1Rules(v1Rules []admissionreg.RuleWithOperations) []admissionreglegacy.RuleWithOperations {
	v1beta1Rules := make([]admissionreglegacy.RuleWithOperations, len(v1Rules))
	for i, v1Rule := range v1Rules {
		v1beta1Operations := make([]admissionreglegacy.OperationType, len(v1Rule.Operations))
		for j, v1Operation := range v1Rule.Operations {
			v1beta1Operations[j] = admissionreglegacy.OperationType(v1Operation)
		}
		v1beta1Rules[i] = admissionreglegacy.RuleWithOperations{
			Rule: admissionreglegacy.Rule{
				APIGroups:   v1Rule.Rule.APIGroups,
				APIVersions: v1Rule.Rule.APIVersions,
				Resources:   v1Rule.Rule.Resources,
			},
			Operations: v1beta1Operations,
		}
	}
	return v1beta1Rules
}
