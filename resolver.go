package gonepassword

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
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

// OPStorage is a struct that holds the data returned by the 1Password CLI.
type OPStorage struct {
	Vaults map[string]opVault
}

// setVaultItem sets the given item in the given vault.
func (o OPStorage) setVaultItem(vault string, itemRef string, item opItem) {
	if _, ok := o.Vaults[vault]; !ok {
		o.Vaults[vault] = opVault{ID: vault, Items: make(map[string]opItem)}
	}
	o.Vaults[vault].Items[itemRef] = item
}

// getVaultItem returns the given item from the given vault, return an error if the item or vault does not exist.
func (o OPStorage) getVaultItem(vault string, item string) (opItem, error) {
	if _, ok := o.Vaults[vault]; !ok {
		return opItem{}, fmt.Errorf("no such vault %s", vault)
	}
	if _, ok := o.Vaults[vault].Items[item]; !ok {
		return opItem{}, fmt.Errorf("no such item %s in vault %s", item, vault)
	}
	return o.Vaults[vault].Items[item], nil
}

type opVault struct {
	ID    string
	Items map[string]opItem
}

type opItem struct {
	ID     string    `json:"id"`
	Fields []opField `json:"fields"`
}

type opField struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	Reference string `json:"reference"`
}

// GetField returns the value of the given field, returns an error if the field does not exist.
func (o opItem) GetField(field string) (string, error) {
	for _, f := range o.Fields {
		if f.ID == field {
			return f.Value, nil
		}
		if f.Label == field {
			return f.Value, nil
		}
	}
	return "", fmt.Errorf("field %s not found", field)
}

func newOPStorage() OPStorage {
	return OPStorage{Vaults: make(map[string]opVault)}
}

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
				logrus.Error(os.Stderr, "it looks like 1password-1problem, let's ask them again...\n")
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

// New1Password creates a new OnePassword instance.
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
	logrus.Info("Resolving 1password entry: ", uri)
	if !strings.HasPrefix(uri, opURIPrefix) {
		return uri, fmt.Errorf("incorrect op uri - it should look like op://vault/item/field")
	}
	vault, item, field, err := parseOpURI(uri)
	if err != nil {
		return "", err
	}
	if !cli.isInstalled {
		opNotInstalledError := "1Password CLI is not installed, visit " +
			"https://developer.1password.com/docs/cli/get-started/#step-1-install-1password-cli"
		logrus.Error(opNotInstalledError)
		return "", fmt.Errorf(opNotInstalledError)
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
