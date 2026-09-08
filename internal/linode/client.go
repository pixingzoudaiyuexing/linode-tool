package linode

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

const apiTimeout = 60 * time.Second

// Client contains the Linode API operations used by the CLI.
type Client interface {
	ListRegions(context.Context, *linodego.ListOptions) ([]linodego.Region, error)
	ListInstances(context.Context, *linodego.ListOptions) ([]linodego.Instance, error)
	CreateInstance(context.Context, linodego.InstanceCreateOptions) (*linodego.Instance, error)
	DeleteInstance(context.Context, int) error
	CreateFirewall(context.Context, linodego.FirewallCreateOptions) (*linodego.Firewall, error)
}

func NewClient(token string) (Client, error) {
	if token == "" {
		return nil, errors.New("Linode API Token 不能为空")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := &http.Client{
		Timeout: apiTimeout,
		Transport: &oauth2.Transport{
			Source: tokenSource,
			Base:   http.DefaultTransport,
		},
	}
	client := linodego.NewClient(httpClient)
	return &client, nil
}
