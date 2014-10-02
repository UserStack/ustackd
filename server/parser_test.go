package server

import (
	"reflect"
	"testing"
)

func TestOneParameterCommands(t *testing.T) {
	cmd, args := parseCmd("disable username")
	if cmd != DISABLE || !reflect.DeepEqual(args, []string{"username"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("enable username")
	if cmd != ENABLE || !reflect.DeepEqual(args, []string{"username"}) {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestClientCommands(t *testing.T) {
	cmd, args := parseCmd("client auth secret")
	if cmd != CLIENT_AUTH || !reflect.DeepEqual(args, []string{"secret"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("CLIent Auth Secret")
	if cmd != CLIENT_AUTH || !reflect.DeepEqual(args, []string{"Secret"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("client foo")
	if cmd != ERR_UNKNOWN_FUNC {
		t.Fatal("failed to parse")
	}

	cmd, args = parseCmd("client")
	if cmd != ERR_UNKNOWN_FUNC {
		t.Fatal("failed to parse")
	}

	cmd, args = parseCmd("client auth")
	if cmd != ERR_MISSING_ARGS {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestChangeCommands(t *testing.T) {
	cmd, args := parseCmd("change password username secret newsecret")
	if cmd != CHANGE_PASSWORD || !reflect.DeepEqual(args, []string{"username", "secret", "newsecret"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("change name username secret newname")
	if cmd != CHANGE_NAME || !reflect.DeepEqual(args, []string{"username", "secret", "newname"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("change foo")
	if cmd != ERR_UNKNOWN_FUNC {
		t.Fatal("failed to parse")
	}

	cmd, args = parseCmd("change")
	if cmd != ERR_UNKNOWN_FUNC {
		t.Fatal("failed to parse")
	}

	cmd, args = parseCmd("change password")
	if cmd != ERR_MISSING_ARGS {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestNoArgCommands(t *testing.T) {
	cmd, args := parseCmd("quit")
	if cmd != QUIT {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("stats")
	if cmd != STATS {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("groups")
	if cmd != GROUPS {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("users")
	if cmd != USERS {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestTwoParameterCommands(t *testing.T) {
	cmd, args := parseCmd("login")
	if cmd != ERR_MISSING_ARGS {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("login username")
	if cmd != ERR_MISSING_ARGS {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("login username secret")
	if cmd != LOGIN || !reflect.DeepEqual(args, []string{"username", "secret"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("get username key")
	if cmd != GET || !reflect.DeepEqual(args, []string{"username", "key"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("add username group")
	if cmd != ADD || !reflect.DeepEqual(args, []string{"username", "group"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("remove username group")
	if cmd != REMOVE || !reflect.DeepEqual(args, []string{"username", "group"}) {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestUserCommands(t *testing.T) {
	cmd, args := parseCmd("user username secret")
	if cmd != USER || !reflect.DeepEqual(args, []string{"username", "secret"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("user groups username")
	if cmd != USER_GROUPS || !reflect.DeepEqual(args, []string{"username"}) {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestGroupCommands(t *testing.T) {
	cmd, args := parseCmd("group name")
	if cmd != GROUP || !reflect.DeepEqual(args, []string{"name"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("group users name")
	if cmd != GROUP_USERS || !reflect.DeepEqual(args, []string{"name"}) {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestDeleteCommands(t *testing.T) {
	cmd, args := parseCmd("delete group name")
	if cmd != DELETE_GROUP || !reflect.DeepEqual(args, []string{"name"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("delete group")
	if cmd != ERR_MISSING_ARGS {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("delete user name")
	if cmd != DELETE_USER || !reflect.DeepEqual(args, []string{"name"}) {
		t.Fatal("failed to parse", cmd, args)
	}

	cmd, args = parseCmd("delete user")
	if cmd != ERR_MISSING_ARGS {
		t.Fatal("failed to parse", cmd, args)
	}
}

func TestThreeParameterCommands(t *testing.T) {
	cmd, args := parseCmd("set username key value")
	if cmd != SET || !reflect.DeepEqual(args, []string{"username", "key", "value"}) {
		t.Fatal("failed to parse", cmd, args)
	}
}
