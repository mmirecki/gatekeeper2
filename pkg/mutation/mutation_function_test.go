package mutation_test

import (
	"fmt"
	"testing"

	mutationsv1alpha1 "github.com/open-policy-agent/gatekeeper/apis/mutations/v1alpha1"
	"github.com/open-policy-agent/gatekeeper/pkg/mutation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const TestValue = "testValue"

func prepareTestPod(t *testing.T) *unstructured.Unstructured {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpod1",
			Namespace: "foo",
			Labels:    map[string]string{"a": "b"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "testname1",
					Ports: []corev1.ContainerPort{{Name: "portName1"}},
				},
				{
					Name: "testname2",
					Ports: []corev1.ContainerPort{
						{Name: "portName2A"},
						{Name: "portName2B"},
					},
				},
				{
					Name:  "testname3",
					Ports: []corev1.ContainerPort{{Name: "portName3"}},
				},
			},
		},
	}
	podObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
	if err != nil {
		t.Errorf("Failed to convert pod to unstructured %v", err)
	}
	return &unstructured.Unstructured{Object: podObject}
}

func TestObjects(t *testing.T) {

	testFunc := func(unstructured *unstructured.Unstructured) {
		labels := unstructured.Object["metadata"].(map[string]interface{})["labels"]
		if labels.(map[string]interface{})["labelA"] != TestValue {
			t.Errorf("Failed to update pod")
		}

	}

	testMutation(
		"metadata.labels.labelA",
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestObjectsAndLists(t *testing.T) {
	testFunc := func(unstructured *unstructured.Unstructured) {
		containers := unstructured.Object["spec"].(map[string]interface{})["containers"]
		for _, container := range containers.([]interface{}) {
			containerAsMap := container.(map[string]interface{})
			if containerAsMap["name"] == "testname2" {
				ports := containerAsMap["ports"]
				for _, port := range ports.([]interface{}) {
					portAsMap := port.(map[string]interface{})
					if portAsMap["name"] == "portName2B" {
						if portAsMap["hostIP"] != TestValue {
							t.Errorf("Failed to update pod")
						}
					} else {
						if _, ok := port.(map[string]interface{})["hostIP"]; ok {
							t.Errorf("Unexpected pod was updated")
						}
					}
				}
			} else {
				for _, port := range container.(map[string]interface{})["ports"].([]interface{}) {
					if _, ok := port.(map[string]interface{})["hostIP"]; ok {
						t.Errorf("Unexpected pod was updated")
					}
				}
			}
		}
	}

	testMutation(
		`spec.containers["name": "testname2"].ports["name": "portName2B"].hostIP`,
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestListsAsLastElement(t *testing.T) {
	testFunc := func(unstructured *unstructured.Unstructured) {
		containers := unstructured.Object["spec"].(map[string]interface{})["containers"]
		for _, container := range containers.([]interface{}) {
			if container.(map[string]interface{})["name"] == TestValue {
				return
			}
		}
		t.Errorf("Failed to update pod")
	}

	testMutation(
		`spec.containers["name": "testValue"]`,
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestListsAsLastElementAlreadyExists(t *testing.T) {
	testFunc := func(unstructured *unstructured.Unstructured) {
		containers := unstructured.Object["spec"].(map[string]interface{})["containers"]
		for _, container := range containers.([]interface{}) {
			if container.(map[string]interface{})["name"] == "testname1" {
				return
			}
		}
		t.Errorf("Expected value missing")
	}

	testMutation(
		`spec.containers["name": "testname1"]`,
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestGlobbedList(t *testing.T) {
	testFunc := func(unstructured *unstructured.Unstructured) {
		containers := unstructured.Object["spec"].(map[string]interface{})["containers"]
		for _, container := range containers.([]interface{}) {
			containerAsMap := container.(map[string]interface{})
			ports := containerAsMap["ports"]
			for _, port := range ports.([]interface{}) {
				if value, ok := port.(map[string]interface{})["protocol"]; !ok || value != TestValue {
					t.Errorf("Expected value was not updated")
				}
			}
		}
	}

	testMutation(
		`spec.containers["name": *].ports["name": *].protocol`,
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestNonExistingPathEntry(t *testing.T) {
	testFunc := func(unstructured *unstructured.Unstructured) {
		element := unstructured.Object["spec"].(map[string]interface{})["element"].(map[string]interface{})["should"].(map[string]interface{})["be"]
		if element.(map[string]interface{})["added"] != TestValue {
			t.Errorf("Failed to update pod")
		}
	}
	testMutation(
		"spec.element.should.be.added",
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestNonExistingListPathEntry(t *testing.T) {
	testFunc := func(unstructured *unstructured.Unstructured) {
		element := unstructured.Object["spec"].(map[string]interface{})["element"]
		element2 := element.([]interface{})[0].(map[string]interface{})["element2"].(map[string]interface{})
		if element2["added"] != TestValue {
			t.Errorf("Failed to update pod")
		}
	}
	testMutation(
		`spec.element["name": "value"].element2.added`,
		TestValue,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func testMutation(
	location string,
	value string,
	unstructured *unstructured.Unstructured,
	testFunc func(*unstructured.Unstructured),
	t *testing.T) {

	assign := mutationsv1alpha1.Assign{
		ObjectMeta: metav1.ObjectMeta{},

		Spec: mutationsv1alpha1.AssignSpec{
			Location: location,
			Parameters: mutationsv1alpha1.Parameters{
				Assign: runtime.RawExtension{
					Raw: []byte(fmt.Sprintf("{\"value\": \"%s\"}", value)),
				},
			},
		},
	}

	mutator, err := mutation.MutatorForAssign(&assign)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	err = mutator.Mutate(unstructured)
	if err != nil {
		t.Error("Unexpected error", err)
	} else {
		testFunc(unstructured)
	}
}
