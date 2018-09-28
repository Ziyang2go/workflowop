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
	Job1Batch string `json:"job1Batch"`
	Job2Batch string `json:"job2Batch"`
}

type WorkflowStatus struct {
	Status     string `json:"status"`
	Job1Status string `json:"job1Status"`
	Job2Status string `json:"job2Status"`
}

type WorkflowInputs struct {
	Jobs []Job `json:"jobs"`
}

type Job struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
}
