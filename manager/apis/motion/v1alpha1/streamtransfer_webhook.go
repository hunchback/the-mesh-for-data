// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *StreamTransfer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-motion-m4d-ibm-com-v1alpha1-streamtransfer,mutating=true,failurePolicy=fail,groups=motion.m4d.ibm.com,resources=streamtransfers,verbs=create;update,versions=v1alpha1,name=mstreamtransfer.kb.io

var _ webhook.Defaulter = &StreamTransfer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *StreamTransfer) Default() {
	log.Printf("Defaulting streamtransfer %s", r.Name)
	if r.Spec.Image == "" {
		if env, b := os.LookupEnv("MOVER_IMAGE"); b {
			r.Spec.Image = env
		} else {
			hostname, b1 := os.LookupEnv("DOCKER_HOSTNAME")
			namespace, b2 := os.LookupEnv("DOCKER_NAMESPACE")
			tagname, b3 := os.LookupEnv("DOCKER_TAGNAME")
			if b1 && b2 && b3 {
				r.Spec.Image = hostname + "/" + namespace + "/mover:" + tagname
			}
		}
	}

	if r.Spec.ImagePullPolicy == "" {
		if env, b := os.LookupEnv("IMAGE_PULL_POLICY"); b {
			r.Spec.ImagePullPolicy = v1.PullPolicy(env)
		} else {
			r.Spec.ImagePullPolicy = v1.PullIfNotPresent
		}
	}

	if r.Spec.SecretProviderURL == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_URL"); b {
			r.Spec.SecretProviderURL = env
		}
	}

	if r.Spec.SecretProviderRole == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_ROLE"); b {
			r.Spec.SecretProviderRole = env
		}
	}

	if r.Spec.TriggerInterval == "" {
		r.Spec.TriggerInterval = "5 seconds"
	}

	defaultDataStoreDescription(&r.Spec.Source)
	defaultDataStoreDescription(&r.Spec.Destination)

	if r.Spec.WriteOperation == "" {
		r.Spec.WriteOperation = Append
	}

	if r.Spec.DataFlowType == "" {
		r.Spec.DataFlowType = Stream
	}

	if r.Spec.ReadDataType == "" {
		r.Spec.ReadDataType = ChangeData
	}

	if r.Spec.WriteDataType == "" {
		r.Spec.WriteDataType = LogData
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-motion-m4d-ibm-com-v1alpha1-streamtransfer,mutating=false,failurePolicy=fail,groups=motion.m4d.ibm.com,resources=streamtransfers,versions=v1alpha1,name=vstreamtransfer.kb.io

var _ webhook.Validator = &StreamTransfer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *StreamTransfer) ValidateCreate() error {
	log.Printf("Validating streamtransfer %s for creation", r.Name)

	return r.validateStreamTransfer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *StreamTransfer) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating streamtransfer %s for update", r.Name)

	return r.validateStreamTransfer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *StreamTransfer) ValidateDelete() error {
	log.Printf("Validating streamtransfer %s for deletion", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *StreamTransfer) validateStreamTransfer() error {
	var allErrs field.ErrorList
	specField := field.NewPath("spec")

	if err := validateObjectName(&r.ObjectMeta); err != nil {
		allErrs = append(allErrs, err)
	}

	if r.Spec.DataFlowType == Batch {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("dataFlowType"), r.Spec.DataFlowType, "'dataFlowType' must be 'Stream' for a StreamTransfer!"))
	}

	if err := validateDataStore(specField.Child("source"), &r.Spec.Source); err != nil {
		allErrs = append(allErrs, err...)
	}
	if err := validateDataStore(specField.Child("destination"), &r.Spec.Destination); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "motion.m4d.ibm.com", Kind: "BatchTransfer"},
		r.Name, allErrs)
}
