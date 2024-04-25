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
}

// OpURI is a struct that holds the parsed 1Password URI.
type OpURI struct {
	vault   string
	item    string
	field   string
	section string
	//query   string // TODO query support some day
	raw string
}

// NewOpURI creates a new OpURI instance.
func NewOpURI(uri string) (*OpURI, error) {
	parts := strings.Split(strings.TrimPrefix(uri, opURIPrefix), "/")
	numParts := len(parts)
	var opURI OpURI
	if numParts < 3 || numParts > 4 {
		return nil, fmt.Errorf("invalid 1Password URI format - expected op://vault/item/field - got '%s'", uri)
	}
	opURI = OpURI{raw: uri, vault: parts[0], item: parts[1], field: parts[2], section: ""}
	if numParts == 4 {
		opURI.section = parts[2]
		opURI.field = parts[3]
	}
	return &opURI, nil
}

const binName string = "op"
const opURIPrefix string = "op://"
const serviceAccountTokenEnv = "OP_SERVICE_ACCOUNT_TOKEN" //nolint
const retryAttempts = 5

// New1Password creates a new OnePassword instance.
// serviceAccountToken can be passed directly to constructor, or it will be read from environment variable.
func New1Password(executor CommandExecutor, serviceAccountToken string) (*OnePassword, error) {
	if executor == nil {
		executor = DefaultCommandExecutor{serviceAccountToken}
	}
	opCli := &OnePassword{executor: executor, opStorage: newOPStorage()}
	opCli.isInstalled = opCli.executor.IsInstalled()
	return opCli, nil
}

// ResolveOpURI resolves the given 1Password URI and returns its value.
// It also caches whole uri item in memory to avoid multiple calls to 1Password CLI
// while fetching other fields from the same item.
func (cli *OnePassword) ResolveOpURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, opURIPrefix) {
		return uri, &InvalidOpURIError{uri: uri}
	}
	logrus.Info("Resolving 1password entry: ", uri)
	opURI, err := NewOpURI(uri)
	if err != nil {
		return "", err
	}
	if !cli.isInstalled {
		logrus.Error(&OnePasswordCliNotInstalledError{})
		return "", &OnePasswordCliNotInstalledError{}
	}
	vaultItem, err := cli.opStorage.getVaultItem(opURI.vault, opURI.item)
	if err != nil {
		output, err := cli.executor.Execute("item", "get", "--format", "json", opURI.item, "--vault", opURI.vault)
		if err != nil {
			return "", err
		}
		var opItem opItem
		err = json.Unmarshal(output, &opItem)
		if err != nil {
			return "", err
		}
		cli.opStorage.setVaultItem(opURI.vault, opURI.item, opItem)
		vaultItem = opItem
	}
	fieldValue, err := vaultItem.GetFieldValue(cli, opURI)
	if err != nil {
		return "", err
	}
	return fieldValue, nil
}
