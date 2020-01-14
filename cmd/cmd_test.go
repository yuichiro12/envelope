package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
)

func TestApplyOperation(t *testing.T) {
	type args struct {
		diffs  ParameterDiffs
		ssmsvc *ssm.SSM
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ApplyOperation(tt.args.diffs, tt.args.ssmsvc); (err != nil) != tt.wantErr {
				t.Errorf("ApplyOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiffOperation(t *testing.T) {
	type args struct {
		path   string
		ssmsvc *ssm.SSM
		envMap map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    ParameterDiffs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiffOperation(tt.args.path, tt.args.envMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiffOperation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPutParameter(t *testing.T) {
	ssmsvc, err := GetSSMService()
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Println(ssmsvc.ServiceName)
	fmt.Println(ssmsvc.SigningRegion)
	in := new(ssm.PutParameterInput).SetName("/hoge/fuga/TEST").SetValue("1").SetType(ssm.ParameterTypeSecureString).SetOverwrite(true)
	if _, err := ssmsvc.PutParameter(in); err != nil {
		t.Fatal(err)
	}
}
func TestDeleteParameter(t *testing.T) {
	ssmsvc, err := GetSSMService()
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Println(ssmsvc.ServiceName)
	fmt.Println(ssmsvc.SigningRegion)
	in := new(ssm.DeleteParameterInput).SetName("/hoge/fuga/TEST")
	if _, err := ssmsvc.DeleteParameter(in); err != nil {
		t.Fatal(err)
	}
}
