package gonepassword

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestOpStorage(t *testing.T) {
	storage := newOPStorage()

	// Test setVaultItem
	vaultName := "testVault"
	itemName := "testItem"
	testItem := opItem{ID: itemName}
	storage.setVaultItem(vaultName, itemName, testItem)

	if _, ok := storage.Vaults[vaultName]; !ok {
		t.Errorf("Vault %s was not set", vaultName)
	}

	if _, ok := storage.Vaults[vaultName].Items[itemName]; !ok {
		t.Errorf("Item %s was not set in vault %s", itemName, vaultName)
	}

	// Test getVaultItem
	retrievedItem, err := storage.getVaultItem(vaultName, itemName)
	if err != nil {
		t.Errorf("Error retrieving item: %s", err)
	}

	if !reflect.DeepEqual(retrievedItem, testItem) {
		t.Errorf("Retrieved item does not match set item. Got %v, want %v", retrievedItem, testItem)
	}
}

func TestOpFieldMatchField(t *testing.T) {
	field := opField{
		ID:      "field-id",
		Type:    "field-type",
		Label:   "field-label",
		Value:   "field-value",
		Section: opSection{ID: "section-id", Label: "section-label"},
	}

	// Test when the ID matches but the section does not
	uri, _ := NewOpURI("op://vault/item/field-id")
	uri.section = "wrong-section"
	if field.matchField(uri) {
		t.Errorf("Expected field not to match URI, but it did")
	}

	// Test when the Label matches but the section does not
	uri.field = "field-label"
	if field.matchField(uri) {
		t.Errorf("Expected field not to match URI, but it did")
	}

	// Test when the section matches but neither the ID nor the Label does
	uri.field = "wrong-field"
	uri.section = "section-id"
	if field.matchField(uri) {
		t.Errorf("Expected field not to match URI, but it did")
	}

	// Test when both the ID and section match
	uri.field = "field-id"
	if !field.matchField(uri) {
		t.Errorf("Expected field to match URI, but it did not")
	}

	// Test when both the Label and section match
	uri.field = "field-label"
	if !field.matchField(uri) {
		t.Errorf("Expected field to match URI, but it did not")
	}

	// Test when section is empty
	uri.section = ""
	field.Section.ID = "add more"
	if !field.matchField(uri) {
		t.Errorf("Expected field to match URI, but it did not")
	}
}

func TestOpFileMatchFile(t *testing.T) {
	file := opFile{
		ID:      "file-id",
		Name:    "file-name",
		Section: opSection{ID: "section-id", Label: "section-label"},
	}

	// Test when the ID matches but the section does not
	uri := &OpURI{
		raw:     "op://vault/item/file-id",
		vault:   "vault",
		item:    "item",
		field:   "file-id",
		section: "wrong-section",
	}
	if file.matchFile(uri) {
		t.Errorf("Expected file not to match URI, but it did")
	}

	// Test when the Name matches but the section does not
	uri.field = "file-name"
	if file.matchFile(uri) {
		t.Errorf("Expected file not to match URI, but it did")
	}

	// Test when the section matches but neither the ID nor the Name does
	uri.field = "wrong-field"
	uri.section = "section-id"
	if file.matchFile(uri) {
		t.Errorf("Expected file not to match URI, but it did")
	}

	// Test when both the ID and section match
	uri.field = "file-id"
	if !file.matchFile(uri) {
		t.Errorf("Expected file to match URI, but it did not")
	}

	// Test when both the Name and section match
	uri.field = "file-name"
	if !file.matchFile(uri) {
		t.Errorf("Expected file to match URI, but it did not")
	}

	// Test when section is empty
	uri.section = ""
	file.Section.ID = "add more"
	if !file.matchFile(uri) {
		t.Errorf("Expected field to match URI, but it did not")
	}
}

func TestGetFieldValue(t *testing.T) {
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
	// Test when the field exists and matches the URI
	uri, _ := NewOpURI("op://vault/item/section-id/field-id")
	value, err := item.GetFieldValue(nil, uri)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if value != "field-value" {
		t.Errorf("Expected field value to be 'field-value', but got '%s'", value)
	}

	// Test when the field does not exist in the item
	uri.field = "nonexistent-field"
	_, err = item.GetFieldValue(nil, uri)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	// Test when the field exists but the section does not match
	uri.field = "field-id"
	uri.section = "wrong-section"
	_, err = item.GetFieldValue(nil, uri)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	// Test when the field is a file
	uri, _ = NewOpURI("op://vault/item/section-id/file-id")
	executor := &SpyCommandExecutor{IsCliInstalled: true, ExecuteOutput: []byte("itsa me - a file!")}
	cli, err := New1Password(executor, "")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	fieldValue, _ := item.GetFieldValue(cli, uri)
	assert.Equal(t, []string{"read", "op://vault/item/section-id/file-id"}, executor.ExecuteArgs)
	assert.Equal(t, "itsa me - a file!", fieldValue)

	// Test when the fetching file fails
	uri, _ = NewOpURI("op://vault/item/section-id/file-id")
	executor = &SpyCommandExecutor{IsCliInstalled: true, ExecuteError: errors.New("oh noez")}
	cli, err = New1Password(executor, "")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	_, err = item.GetFieldValue(cli, uri)
	assert.Equal(t, []string{"read", "op://vault/item/section-id/file-id"}, executor.ExecuteArgs)
	assert.Equal(t, "oh noez", err.Error())
}
