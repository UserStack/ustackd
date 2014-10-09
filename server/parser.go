package server

import (
	"strings"
)

type Command int

const (
	CLIENT_AUTH Command = iota
	QUIT
	LOGIN
	DISABLE
	ENABLE
	SET
	GET
	CHANGE_PASSWORD
	CHANGE_NAME
	USER_GROUPS
	USER
	DELETE_USER
	USERS
	ADD
	REMOVE
	DELETE_GROUP
	GROUPS
	GROUP_USERS
	GROUP
	STATS
	LOGINSTATS

	ERR_UNKNOWN_FUNC
	ERR_MISSING_ARGS
	ERR_INVALID_ARGS
)

var NOARGS []string

type SubParser func(line string) (Command, []string)

func parseCmd(line string) (Command, []string) {
	parts := strings.SplitN(line, " ", 2)
	switch strings.ToLower(parts[0]) {
	case "login":
		return parseTwoArgumentCmd(LOGIN, parts)
	case "set":
		return parseThreeArgumentCmd(SET, parts)
	case "get":
		return parseTwoArgumentCmd(GET, parts)
	case "stats":
		return STATS, NOARGS
	case "loginstats":
		return parseOneArgumentCmd(LOGINSTATS, parts)
	case "add":
		return parseTwoArgumentCmd(ADD, parts)
	case "remove":
		return parseTwoArgumentCmd(REMOVE, parts)
	case "enable":
		return parseOneArgumentCmd(ENABLE, parts)
	case "disable":
		return parseOneArgumentCmd(DISABLE, parts)
	case "user":
		return expectTwoParts(parts, parseUserCmd)
	case "group":
		return expectTwoParts(parts, parseGroupCmd)
	case "delete":
		return expectTwoParts(parts, parseDeleteCmd)
	case "change":
		return expectTwoParts(parts, parseChangeCmd)
	case "client":
		return expectTwoParts(parts, parseClientCmd)
	case "quit":
		return QUIT, NOARGS
	case "groups":
		return GROUPS, NOARGS
	case "users":
		return USERS, NOARGS
	}
	return ERR_UNKNOWN_FUNC, NOARGS
}

func parseClientCmd(line string) (Command, []string) {
	parts := strings.SplitN(line, " ", 2)
	switch strings.ToLower(parts[0]) {
	case "auth":
		return parseOneArgumentCmd(CLIENT_AUTH, parts)
	}
	return ERR_UNKNOWN_FUNC, NOARGS
}

func parseChangeCmd(line string) (Command, []string) {
	parts := strings.SplitN(line, " ", 2)
	switch strings.ToLower(parts[0]) {
	case "password":
		return parseThreeArgumentCmd(CHANGE_PASSWORD, parts)
	case "name":
		return parseThreeArgumentCmd(CHANGE_NAME, parts)
	}
	return ERR_UNKNOWN_FUNC, NOARGS
}

func parseUserCmd(line string) (Command, []string) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return ERR_MISSING_ARGS, NOARGS
	}
	switch strings.ToLower(parts[0]) {
	case "groups":
		return USER_GROUPS, []string{parts[1]}
	default:
		return USER, parts
	}
}

func parseDeleteCmd(line string) (Command, []string) {
	parts := strings.SplitN(line, " ", 2)
	switch strings.ToLower(parts[0]) {
	case "user":
		return parseOneArgumentCmd(DELETE_USER, parts)
	case "group":
		return parseOneArgumentCmd(DELETE_GROUP, parts)
	}
	return ERR_UNKNOWN_FUNC, NOARGS
}

func parseGroupCmd(line string) (Command, []string) {
	parts := strings.SplitN(line, " ", 2)
	switch strings.ToLower(parts[0]) {
	case "users":
		if len(parts) != 2 {
			return ERR_MISSING_ARGS, NOARGS
		}
		return GROUP_USERS, []string{parts[1]}
	default:
		return GROUP, parts
	}
}

func expectTwoParts(parts []string, fn SubParser) (Command, []string) {
	if len(parts) != 2 {
		return ERR_UNKNOWN_FUNC, NOARGS
	}
	return fn(parts[1])
}

func parseOneArgumentCmd(cmd Command, parts []string) (Command, []string) {
	if len(parts) != 2 {
		return ERR_MISSING_ARGS, NOARGS
	}
	return cmd, []string{parts[1]}
}

func parseTwoArgumentCmd(cmd Command, parts []string) (Command, []string) {
	if len(parts) != 2 {
		return ERR_MISSING_ARGS, NOARGS
	}
	parts = strings.SplitN(parts[1], " ", 2)
	if len(parts) != 2 {
		return ERR_MISSING_ARGS, NOARGS
	}
	return cmd, parts
}

func parseThreeArgumentCmd(cmd Command, parts []string) (Command, []string) {
	if len(parts) != 2 {
		return ERR_MISSING_ARGS, NOARGS
	}
	parts = strings.SplitN(parts[1], " ", 3)
	if len(parts) != 3 {
		return ERR_MISSING_ARGS, NOARGS
	}
	return cmd, parts
}
