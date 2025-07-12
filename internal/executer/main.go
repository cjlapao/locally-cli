package executer

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

type ExecuteOutput struct {
	StdOut    string
	StdErr    string
	ErrorCode string
}

func (exe ExecuteOutput) GetAllOutput() string {
	output := ""
	if exe.StdOut != "" {
		output += exe.StdOut
	}
	if exe.StdErr != "" {
		output += exe.StdErr
	}

	return output
}

func Execute(command string, args ...string) (ExecuteOutput, error) {
	result := ExecuteOutput{}
	cmd := exec.Command(command, args...)
	var stdOut, stdIn bytes.Buffer

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdOut)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stdOut)
	cmd.Stdin = &stdIn

	if err := cmd.Run(); err != nil {
		result.StdErr = stdOut.String()
		result.StdOut = stdOut.String()
		result.ErrorCode = err.Error()
		return result, err
	}

	result.StdErr = stdOut.String()
	result.StdOut = stdOut.String()
	return result, nil
}

func ExecuteWithNoOutput(command string, args ...string) (ExecuteOutput, error) {
	result := ExecuteOutput{}
	cmd := exec.Command(command, args...)
	var stdOut, stdIn, stdErr bytes.Buffer

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Stdin = &stdIn

	if err := cmd.Run(); err != nil {
		result.StdErr = stdErr.String()
		result.StdOut = stdOut.String()
		return result, err
	}

	result.StdErr = stdErr.String()
	result.StdOut = stdOut.String()
	return result, nil
}

func ExecuteWithNoOutputContext(ctx context.Context, command string, args ...string) (ExecuteOutput, error) {
	result := ExecuteOutput{}
	cmd := exec.CommandContext(ctx, command, args...)
	var stdOut, stdIn, stdErr bytes.Buffer

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Stdin = &stdIn

	if err := cmd.Run(); err != nil {
		result.StdErr = stdErr.String()
		result.StdOut = stdOut.String()
		return result, err
	}

	result.StdErr = stdErr.String()
	result.StdOut = stdOut.String()
	return result, nil
}

func ExecuteAndWatch(command string, args ...string) (ExecuteOutput, error) {
	result := ExecuteOutput{}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	var stdOut, stdErr, stdIn bytes.Buffer

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdOut)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stdErr)
	cmd.Stdin = &stdIn

	if err := cmd.Start(); err != nil {
		result.StdErr = stdErr.String()
		result.StdOut = stdOut.String()
		return result, err
	}

	go func() {
		<-ctx.Done()
	}()

	if err := cmd.Wait(); err != nil {
		result.StdErr = stdErr.String()
		result.StdOut = stdOut.String()
		return result, err
	}

	result.StdErr = stdErr.String()
	result.StdOut = stdOut.String()
	return result, nil
}
