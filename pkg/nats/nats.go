// Copyright © 2019 Emando B.V.

package nats

import (
	nats "github.com/nats-io/go-nats"
	stan "github.com/nats-io/go-nats-streaming"
)

// Options contains options for NATS streaming.
type Options struct {
	URL,
	Username,
	Password string
	UseTLS bool
	ClusterID,
	ClientID string
}

// Conn is a connection to NATS Streaming Server.
type Conn struct {
	nats *nats.Conn
	stan stan.Conn
}

// Connect connects to NATS Streaming Server.
func Connect(opts Options) (*Conn, error) {
	var options []nats.Option
	if opts.Username != "" {
		options = append(options, nats.UserInfo(opts.Username, opts.Password))
	}
	if opts.UseTLS {
		options = append(options, nats.Secure())
	}
	natsConn, err := nats.Connect(opts.URL, options...)
	if err != nil {
		return nil, err
	}
	stanConn, err := stan.Connect(opts.ClusterID, opts.ClientID, stan.NatsConn(natsConn))
	if err != nil {
		natsConn.Close()
		return nil, err
	}
	return &Conn{
		nats: natsConn,
		stan: stanConn,
	}, nil
}

// Close closes the connection.
func (c *Conn) Close() error {
	if err := c.stan.Close(); err != nil {
		return err
	}
	c.nats.Close()
	return nil
}
