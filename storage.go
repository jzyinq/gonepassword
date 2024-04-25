package gonepassword

import "fmt"

// opStorage is a struct that holds the data returned by the 1Password CLI.
type opStorage struct {
	Vaults map[string]opVault
}

func newOPStorage() opStorage {
	return opStorage{Vaults: make(map[string]opVault)}
}

// setVaultItem sets the given item in the given vault.
func (o opStorage) setVaultItem(vault string, itemRef string, item opItem) {
	if _, ok := o.Vaults[vault]; !ok {
		o.Vaults[vault] = opVault{ID: vault, Items: make(map[string]opItem)}
	}
	o.Vaults[vault].Items[itemRef] = item
}

// getVaultItem returns the given item from the given vault, return an error if the item or vault does not exist.
func (o opStorage) getVaultItem(vault string, item string) (opItem, error) {
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
	Files  []opFile  `json:"files"`
}

type opField struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"`
	Label   string    `json:"label"`
	Value   string    `json:"value"`
	Section opSection `json:"section"`
}

func (of opField) matchField(uri *OpURI) bool {
	return (of.ID == uri.field || of.Label == uri.field) && of.Section.matchSection(uri.section)
}

type opFile struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Section opSection `json:"section"`
}

func (of opFile) matchFile(uri *OpURI) bool {
	return (of.Name == uri.field || of.ID == uri.field) && of.Section.matchSection(uri.section)
}

type opSection struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}

func (os opSection) matchSection(section string) bool {
	return os.Id == section || os.Label == section
}

// GetFieldValue returns the value of the given field, returns an error if the field does not exist.
func (o opItem) GetFieldValue(cli *OnePassword, uri *OpURI) (string, error) {
	for _, f := range o.Fields {
		if f.matchField(uri) {
			return f.Value, nil
		}
	}
	for _, f := range o.Files {
		if f.matchFile(uri) {
			output, err := cli.executor.Execute("read", uri.raw)
			if err != nil {
				return "", err
			}
			return string(output), nil
		}
	}
	return "", fmt.Errorf("field %s not found", uri.field)
}
