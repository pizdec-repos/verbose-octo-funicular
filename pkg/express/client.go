package express

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/token"
	"go.uber.org/zap"
)

type Client struct {
	host       string
	chatID     string
	tokenGen   token.Generator
	httpClient *http.Client
	logger     *zap.Logger
}

func NewClient(host, chatID string, tokenGen token.Generator, logger *zap.Logger) *Client {
	return &Client{
		host:     host,
		chatID:   chatID,
		tokenGen: tokenGen,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		},
		logger: logger,
	}
}

type NotificationPayload struct {
	GroupChatID  string       `json:"group_chat_id"`
	Notification Notification `json:"notification"`
}

type Notification struct {
	Body string `json:"body"`
}

func (c *Client) SendAlert(ctx context.Context, text string) error {
	jwtToken, err := c.tokenGen.Generate()
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	payload := NotificationPayload{
		GroupChatID:  c.chatID,
		Notification: Notification{Body: text},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/api/v4/botx/notifications/direct", c.host)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			c.logger.Warn("retrying alert send",
				zap.Int("attempt", attempt),
				zap.Error(lastErr),
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("send request: %w", err)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 300 {
			c.logger.Debug("alert delivered",
				zap.String("host", c.host),
				zap.Int("status", resp.StatusCode),
			)
			return nil
		}

		lastErr = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}
	}

	return fmt.Errorf("all retries exhausted: %w", lastErr)
}
