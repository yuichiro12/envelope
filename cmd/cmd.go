package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

type State string

const (
	StateCreate   State = "create"
	StateUpdate   State = "update"
	StateDelete   State = "delete"
	StateUnchange State = "unchange"
)

func (s State) Colored() string {
	switch s {
	case StateCreate:
		return color.GreenString(string(s) + "  ")
	case StateUpdate:
		return color.HiMagentaString(string(s) + "  ")
	case StateDelete:
		return color.RedString(string(s) + "  ")
	case StateUnchange:
		return string(s)
	default:
		panic("no such state: " + s)
	}
}

type ParameterDiff struct {
	Path     string
	Name     string
	OldValue string
	NewValue string
	Input    *ssm.PutParameterInput
}

type ParameterDiffs []*ParameterDiff

func (d *ParameterDiff) State() State {
	if d.OldValue == d.NewValue {
		return StateUnchange
	} else if d.OldValue == "" {
		return StateCreate
	} else if d.NewValue == "" {
		return StateDelete
	} else {
		return StateUpdate
	}
}

func (d *ParameterDiff) String() string {
	return fmt.Sprintf(" %s\t%s", d.State().Colored(), d.Name)
}

func (d ParameterDiffs) Get(name string) *ParameterDiff {
	for _, v := range d {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (d *ParameterDiff) PutParameterInput() *ssm.PutParameterInput {
	return new(ssm.PutParameterInput).SetName(d.Path + d.Name).SetValue(d.NewValue).SetType(ssm.ParameterTypeSecureString).SetOverwrite(true)
}

func GetSSMService() (*ssm.SSM, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	return ssm.New(sess, aws.NewConfig()), nil
}

func List(c *cli.Context) error {
	ssmsvc, err := GetSSMService()
	if err != nil {
		return err
	}
	path := "/" + strings.Trim(c.Args().Get(0), "/") + "/"
	withDecryption := true
	params, err := ssmsvc.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           &path,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return err
	}
	for _, param := range params.Parameters {
		fmt.Printf("%s=\"%s\"\n", strings.Replace(*param.Name, path, "", 1), strings.Replace(*param.Value, "\n", "\\n", -1))
	}
	return nil
}

func ApplyOperation(diffs ParameterDiffs, ssmsvc *ssm.SSM) error {
	fmt.Println(color.HiCyanString(" State   \tName"))
	for _, diff := range diffs {
		switch diff.State() {
		case StateCreate, StateUpdate:
			if _, err := ssmsvc.PutParameter(diff.PutParameterInput()); err != nil {
				return err
			}
			fmt.Println(diff.String())
		case StateDelete:
			name := diff.Path + diff.Name
			if _, err := ssmsvc.DeleteParameter(&ssm.DeleteParameterInput{Name: &name}); err != nil {
				return err
			}
			fmt.Println(diff.String())
		}

	}
	return nil
}

func Apply(c *cli.Context) error {
	ssmsvc, err := GetSSMService()
	if err != nil {
		return err
	}
	path := "/" + strings.Trim(c.Args().Get(0), "/") + "/"
	filename := c.String("file")
	envMap, err := godotenv.Read(filename)
	if err != nil {
		return err
	}
	diffs, err := DiffOperation(path, ssmsvc, envMap)
	if err != nil {
		return err
	}
	fmt.Printf("Do you want to update? [y/N]:")
	var in string
	_, err = fmt.Scan(&in)
	if err != nil {
		return err
	}
	if in == "y" {
		return ApplyOperation(diffs, ssmsvc)
	}
	fmt.Println("Operation cancelled.")
	return nil
}

func DiffOperation(path string, ssmsvc *ssm.SSM, envMap map[string]string) (ParameterDiffs, error) {
	withDecryption := true
	params, err := ssmsvc.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           &path,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, err
	}
	diffs := ParameterDiffs{}
	for _, param := range params.Parameters {
		diffs = append(diffs, &ParameterDiff{Path: path, Name: strings.Replace(*param.Name, path, "", 1), OldValue: *param.Value})
	}
	for name, value := range envMap {
		if diff := diffs.Get(name); diff != nil {
			diff.Name = name
			diff.NewValue = value
		} else {
			diffs = append(diffs, &ParameterDiff{Path: path, Name: name, NewValue: value})
		}
	}
	fmt.Println(color.HiCyanString(" State   \tName"))
	for _, diff := range diffs {
		fmt.Println(diff.String())
	}
	return diffs, nil
}

func Diff(c *cli.Context) error {
	ssmsvc, err := GetSSMService()
	if err != nil {
		return err
	}
	path := "/" + strings.Trim(c.Args().Get(0), "/") + "/"
	filename := c.String("file")
	envMap, err := godotenv.Read(filename)
	if err != nil {
		return err
	}
	_, err = DiffOperation(path, ssmsvc, envMap)
	if err != nil {
		return err
	}
	return nil
}
