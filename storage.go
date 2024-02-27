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
