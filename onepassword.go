// Package gonepassword provides utilities for fetching secrets from 1Password cli client.
package gonepassword

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// OnePassword is a wrapper around the 1Password CLI.
type OnePassword struct {
	executor commandExecutor
	OPStorage
	isInstalled bool
}

const binName string = "op"
const opURIPrefix string = "op://"
const serviceAccountTokenEnv = "OP_SERVICE_ACCOUNT_TOKEN" //nolint
const retryAttempts = 5

// New1Password creates a new OnePassword instance.
// serviceAccountToken can be passed directly to constructor, or it will be read from environment variable.
func New1Password(executor commandExecutor, serviceAccountToken string) (*OnePassword, error) {
	if executor == nil {
		executor = defaultCommandExecutor{serviceAccountToken}
	}
	opCli := &OnePassword{executor: executor, OPStorage: newOPStorage()}
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
	vaultItem, err := cli.OPStorage.getVaultItem(vault, item)
	if err != nil {
		output, err := cli.executor.Execute("item", "get", "--format", "json", item, "--vault", vault)
		if err != nil {
			return "", err
		}
		var opItem opItem
		err = json.Unmarshal(output, &opItem)
		if err != nil {
			return "", err
		}
		cli.OPStorage.setVaultItem(vault, item, opItem)
		vaultItem = opItem
	}
	fieldValue, err := vaultItem.GetField(field)
	if err != nil {
		return "", err
	}
	return fieldValue, nil
}
