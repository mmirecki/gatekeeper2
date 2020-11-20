package webhook

import (
	"context"
	"testing"

	"github.com/ghodss/yaml"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	atypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	validAssignMeta = `
apiVersion: mutations.gatekeeper.sh
kind: AssignMetadata
metadata:
  name: testAssignMeta
spec:
  location: metadata.labels.foo
  parameters:
    assign:
      value: bar  
`
	assignMetaInvalidPath = `
apiVersion: mutations.gatekeeper.sh
kind: AssignMetadata
metadata:
  name: testAssignMeta
spec:
  location: metadata.foo.bar
  parameters:
    assign:
      value: bar  
`
	assignMetaInvalidAssign = `
apiVersion: mutations.gatekeeper.sh
kind: AssignMetadata
metadata:
  name: testAssignMeta
spec:
  location: metadata.labels.bar
  parameters:
    assign:
      foo: bar  
`
	assignMetaAssignNotString = `
apiVersion: mutations.gatekeeper.sh
kind: AssignMetadata
metadata:
  name: testAssignMeta
spec:
  location: metadata.labels.bar
  parameters:
    assign:
      value:
        foo: bar  
`
	assignMetaNoValue = `
apiVersion: mutations.gatekeeper.sh
kind: AssignMetadata
metadata:
  name: testAssignMeta
spec:
  location: metadata.labels.bar
  parameters:
    assign:
      zzz:
        foo: bar  
`
	validAssign = `
apiVersion: mutations.gatekeeper.sh
kind: Assign
metadata:
  name: goodAssign
spec:
  location: "spec.containers[name:test].foo"
  parameters:
    assign:
      value: bar
`
	invalidAssignChangesMetadata = `
apiVersion: mutations.gatekeeper.sh
kind: Assign
metadata:
  name: assignExample
spec:
  location: metadata.foo.bar
  parameters:
    assign:
      value: bar  
`
	invalidAssignNoValue = `
apiVersion: mutations.gatekeeper.sh
kind: Assign
metadata:
  name: assignExample
spec:
  location: spec.containers
  parameters:
    assign:
      zzz: bar  
`
	invalidAssignNoAssign = `
apiVersion: mutations.gatekeeper.sh
kind: Assign
metadata:
  name: assignExample
spec:
  location: spec.containers
`
)

func TestAssignMetaValidation(t *testing.T) {
	tc := []struct {
		Name          string
		AssignMeta    string
		ErrorExpected bool
	}{
		{
			Name:          "Valid Assign",
			AssignMeta:    validAssignMeta,
			ErrorExpected: false,
		},
		{
			Name:          "Invalid Path",
			AssignMeta:    assignMetaInvalidPath,
			ErrorExpected: true,
		},
		{
			Name:          "Invalid Assign",
			AssignMeta:    assignMetaInvalidAssign,
			ErrorExpected: true,
		},
		{
			Name:          "Assign not a string",
			AssignMeta:    assignMetaAssignNotString,
			ErrorExpected: true,
		},
		{
			Name:          "Assign no value",
			AssignMeta:    assignMetaNoValue,
			ErrorExpected: true,
		},
	}
	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			handler := mutationHandler{webhookHandler: webhookHandler{}}
			b, err := yaml.YAMLToJSON([]byte(tt.AssignMeta))
			if err != nil {
				t.Fatalf("Error parsing yaml: %s", err)
			}
			review := atypes.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Group:   "mutations.gatekeeper.sh",
						Version: "v1alpha1",
						Kind:    "AssignMetadata",
					},
					Object: runtime.RawExtension{
						Raw: b,
					},
				},
			}
			_, err = handler.validateGatekeeperResources(context.Background(), review)
			if err != nil && !tt.ErrorExpected {
				t.Errorf("err = %s; want nil", err)
			}
			if err == nil && tt.ErrorExpected {
				t.Error("err = nil; want non-nil")
			}
		})
	}
}

func TestAssignValidation(t *testing.T) {
	tc := []struct {
		Name          string
		Assign        string
		ErrorExpected bool
	}{
		{
			Name:          "Valid Assign",
			Assign:        validAssign,
			ErrorExpected: false,
		},
		{
			Name:          "Changes Metadata",
			Assign:        invalidAssignChangesMetadata,
			ErrorExpected: true,
		},
		{
			Name:          "No Value",
			Assign:        invalidAssignNoValue,
			ErrorExpected: true,
		},
		{
			Name:          "No Assign",
			Assign:        invalidAssignNoAssign,
			ErrorExpected: true,
		},
	}
	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			handler := mutationHandler{webhookHandler: webhookHandler{}}
			b, err := yaml.YAMLToJSON([]byte(tt.Assign))
			if err != nil {
				t.Fatalf("Error parsing yaml: %s", err)
			}
			review := atypes.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Group:   "mutations.gatekeeper.sh",
						Version: "v1alpha1",
						Kind:    "Assign",
					},
					Object: runtime.RawExtension{
						Raw: b,
					},
				},
			}
			_, err = handler.validateGatekeeperResources(context.Background(), review)
			if err != nil && !tt.ErrorExpected {
				t.Errorf("err = %s; want nil", err)
			}
			if err == nil && tt.ErrorExpected {
				t.Error("err = nil; want non-nil")
			}
		})
	}
}
