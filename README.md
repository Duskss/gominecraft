# gominecraft

A library to access unconventional endpoints on Mojang via Go.

## Installation

Use the package manager [go](https://golang.org/) to install **gominecraft**.

```bash
go get github.com/discoli/gominecraft
```

## Usage

```python
import "github.com/discoli/gominecraft

c := gominecraft.Client{} # Initialize the client
s, err := c.LoginC("email", "pass", "captcha") # Authenticate, see c.Login for an alternative
err = s.Allocate("Why") # Allocate a username to a non-premium Minecraft account
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.
