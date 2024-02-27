package gonepassword

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

type commandExecutor interface {
	IsInstalled() bool
	Execute(arg ...string) ([]byte, error)
}

type defaultCommandExecutor struct {
	serviceAccountToken string
}

func (e defaultCommandExecutor) Execute(arg ...string) ([]byte, error) {
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
func (e defaultCommandExecutor) IsInstalled() bool {
	_, err := exec.LookPath(binName)
	return err == nil
}
