package gonepassword

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SpyCommandExecutor struct {
	IsExecuteCalled bool
	IsCliInstalled  bool
	ExecuteArgs     []string
	ExecuteError    error
	ExecuteOutput   []byte
}

func (e *SpyCommandExecutor) IsInstalled() bool {
	return e.IsCliInstalled
}

func (e *SpyCommandExecutor) Execute(arg ...string) ([]byte, error) {
	e.ExecuteArgs = arg
	e.IsExecuteCalled = true
	return e.ExecuteOutput, e.ExecuteError
}

func TestResolveOpURI(t *testing.T) { //nolint
	opItem := opItem{
		ID: "item",
		Fields: []opField{
			{
				ID:    "field-id",
				Label: "field-label",
				Value: "resolved-value",
			},
		},
	}
	opItemJSON, err := json.Marshal(opItem)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		executor       *SpyCommandExecutor
		uri            string
		expectedError  string
		expectedOutput string
		expectedArgs   []string
	}{
		{
			name:          "should return error when uri is too short",
			executor:      &SpyCommandExecutor{IsCliInstalled: true},
			uri:           "op://vault/item",
			expectedError: "invalid 1Password URI format - expected op://vault/item/field - got 'op://vault/item'",
		},
		{
			name:     "should return error when uri is too long",
			executor: &SpyCommandExecutor{IsCliInstalled: true},
			uri:      "op://vault/item/section/field/extra",
			expectedError: "invalid 1Password URI format - expected op://vault/item/field - " +
				"got 'op://vault/item/section/field/extra'",
		},
		{
			name:     "should return error when op binary is not present",
			executor: &SpyCommandExecutor{IsCliInstalled: false},
			uri:      "op://vault/item/field",
			expectedError: "1Password CLI is not installed, visit https://support.1password.com/command-line/ " +
				"for installation instructions",
		},
		{
			name:           "should return error when it's not an op uri and op binary is not present",
			executor:       &SpyCommandExecutor{IsCliInstalled: false},
			uri:            "regular-value",
			expectedOutput: "",
			expectedError:  "incorrect op uri - it should look like op://vault/item/field - got regular-value",
		},
		{
			name:          "should return error when op command fails",
			executor:      &SpyCommandExecutor{IsCliInstalled: true, ExecuteError: errors.New("command failed")},
			uri:           "op://vault/item/field",
			expectedError: "command failed",
		},
		{
			name: "should return resolved value by id",
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
				ExecuteOutput:  opItemJSON,
			},
			expectedArgs:   []string{"item", "get", "--format", "json", "item", "--vault", "vault"},
			uri:            "op://vault/item/field-id",
			expectedOutput: "resolved-value",
		},
		{
			name: "should return resolved value by label",
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
				ExecuteOutput:  opItemJSON,
			},
			expectedArgs:   []string{"item", "get", "--format", "json", "item", "--vault", "vault"},
			uri:            "op://vault/item/field-label",
			expectedOutput: "resolved-value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cli *OnePassword
			var err error
			var result string

			if cli, err = New1Password(tc.executor, ""); err == nil {
				result, err = cli.ResolveOpURI(tc.uri)
			}

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, result)
				assert.Equal(t, tc.expectedArgs, tc.executor.ExecuteArgs)
			}
		})
	}
}

func Test1PasswordStorage(t *testing.T) {
	opItem := opItem{
		ID: "item",
		Fields: []opField{
			{
				ID:    "field-id",
				Label: "field-label",
				Value: "resolved-value",
			},
			{
				ID:    "another-id",
				Label: "another-label",
				Value: "another-value",
			},
			{
				ID:    "third-id",
				Label: "third-label",
				Value: "third-value",
			},
		},
	}
	opItemJSON, err := json.Marshal(opItem)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		executor       *SpyCommandExecutor
		uris           []string
		expectedOutput string
		executorCalls  int
	}{
		{
			name: "should call cli once while fetching multiple fields",
			executor: &SpyCommandExecutor{
				IsCliInstalled: true,
				ExecuteOutput:  opItemJSON,
			},
			uris: []string{
				"op://vault/item/field-id",
				"op://vault/item/another-id",
				"op://vault/item/third-id",
			},
			executorCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cli *OnePassword
			var err error
			executeCalls := 0

			if cli, err = New1Password(tc.executor, ""); err == nil {
				for _, uri := range tc.uris {
					_, err = cli.ResolveOpURI(uri)
					if tc.executor.IsExecuteCalled {
						executeCalls++
						tc.executor.IsExecuteCalled = false
					}
				}
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.executorCalls, executeCalls)
		})
	}
}

func TestNewOpURI(t *testing.T) {
	testCases := []struct {
		name          string
		uri           string
		expectedError string
		expectedOpURI *OpURI
	}{
		{
			name:          "should return error when uri is too short",
			uri:           "op://vault/item",
			expectedError: "invalid 1Password URI format - expected op://vault/item/field - got 'op://vault/item'",
		},
		{
			name: "should return error when uri is too long",
			uri:  "op://vault/item/section/field/extra",
			expectedError: "invalid 1Password URI format - expected op://vault/item/field - " +
				"got 'op://vault/item/section/field/extra'",
		},
		{
			name: "should return valid OpURI when uri is valid with three parts",
			uri:  "op://vault/item/field",
			expectedOpURI: &OpURI{
				raw:     "op://vault/item/field",
				vault:   "vault",
				item:    "item",
				field:   "field",
				section: "",
			},
		},
		{
			name: "should return valid OpURI when uri is valid with four parts",
			uri:  "op://vault/item/section/field",
			expectedOpURI: &OpURI{
				raw:     "op://vault/item/section/field",
				vault:   "vault",
				item:    "item",
				field:   "field",
				section: "section",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewOpURI(tc.uri)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOpURI, result)
			}
		})
	}
}
