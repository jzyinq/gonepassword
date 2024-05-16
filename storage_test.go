package gonepassword

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpStorage(t *testing.T) {
	storage := newOPStorage()

	vaultName := "testVault"
	itemName := "testItem"
	testItem := opItem{ID: itemName}
	storage.setVaultItem(vaultName, itemName, testItem)

	assert.NotNil(t, storage.Vaults[vaultName], "Vault was not set")
	assert.NotNil(t, storage.Vaults[vaultName].Items[itemName], "Item was not set in vault")

	retrievedItem, err := storage.getVaultItem(vaultName, itemName)
	assert.NoError(t, err, "Error retrieving item")
	assert.Equal(t, testItem, retrievedItem, "Retrieved item does not match set item")
}

func TestOpFieldMatchField(t *testing.T) {
	field := opField{
		ID:      "field-id",
		Type:    "field-type",
		Label:   "field-label",
		Value:   "field-value",
		Section: opSection{ID: "add more", Label: "section-label"},
	}

	testCases := []struct {
		name     string
		uri      *OpURI
		expected bool
	}{
		{
			name: "ID matches but the section does not",
			uri: &OpURI{
				raw:     "op://vault/item/field-id",
				vault:   "vault",
				item:    "item",
				field:   "field-id",
				section: "wrong-section",
			},
			expected: false,
		},
		{
			name: "Section matches but neither the ID nor the Label does",
			uri: &OpURI{
				raw:     "op://vault/item/wrong-field",
				vault:   "vault",
				item:    "item",
				field:   "wrong-field",
				section: "section-id",
			},
			expected: false,
		},
		{
			name: "Both the ID and section match",
			uri: &OpURI{
				raw:     "op://vault/item/field-id",
				vault:   "vault",
				item:    "item",
				field:   "field-id",
				section: "add more",
			},
			expected: true,
		},
		{
			name: "Section is empty",
			uri: &OpURI{
				raw:     "op://vault/item/field-id",
				vault:   "vault",
				item:    "item",
				field:   "field-id",
				section: "",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, field.matchField(tc.uri))
		})
	}
}

func TestOpFileMatchFile(t *testing.T) { //nolint:funlen
	file := opFile{
		ID:      "file-id",
		Name:    "file-name",
		Section: opSection{ID: "add more", Label: "section-label"},
	}

	testCases := []struct {
		name     string
		uri      *OpURI
		expected bool
	}{
		{
			name: "ID matches but the section does not",
			uri: &OpURI{
				raw:     "op://vault/item/file-id",
				vault:   "vault",
				item:    "item",
				field:   "file-id",
				section: "wrong-section",
			},
			expected: false,
		},
		{
			name: "Name matches but the section does not",
			uri: &OpURI{
				raw:     "op://vault/item/file-name",
				vault:   "vault",
				item:    "item",
				field:   "file-name",
				section: "wrong-section",
			},
			expected: false,
		},
		{
			name: "Section matches but neither the ID nor the Name does",
			uri: &OpURI{
				raw:     "op://vault/item/wrong-field",
				vault:   "vault",
				item:    "item",
				field:   "wrong-field",
				section: "section-id",
			},
			expected: false,
		},
		{
			name: "Both the ID and section match",
			uri: &OpURI{
				raw:     "op://vault/item/file-id",
				vault:   "vault",
				item:    "item",
				field:   "file-id",
				section: "add more",
			},
			expected: true,
		},
		{
			name: "Both the Name and section match",
			uri: &OpURI{
				raw:     "op://vault/item/file-name",
				vault:   "vault",
				item:    "item",
				field:   "file-name",
				section: "add more",
			},
			expected: true,
		},
		{
			name: "Section is empty",
			uri: &OpURI{
				raw:     "op://vault/item/file-id",
				vault:   "vault",
				item:    "item",
				field:   "file-id",
				section: "",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, file.matchFile(tc.uri))
		})
	}
}

func TestGetFieldValue(t *testing.T) { //nolint:funlen
	item := opItem{
		ID: "item-id",
		Fields: []opField{
			{
				ID:      "field-id",
				Type:    "field-type",
				Label:   "field-label",
				Value:   "field-value",
				Section: opSection{ID: "section-id", Label: "section-label"},
			},
		},
		Files: []opFile{
			{
				ID:      "file-id",
				Name:    "file-name",
				Section: opSection{ID: "section-id", Label: "section-label"},
			},
		},
	}

	testCases := []struct {
		name           string
		uri            *OpURI
		executor       *SpyCommandExecutor
		expectedOutput string
		expectedError  string
	}{
		{
			name: "Field exists and matches the URI",
			uri: &OpURI{
				raw:     "op://vault/item/section-id/field-id",
				vault:   "vault",
				item:    "item",
				field:   "field-id",
				section: "section-id",
			},
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
			},
			expectedOutput: "field-value",
		},
		{
			name: "Field does not exist in the item",
			uri: &OpURI{
				raw:     "op://vault/item/section-id/nonexistent-field",
				vault:   "vault",
				item:    "item",
				field:   "nonexistent-field",
				section: "section-id",
			},
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
			},
			expectedError: "field nonexistent-field not found",
		},
		{
			name: "Field exists but the section does not match",
			uri: &OpURI{
				raw:     "op://vault/item/wrong-section/field-id",
				vault:   "vault",
				item:    "item",
				field:   "field-id",
				section: "wrong-section",
			},
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
			},
			expectedError: "field field-id not found",
		},
		{
			name: "Field is a file",
			uri: &OpURI{
				raw:     "op://vault/item/section-id/file-id",
				vault:   "vault",
				item:    "item",
				field:   "file-id",
				section: "section-id",
			},
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
				ExecuteOutput:  []byte("itsa me - a file!"),
			},
			expectedOutput: "itsa me - a file!",
		},
		{
			name: "Fetching file fails",
			uri: &OpURI{
				raw:     "op://vault/item/section-id/file-id",
				vault:   "vault",
				item:    "item",
				field:   "file-id",
				section: "section-id",
			},
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
				ExecuteError:   errors.New("oh noez"),
			},
			expectedError: "oh noez",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli, err := New1Password(tc.executor, OnePasswordOptions{})
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}

			fieldValue, err := item.GetFieldValue(cli, tc.uri)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, fieldValue)
			}
		})
	}
}
