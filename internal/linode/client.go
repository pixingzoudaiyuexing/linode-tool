package linode

import (
	"errors"
	"os"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func NewClient() (*linodego.Client, error) {
	token := os.Getenv("LINODE_TOKEN")
	if token == "" {
		return nil, errors.New("missing LINODE_TOKEN environment variable")
	}

	oauth := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(oauth2.NoContext, oauth)
	return linodego.NewClient(httpClient), nil
}
