// Copyright © 2020 Emando B.V.

package hub

import (
	"context"
	"encoding/json"

	"github.com/emando/vantage-events/pkg/events"
	"github.com/gorilla/websocket"
)

// Client is a Hub client.
type Client struct {
	conn *websocket.Conn
}

// Connect connects to the Hub and returns a Client.
func Connect(ctx context.Context, url string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	conn.SetPingHandler(nil)
	return &Client{
		conn: conn,
	}, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Read reads events from the Hub.
func (c *Client) Read(ctx context.Context, ch chan<- *events.Raw) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		messageType, buf, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		switch messageType {
		case websocket.TextMessage:
			event := &events.Raw{
				Bytes: buf,
			}
			if err := json.Unmarshal(buf, &event); err != nil {
				return err
			}
			ch <- event
		}
	}
}
