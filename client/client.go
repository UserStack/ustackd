package client

import (
	"fmt"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"

	"github.com/UserStack/ustackd/backends"
)

/* parts of the code are taken from smtp.go from the core library */

type Client struct {
	mutex sync.Mutex
	Text *textproto.Conn
	conn net.Conn
}

// Dial returns a new Client connected to an ustack server at addr.
// The addr must include a port number.
func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host)
}

// NewClient returns a new Client using an existing connection and host as a
// server name to be used when authenticating.
func NewClient(conn net.Conn, host string) (*Client, error) {
	text := textproto.NewConn(conn)
	line, err := text.ReadLine()
	if err != nil {
		text.Close()
		return nil, err
	}
	if !strings.Contains(line, "ustack") {
		text.Close()
		return nil, fmt.Errorf("Not a ustackd server")
	}
	c := &Client{Text: text, conn: conn}
	return c, nil
}

func (client *Client) CreateUser(name string, password string) (int64, *backends.Error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("user %s %s", name, password)
	if err != nil {
		return 0, &backends.Error{"EFAULT", err.Error()}
	}
	return client.handleIntResponse()
}

func (client *Client) DisableUser(nameuid string) *backends.Error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("disable %s", nameuid)
	if err != nil {
		return &backends.Error{"EFAULT", err.Error()}
	}
	return client.handleResponse()
}

func (client *Client) EnableUser(nameuid string) *backends.Error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("enable %s", nameuid)
	if err != nil {
		return &backends.Error{"EFAULT", err.Error()}
	}
	return client.handleResponse()
}

func (client *Client) SetUserData(nameuid string, key string, value string) *backends.Error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("set %s %s %s", nameuid, key, value)
	if err != nil {
		return &backends.Error{"EFAULT", err.Error()}
	}
	return client.handleResponse()
}

func (client *Client) GetUserData(nameuid string, key string) (string, *backends.Error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("get %s %s", nameuid, key)
	if err != nil {
		return "", &backends.Error{"EFAULT", err.Error()}
	}
	line, rerr := client.Text.ReadLine()
	if rerr != nil {
		return "", &backends.Error{"EFAULT", rerr.Error()}
	}
	if strings.HasPrefix(line, "- E") {
		ret := strings.Split(line, " ")
		return "", &backends.Error{ret[1], "remote failure"}
	}
	herr := client.handleResponse()
	return line, herr
}

func (client *Client) LoginUser(name string, password string) (int64, *backends.Error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("login %s %s", name, password)
	if err != nil {
		return 0, &backends.Error{"EFAULT", err.Error()}
	}
	return client.handleIntResponse()
}

func (client *Client) ChangeUserPassword(nameuid string, password string, newpassword string) *backends.Error {
	return nil
}

func (client *Client) ChangeUserName(nameuid string, password string, newname string) *backends.Error {
	return nil
}

func (client *Client) UserGroups(nameuid string) ([]backends.Group, *backends.Error) {
	return nil, nil
}

func (client *Client) DeleteUser(nameuid string) *backends.Error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("delete user %s", nameuid)
	if err != nil {
		return &backends.Error{"EFAULT", err.Error()}
	}
	return client.handleResponse()
}

func (client *Client) Users() ([]backends.User, *backends.Error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	_, err := client.Text.Cmd("users")
	if err != nil {
		return nil, &backends.Error{"EFAULT", err.Error()}
	}
	var users []backends.User
	for {
		line, rerr := client.Text.ReadLine()
		if rerr != nil {
			return nil, &backends.Error{"EFAULT", rerr.Error()}
		}
		if strings.HasPrefix(line, "- E") {
			ret := strings.Split(line, " ")
			return nil, &backends.Error{ret[1], "remote failure"}
		} else if strings.HasPrefix(line, "+ ") {
			return users, nil
		}
		args := strings.Split(line, ":")
		if len(args) != 2 {
			return nil, &backends.Error{"EFAULT", "expected two values: " + line}
		}
		uid, perr:= strconv.ParseInt(args[1], 10, 64)
		if perr != nil {
			return nil, &backends.Error{"EFAULT", perr.Error()}
		}
		users = append(users, backends.User{
			Uid: uid,
			Name: args[0],
		})
	}
}

func (client *Client) Group(name string) (int64, *backends.Error) {
	return 0, nil
}

func (client *Client) AddUserToGroup(nameuid string, groupgid string) *backends.Error {
	return nil
}

func (client *Client) RemoveUserFromGroup(nameuid string, groupgid string) *backends.Error {
	return nil
}

func (client *Client) DeleteGroup(groupgid string) *backends.Error {
	return nil
}

func (client *Client) Groups() ([]backends.Group, *backends.Error) {
	return nil, nil
}

func (client *Client) GroupUsers(groupgid string) ([]backends.User, *backends.Error) {
	return nil, nil
}

func (client *Client) Close() {
	client.Text.Close()
}

// Helpers

func (client *Client) handleIntResponse() (int64, *backends.Error) {
	line, rerr := client.Text.ReadLine()
	if rerr != nil {
		return 0, &backends.Error{"EFAULT", rerr.Error()}
	}
	ret := strings.Split(line, " ")
	if ret[0] == "-" {
		return 0, &backends.Error{ret[1], "remote failure"}
	}
	val, perr := strconv.ParseInt(ret[2], 10, 64)
	if perr != nil {
		return 0, &backends.Error{"EFAULT", perr.Error()}
	}
	return val, nil
}

func (client *Client) handleResponse() *backends.Error {
	line, rerr := client.Text.ReadLine()
	if rerr != nil {
		return &backends.Error{"EFAULT", rerr.Error()}
	}
	ret := strings.Split(line, " ")
	if ret[0] == "-" {
		return &backends.Error{ret[1], "remote failure"}
	}
	return nil
}
