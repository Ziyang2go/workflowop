package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Workflow `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Inputs            WorkflowInputs `json:"inputs"`
	Spec              WorkflowSpec   `json:"spec"`
	Status            WorkflowStatus `json:"status,omitempty"`
}

type WorkflowSpec struct {
	JobBatch map[string]*BatchReference `json:"jobBatch"`
}

type WorkflowStatus struct {
	Status    string            `json:"status"`
	JobStatus map[string]string `json:"jobStatus"`
}

type WorkflowInputs struct {
	Jobs []Job `json:"jobs"`
}

type Job struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
}

type BatchReference struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Logs string `json:"logs"`
}
