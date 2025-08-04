package docker

// import (
// 	"errors"
// 	"fmt"
// 	"os"

// 	"github.com/cjlapao/locally-cli/internal/common"
// )

// type BuildImageOptions struct {
// 	Name       string
// 	Tag        string
// 	Context    string
// 	FilePath   string
// 	Parameters map[string]string
// 	UseCache   bool
// }

// func (b BuildImageOptions) GetArguments() ([]string, error) {
// 	args := make([]string, 0)
// 	if b.Name == "" {
// 		err := errors.New("image name cannot be empty")
// 		notify.Error(err.Error())
// 		return args, err
// 	}

// 	if b.Tag == "" {
// 		b.Tag = "latest"
// 	}

// 	if b.Context == "" {
// 		b.Context = "."
// 	}

// 	args = append(args, "build")
// 	args = append(args, "-t")
// 	args = append(args, fmt.Sprintf("%v:%v", b.Name, b.Tag))

// 	// Reading all the build parameters
// 	for key, value := range b.Parameters {
// 		os.Setenv(key, value)
// 		args = append(args, "--build-arg")
// 		args = append(args, key)
// 	}

// 	if b.FilePath != "" {
// 		args = append(args, "-f")
// 		args = append(args, b.FilePath)
// 	}

// 	if !b.UseCache {
// 		args = append(args, "--no-cache")
// 	}

// 	if common.IsVerbose() {
// 		args = append(args, "--progress=plain")
// 	}

// 	args = append(args, b.Context)

// 	return args, nil
// }
