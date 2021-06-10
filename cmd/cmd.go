package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

type State string

var aws_region string

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

func (d *ParameterDiff) GetPutParameterInput() *ssm.PutParameterInput {
	return new(ssm.PutParameterInput).SetName(d.Path + d.Name).SetValue(d.NewValue).SetType(ssm.ParameterTypeSecureString).SetOverwrite(true)
}

func GetSSMService() (*ssm.SSM, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	var config *aws.Config

	if aws_region == "" {
		config = aws.NewConfig()
	} else {
		config = &aws.Config{
			Region: aws.String(aws_region),
		}
	}
	return ssm.New(sess, config), nil
}

func GetParametersByPath(path string) ([]*ssm.Parameter, error) {
	var params []*ssm.Parameter
	ssmsvc, err := GetSSMService()
	if err != nil {
		return nil, err
	}
	withDecryption := true
	result, err := ssmsvc.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           &path,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, err
	}
	params = append(params, result.Parameters...)
	for result.NextToken != nil {
		result, err = ssmsvc.GetParametersByPath(&ssm.GetParametersByPathInput{
			Path:           &path,
			NextToken:      result.NextToken,
			WithDecryption: &withDecryption,
		})
		if err != nil {
			return nil, err
		}
		params = append(params, result.Parameters...)
	}
	sort.Slice(params, func(i, j int) bool {
		s := []string{*(params[i].Name), *(params[j].Name)}
		sort.Strings(s)
		return s[0] == *(params[i].Name)
	})
	return params, nil
}

func List(c *cli.Context) error {
	aws_region = c.String("region")
	path := "/" + strings.Trim(c.Args().Get(0), "/") + "/"
	params, err := GetParametersByPath(path)
	if err != nil {
		return err
	}
	for _, param := range params {
		fmt.Printf("%s=%s\n", strings.Replace(*param.Name, path, "", 1), strings.Replace(*param.Value, "\n", "\\n", -1))
	}
	return nil
}

func ApplyOperation(diffs ParameterDiffs, ssmsvc *ssm.SSM) error {
	fmt.Println(color.HiCyanString(" State   \tName"))
	for _, diff := range diffs {
		switch diff.State() {
		case StateCreate, StateUpdate:
			if _, err := ssmsvc.PutParameter(diff.GetPutParameterInput()); err != nil {
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
	aws_region = c.String("region")
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
	diffs, err := DiffOperation(path, envMap)
	if err != nil {
		return err
	}
	if c.Bool("y") {
		fmt.Println("applying changes")
		return ApplyOperation(diffs, ssmsvc)
	} else {
		fmt.Printf("Do you want to update? [y/N]:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Text() == "y" {
			fmt.Println("applying changes")
			return ApplyOperation(diffs, ssmsvc)
		}
	}
	fmt.Println("Operation cancelled by user.")
	return nil
}

func DiffOperation(path string, envMap map[string]string) (ParameterDiffs, error) {
	params, err := GetParametersByPath(path)
	if err != nil {
		return nil, err
	}
	diffs := ParameterDiffs{}
	for _, param := range params {
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
	sort.Slice(diffs, func(i, j int) bool {
		s := []string{diffs[i].Name, diffs[j].Name}
		sort.Strings(s)
		return s[0] == diffs[i].Name
	})

	for _, diff := range diffs {
		fmt.Println(diff.String())
	}
	return diffs, nil
}

func Diff(c *cli.Context) error {
	aws_region = c.String("region")
	path := "/" + strings.Trim(c.Args().Get(0), "/") + "/"
	filename := c.String("file")
	envMap, err := godotenv.Read(filename)
	if err != nil {
		return err
	}
	_, err = DiffOperation(path, envMap)
	if err != nil {
		return err
	}
	return nil
}
