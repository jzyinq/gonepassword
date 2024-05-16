// Package gonepassword provides utilities for fetching secrets from 1Password cli client.
package gonepassword

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// OnePasswordClient is an interface for fetching secrets from 1Password.
type OnePasswordClient interface {
	ResolveOpURI(uri string) (string, error)
}

// OnePassword is a wrapper around the 1Password CLI.
type OnePassword struct {
	executor CommandExecutor
	opStorage
	isInstalled bool
	options     OnePasswordOptions
}

// OnePasswordOptions is a struct that holds the options for the 1Password client.
type OnePasswordOptions struct {
	// ServiceAccountToken is the token used to authenticate with 1Password instead of an app
	ServiceAccountToken string
	// Account is the `--account` op cli argument to use when fetching secrets.
	Account string
}

const binName string = "op"
const opURIPrefix string = "op://"
const serviceAccountTokenEnv = "OP_SERVICE_ACCOUNT_TOKEN" //nolint
const retryAttempts = 5

// New1Password creates a new OnePassword instance.
// serviceAccountToken can be passed directly to constructor, or it will be read from environment variable.
func New1Password(executor CommandExecutor, options OnePasswordOptions) (*OnePassword, error) {
	if executor == nil {
		executor = DefaultCommandExecutor{options.ServiceAccountToken}
	}
	opCli := &OnePassword{executor: executor, opStorage: newOPStorage(), options: options}
	opCli.isInstalled = opCli.executor.IsInstalled()
	return opCli, nil
}

// parseOpURI parses the given 1Password URI and returns its components.
func parseOpURI(uri string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(uri, opURIPrefix), "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid 1Password URI format - expected op://vault/item/field - got '%s'", uri)
	}
	return parts[0], parts[1], parts[2], nil
}

// ResolveOpURI resolves the given 1Password URI and returns its value.
// It also caches whole uri item in memory to avoid multiple calls to 1Password CLI
// while fetching other fields from the same item.
func (cli *OnePassword) ResolveOpURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, opURIPrefix) {
		return uri, &InvalidOpURIError{uri: uri}
	}
	logrus.Info("Resolving 1password entry: ", uri)
	vault, item, field, err := parseOpURI(uri)
	if err != nil {
		return "", err
	}
	if !cli.isInstalled {
		logrus.Error(&OnePasswordCliNotInstalledError{})
		return "", &OnePasswordCliNotInstalledError{}
	}
	vaultItem, err := cli.opStorage.getVaultItem(vault, item)
	if err != nil {
		executorCmd := []string{"item", "get", "--format", "json", item, "--vault", vault}
		if cli.options.Account != "" {
			executorCmd = append(executorCmd, "--account", cli.options.Account)
		}
		output, err := cli.executor.Execute(executorCmd...)
		if err != nil {
			return "", err
		}
		var opItem opItem
		err = json.Unmarshal(output, &opItem)
		if err != nil {
			return "", err
		}
		cli.opStorage.setVaultItem(vault, item, opItem)
		vaultItem = opItem
	}
	fieldValue, err := vaultItem.GetField(field)
	if err != nil {
		return "", err
	}
	return fieldValue, nil
}
