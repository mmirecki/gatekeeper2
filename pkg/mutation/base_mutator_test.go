package mutation_test

import (
	"testing"

	. "github.com/open-policy-agent/gatekeeper/pkg/mutation"
	"github.com/open-policy-agent/gatekeeper/pkg/mutation/path/parser"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const TestValue = "testValue"

func toPtr(s string) *string {
	return &s
}

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
					Name: "testname1",
					Ports: []corev1.ContainerPort{
						{
							Name: "portName1",
						},
					},
				},
				{
					Name: "testname2",
					Ports: []corev1.ContainerPort{
						{
							Name: "portName2A",
						},
						{
							Name: "portName2B",
						},
					},
				},
				{
					Name: "testname3",
					Ports: []corev1.ContainerPort{
						{
							Name: "portName3",
						},
					},
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
	pathEntries := []parser.Node{
		parser.Object{Reference: "metadata"},
		parser.Object{Reference: "labels"},
		parser.Object{Reference: "labelA"},
	}

	testFunc := func(unstructured *unstructured.Unstructured) {
		labels := unstructured.Object["metadata"].(map[string]interface{})["labels"]
		if labels.(map[string]interface{})["labelA"] != TestValue {
			t.Errorf("Failed to update pod")
		}

	}

	testMutation(
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestObjectsAndLists(t *testing.T) {
	pathEntries := []parser.Node{
		parser.Object{Reference: "spec"},
		parser.Object{Reference: "containers"},
		parser.List{KeyField: "name", KeyValue: toPtr("testname2")},
		parser.Object{Reference: "ports"},
		parser.List{KeyField: "name", KeyValue: toPtr("portName2B")},
		parser.Object{Reference: "hostIP"},
	}
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
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestListsAsLastElement(t *testing.T) {
	pathEntries := []parser.Node{
		parser.Object{Reference: "spec"},
		parser.Object{Reference: "containers"},
		parser.List{KeyField: "name", KeyValue: toPtr(TestValue)},
	}
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
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestListsAsLastElementAlreadyExists(t *testing.T) {
	pathEntries := []parser.Node{
		parser.Object{Reference: "spec"},
		parser.Object{Reference: "containers"},
		parser.List{KeyField: "name", KeyValue: toPtr("testname1")},
	}
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
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestGlobbedList(t *testing.T) {
	pathEntries := []parser.Node{
		parser.Object{Reference: "spec"},
		parser.Object{Reference: "containers"},
		parser.List{KeyField: "name", Glob: true},
		parser.Object{Reference: "ports"},
		parser.List{KeyField: "name", Glob: true},
		parser.Object{Reference: "protocol"},
	}

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
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestNonExistingPathEntry(t *testing.T) {
	pathEntries := []parser.Node{
		parser.Object{Reference: "spec"},
		parser.Object{Reference: "element"},
		parser.Object{Reference: "should"},
		parser.Object{Reference: "be"},
		parser.Object{Reference: "added"},
	}
	testFunc := func(unstructured *unstructured.Unstructured) {
		element := unstructured.Object["spec"].(map[string]interface{})["element"].(map[string]interface{})["should"].(map[string]interface{})["be"]
		if element.(map[string]interface{})["added"] != TestValue {
			t.Errorf("Failed to update pod")
		}
	}
	testMutation(
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func TestNonExistingListPathEntry(t *testing.T) {
	pathEntries := []parser.Node{
		parser.Object{Reference: "spec"},
		parser.Object{Reference: "element"},
		parser.List{KeyField: "name", KeyValue: toPtr("value")},
		parser.Object{Reference: "element2"},
		parser.Object{Reference: "added"},
	}
	testFunc := func(unstructured *unstructured.Unstructured) {
		element := unstructured.Object["spec"].(map[string]interface{})["element"]
		element2 := element.([]interface{})[0].(map[string]interface{})["element2"].(map[string]interface{})
		if element2["added"] != TestValue {
			t.Errorf("Failed to update pod")
		}
	}
	testMutation(
		pathEntries,
		prepareTestPod(t),
		testFunc,
		t,
	)
}

func testMutation(
	nodes []parser.Node,
	unstructured *unstructured.Unstructured,
	testFunc func(*unstructured.Unstructured),
	t *testing.T) {

	mutator := NewMutator(parser.Path{Nodes: nodes}, TestValue)
	err := mutator.Mutate(unstructured)
	if err != nil {
		t.Error("Unexpected error", err)
	} else {
		testFunc(unstructured)
	}
}
