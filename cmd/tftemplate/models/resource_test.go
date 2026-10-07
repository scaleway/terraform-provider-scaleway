package models

import (
	"reflect"
	"testing"
)

// K8SCluster
// IamSSHKey
func Test_splitByWord(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"FunctionNamespace",
			[]string{"Function", "Namespace"},
		},
		{
			"K8SCluster",
			[]string{"K8S", "Cluster"},
		},
		{
			"IamSSHKey",
			[]string{"Iam", "SSH", "Key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitByWord(tt.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitByWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
