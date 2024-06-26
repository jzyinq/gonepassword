# gonepassword

This project provides a wrapper around
the [1Password](https://1password.com) [CLI client](https://developer.1password.com/docs/cli/) to fetch secrets from
1password vaults.

Unlike official [connect-sdk-go](https://github.com/1Password/connect-sdk-go) it utilize 1password cli client either
through
[1password native app](https://1password.com/downloads/)
or `SERVICE_ACCOUNT_TOKEN` ([source](https://developer.1password.com/docs/service-accounts/use-with-1password-cli).))

## Getting Started

### Prerequisites

Those are the requirements to use this tool without service account token:

- [1 password cli client](https://developer.1password.com/docs/cli/get-started/#step-1-install-1password-cli)
- [1 password native app](https://1password.com/downloads/)

In native app settings you have to make some changes:

- under `Security` section, enable `Unlock using system authentication service`
- under `Developer` section, enable `Integrate with 1Password CLI` option

### Installing

Clone the repository:

```bash
go get -v -u github.com/jzyinq/gonepassword
```

## Usage

This project provides a wrapper around the 1Password CLI. It includes functionalities such as checking if the 1Password
CLI is installed, executing commands, and resolving 1Password URIs.

Here is a basic example of how to use it:

```go
package main

import (
	"fmt"
	"log"

	"github.com/jzyinq/gonepassword"
)

func main() {
	// Create a new 1Password client
	opCli, err := gonepassword.New1Password(nil, gonepassword.OnePasswordOptions{})
	if err != nil {
		log.Fatal(err)
	}

	uri := "op://vault/item/field"
	value, err := opCli.ResolveOpURI(uri)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(value)
}
```

## Running the tests

To run the tests, use the following command:

```bash
make tests
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details