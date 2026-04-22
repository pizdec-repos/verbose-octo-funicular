package express

import (
	"context"
	"fmt"
)

type Client struct {
	apiToken string
}

func NewClient(token string) *Client {
	return &Client{apiToken: token}
}

func (c *Client) SendMessage(ctx context.Context, text string) error {
	fmt.Printf("[eXpress] Push Alert: %s\n", text)
	return nil
}
