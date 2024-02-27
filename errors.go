package gonepassword

import "fmt"

type InvalidOpURIError struct {
	uri string
}

func (e InvalidOpURIError) Error() string {
	return fmt.Sprintf("incorrect op uri - it should look like op://vault/item/field - got %s", e.uri)
}

type OnePasswordCliNotInstalledError struct {
}

func (e OnePasswordCliNotInstalledError) Error() string {
	return "1Password CLI is not installed, visit https://support.1password.com/command-line/ for installation instructions"
}
