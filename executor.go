package gonepassword

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

// CommandExecutor is an interface for executing commands through op CLI.
type CommandExecutor interface {
	IsInstalled() bool
	Execute(arg ...string) ([]byte, error)
}

// DefaultCommandExecutor is the default implementation of CommandExecutor.
type DefaultCommandExecutor struct {
	serviceAccountToken string
}

// Execute executes the given command and returns its output.
func (e DefaultCommandExecutor) Execute(arg ...string) ([]byte, error) {
	output, err := retry(retryAttempts, exponentialBackoff, func() (any, error) {
		var stdErr bytes.Buffer
		executor := exec.Command(binName, arg...)
		if e.serviceAccountToken != "" {
			executor.Env = append(
				os.Environ(), fmt.Sprintf("%s=%s", serviceAccountTokenEnv, e.serviceAccountToken),
			)
		}
		executor.Stderr = &stdErr
		output, err := executor.Output()
		_, _ = os.Stderr.Write(stdErr.Bytes())
		if err != nil {
			if strings.Contains(stdErr.String(), "https://") {
				logrus.Error("it looks like 1password-1problem, let's ask them again...\n")
				return output, fmt.Errorf(stdErr.String())
			}
			return output, &nonRetryableError{stdErr.String()}
		}
		return output, err
	})
	return output.([]byte), err
}

// IsInstalled returns true if the 1Password CLI is installed.
func (e DefaultCommandExecutor) IsInstalled() bool {
	_, err := exec.LookPath(binName)
	return err == nil
}
