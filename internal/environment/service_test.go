package environment

import (
	"reflect"
	"testing"
)

func TestEnvironment_extract(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name string
		env  *Environment
		args args
		want []string
	}{
		{
			"no fragments",
			&Environment{},
			args{
				source: "/test/foo/bar",
			},
			[]string{
				"/test/foo/bar",
			},
		},
		{
			"with fragments at beginning",
			&Environment{},
			args{
				source: "${{ config }}/test/foo/bar",
			},
			[]string{
				"${{ config }}",
				"/test/foo/bar",
			},
		},
		{
			"with fragments at middle",
			&Environment{},
			args{
				source: "/${{ config }}/foo/bar",
			},
			[]string{
				"/",
				"${{ config }}",
				"/foo/bar",
			},
		},
		{
			"with multiple fragments at middle",
			&Environment{},
			args{
				source: "/${{ config }}/${{ foo }}/bar",
			},
			[]string{
				"/",
				"${{ config }}",
				"/",
				"${{ foo }}",
				"/bar",
			},
		},
		{
			"with multiple fragments at beginning, middle and end",
			&Environment{},
			args{
				source: "${{ config }}/test/${{ foo }}/bar/${{ end }}",
			},
			[]string{
				"${{ config }}",
				"/test/",
				"${{ foo }}",
				"/bar/",
				"${{ end }}",
			},
		},
		{
			"with multiple fragments at beginning, middle and end, no closure",
			&Environment{},
			args{
				source: "${{ config }}/test/${{ foo/bar/${{ end }}",
			},
			[]string{
				"${{ config }}",
				"/test/",
				"${{ foo/bar/${{ end }}",
			},
		},
		{
			"with multiple fragments at beginning, middle and end, no closure",
			&Environment{},
			args{
				source: "config }}/test/${{ foo/bar/${{ end }}",
			},
			[]string{
				"config }}/test/${{ foo/bar/${{ end }}",
			},
		},
		{
			"with only beginning",
			&Environment{},
			args{
				source: "xyz${{xyz",
			},
			[]string{
				"xyz${{xyz",
			},
		},
		{
			"with only closure",
			&Environment{},
			args{
				source: "xyz}}xyz",
			},
			[]string{
				"xyz}}xyz",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.extract(tt.args.source); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Environment.extract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironment_Replace(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name string
		env  *Environment
		args args
		want string
	}{
		{
			"should change complex",
			&Environment{},
			args{
				source: "/${{config.test}}/foo/${{config.bar}}",
			},
			"/test_config/foo/bar",
		},
		{
			"should change individual",
			&Environment{},
			args{
				source: "${{config.test}}",
			},
			"test_config",
		},
		{
			"should not change individual",
			&Environment{},
			args{
				source: "${{config.test_not_found }}",
			},
			"${{config.test_not_found }}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.env.variables = map[string]map[string]interface{}{}
			tt.env.Add("config", "test", "test_config")
			tt.env.Add("config", "bar", "bar")
			if got := tt.env.Replace(tt.args.source); got != tt.want {
				t.Errorf("Environment.Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}
