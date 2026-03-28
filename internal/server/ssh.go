package server

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client wraps an SSH connection to a remote server.
type Client struct {
	conn *ssh.Client
	Host string
	User string
}

// Connect establishes an SSH connection using a private key file.
func Connect(host, user, keyPath string) (*Client, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("ssh: read key %s: %w", keyPath, err)
	}
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("ssh: parse key: %w", err)
	}

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: replace with known_hosts verification
		Timeout:         15 * time.Second,
	}

	conn, err := ssh.Dial("tcp", host+":22", cfg)
	if err != nil {
		return nil, fmt.Errorf("ssh: dial %s: %w", host, err)
	}
	return &Client{conn: conn, Host: host, User: user}, nil
}

// Run executes a command on the remote host and returns combined stdout+stderr.
func (c *Client) Run(cmd string) (string, error) {
	sess, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh: new session: %w", err)
	}
	defer sess.Close()

	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("ssh: run %q: %w\n%s", cmd, err, out)
	}
	return string(out), nil
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
