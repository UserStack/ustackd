package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/client"
	"github.com/UserStack/ustackd/server"
)

var serverInstance *server.Server
var started bool

func newClient() (conn *client.Client) {
	if started == false {
		serverInstance = server.NewServer()
		var configPath string
		if configPath = os.Getenv("TEST_CONFIG"); configPath == "" {
			configPath = "config/test_sqlite.conf"
		}
		go (func() {
			serverInstance.Run([]string{
				"./ustackd", "-f", "-c", configPath,
			})
		})()
		started = true
	}
	// try to connect 25 times in with 100ms interval until giving up = max 2.5s
	var err error
	for i := 0; i < 25; i++ {
		conn, err = client.Dial("localhost:35786")
		if err != nil {
			time.Sleep(100 * time.Millisecond)
		} else {
			return
		}
	}
	panic("client was unable to connect to server: " + err.Error())
}

func uniqName() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}

func TestConnect(t *testing.T) {
	client := newClient()
	defer client.Close()
}

func TestConnectTls(t *testing.T) {
	client := newClient()
	defer client.Close()
	aerr := client.StartTlsWithCert("config/cert.pem")
	if aerr != nil {
		t.Fatal("unable to establish tls", aerr)
	}
	username := uniqName()
	id, serr := client.CreateUser(username, "secret")
	defer client.DeleteUser(username)
	if id <= 0 {
		t.Fatal("user not created, expected id bigger than 0 got", id, serr)
	}
}

func TestCreateUser(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	defer client.DeleteUser(username)
	id, err := client.CreateUser(username, "secret")
	if id <= 0 {
		t.Fatal("user not created, expected id bigger than 0 got", id, err)
	}
}

func TestDisableUser(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	err := client.DisableUser(username)
	if err != nil {
		t.Fatalf("unable to disable user, %v", err)
	}
	uid, lerr := client.LoginUser(username, "secret")
	if uid != 0 {
		t.Fatalf("should not be able to login, %v", lerr)
	}
}

func TestEnableUser(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	derr := client.DisableUser(username)
	if derr != nil {
		t.Fatalf("unable to disable user, %v", derr)
	}
	uid, lerr := client.LoginUser(username, "secret")
	if uid != 0 {
		t.Fatalf("should not be able to login, %v", lerr)
	}
	err := client.EnableUser(username)
	if err != nil {
		t.Fatalf("unable to enable user, %v", err)
	}
	uid, serr := client.LoginUser(username, "secret")
	if uid == 0 {
		t.Fatalf("should be able to login, %v", serr)
	}
}

func TestSetGetUserData(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	client.SetUserData(username, "firstname", "Tester")
	firstname, err := client.GetUserData(username, "firstname")
	if firstname != "Tester" {
		t.Fatal("firstname should have been 'Tester' but was", firstname, err)
	}
}

func TestLoginUser(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	_, lerr := client.LoginUser(username, "secret")
	if lerr != nil {
		t.Fatal("should be able to login", lerr)
	}
	client.DisableUser(username)
	_, lerr1 := client.LoginUser(username, "secret")
	if lerr1 == nil {
		t.Fatal("should not be able to login, but was able -> not disabled")
	}
	client.EnableUser(username)
	_, lerr2 := client.LoginUser(username, "secret")
	if lerr2 != nil {
		t.Fatal("should be able to login again", lerr2)
	}
}

func TestChangeUserPassword(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()

	ierr1 := client.ChangeUserPassword("", "secret2", "secret2")
	if ierr1.Code != "EINVAL" {
		t.Fatal("should fail since user value is empty")
	}

	ierr2 := client.ChangeUserPassword(username, "", "secret2")
	if ierr2.Code != "EINVAL" {
		t.Fatal("should fail since passwd value is empty")
	}

	ierr3 := client.ChangeUserPassword(username, "secret2", "")
	if ierr3.Code != "EINVAL" {
		t.Fatal("should fail since new passwd value is empty")
	}

	client.CreateUser(username, "secret")
	client.LoginUser(username, "secret")
	serr := client.ChangeUserPassword(username, "secret2", "secret2")
	// fails if password is wrong
	if serr.Code != "ENOENT" {
		t.Fatal("Should have failed to change with wrong password")
	}
	client.ChangeUserPassword(username, "secret", "secret2")
	_, err := client.LoginUser(username, "secret2")
	if err != nil {
		t.Fatalf("User passwd should have been changed: %s, (%s)", err.Code, err.Message)
	}
}

func TestChangeUserName(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	newusername := uniqName()

	ierr1 := client.ChangeUserName("", "secret2", newusername)
	if ierr1.Code != "EINVAL" {
		t.Fatal("should fail since name value is empty")
	}

	ierr2 := client.ChangeUserName(username, "", newusername)
	if ierr2.Code != "EINVAL" {
		t.Fatal("should fail since passwd value is empty")
	}

	ierr3 := client.ChangeUserName(username, "secret2", "")
	if ierr3.Code != "EINVAL" {
		t.Fatal("should fail since new name value is empty")
	}

	client.CreateUser(username, "secret")
	client.LoginUser(username, "secret")
	serr := client.ChangeUserName(username, "secret2", newusername)
	if serr.Code != "ENOENT" {
		t.Fatal("Should have failed to change with wrong password")
	}
	client.ChangeUserName(username, "secret", newusername)
	_, err := client.LoginUser(newusername, "secret")
	if err != nil {
		t.Fatalf("User name should have been changed: %s, (%s)", err.Code, err.Message)
	}
}

func TestDeleteUser(t *testing.T) {
	client := newClient()
	defer client.Close()
	username := uniqName()
	client.CreateUser(username, "secret")
	_, lerr := client.LoginUser(username, "secret")
	if lerr != nil {
		t.Fatal("should be able to login", lerr)
	}
	derr := client.DeleteUser(username)
	if derr != nil {
		t.Fatal("unable to delete the user", derr.Code)
	}
	_, lerr1 := client.LoginUser(username, "secret")
	if lerr1 == nil {
		t.Fatal("should not be able to login, but was able -> not deleted")
	}
}

func TestUsers(t *testing.T) {
	client := newClient()
	defer client.Close()
	username0 := uniqName()
	uid0, _ := client.CreateUser(username0, "secret")
	username1 := uniqName()
	uid1, _ := client.CreateUser(username1, "secret")
	users, _ := client.Users()
	if len(users) < 2 {
		t.Fatal("should have at least two users")
	}
	found0, found1 := false, false
	for _, user := range users {
		if user.Name == username0 && user.Uid == uid0 {
			found0 = true
		} else if user.Name == username1 && user.Uid == uid1 {
			found1 = true
		}
	}
	if found0 && found1 != true {
		t.Fatalf("should have found created users %s and %s in %v",
			username0, username1, users)
	}
}

func TestGroup(t *testing.T) {
	client := newClient()
	defer client.Close()
	group := uniqName()

	_, err := client.CreateGroup("")
	if err.Code != "EFAULT" {
		t.Fatal("should return EINVAL instead of", err.Code)
	}

	gid, cerr := client.CreateGroup(group)
	defer client.DeleteGroup(group)
	if gid <= 0 {
		t.Fatal("should have created group", cerr.Code, cerr.Message)
	}

	_, eerr := client.CreateGroup(group)
	if eerr.Code != "EEXIST" {
		t.Fatal("should return EEXIST instead of", eerr.Code, eerr.Message)
	}
}

func TestDeleteGroup(t *testing.T) {
	client := newClient()
	defer client.Close()

	groups, _ := client.Groups()
	groupCount := len(groups)

	group0 := uniqName()
	group1 := uniqName()

	err := client.DeleteUser("")
	if err.Code != "EFAULT" {
		t.Fatal("should error on missing parameter", err.Code)
	}

	// create two groups
	client.CreateGroup(group0)
	gid, _ := client.CreateGroup(group1)
	defer client.DeleteGroup(group1)

	// now has two users
	groups, _ = client.Groups()
	if diff := len(groups) - groupCount; diff != 2 {
		t.Fatalf("group count should have been increased by 2, but was increased by %d", diff)
	}

	// delete one using uid one using name
	derr1 := client.DeleteGroup(group0)
	if derr1 != nil {
		t.Fatal("should not error on delete", derr1.Code, derr1.Message)
	}
	derr1 = client.DeleteGroup(fmt.Sprintf("%d", gid))
	if derr1 != nil {
		t.Fatal("should not error on delete", derr1.Code, derr1.Message)
	}
	groups, _ = client.Groups()
	if len(groups) != groupCount {
		t.Fatalf("group count should have been back to %d but was %d", groupCount, len(groups))
	}
}

func TestGroups(t *testing.T) {
	client := newClient()
	defer client.Close()

	group0 := uniqName()
	group1 := uniqName()
	group2 := uniqName()

	client.CreateGroup(group0)
	defer client.DeleteGroup(group0)
	client.CreateGroup(group1)
	defer client.DeleteGroup(group1)
	client.CreateGroup(group2)
	defer client.DeleteGroup(group2)
	groups, _ := client.Groups()

	if len(groups) < 3 {
		t.Fatalf("expected to have at least 3 groups but got %v", groups)
	}
}

func TestUsersAndGroupsAssociations(t *testing.T) {
	client := newClient()
	defer client.Close()

	joe := uniqName()
	mike := uniqName()
	developers := uniqName()
	admins := uniqName()
	gid0, _ := client.CreateGroup(developers)
	defer client.DeleteGroup(developers)
	gid1, _ := client.CreateGroup(admins)
	defer client.DeleteGroup(admins)
	uid0, _ := client.CreateUser(joe, "secret")
	defer client.DeleteUser(joe)
	uid1, _ := client.CreateUser(mike, "secret")
	defer client.DeleteUser(mike)

	gid0s := fmt.Sprintf("%d", gid0)
	uid0s := fmt.Sprintf("%d", uid0)

	// User Groups
	groups, _ := client.UserGroups(joe)
	if len(groups) != 0 {
		t.Fatal("expected joe not to have any groups")
	}
	client.AddUserToGroup(joe, gid0s)
	client.AddUserToGroup(uid0s, admins)
	expectedGroups := []backends.Group{
		backends.Group{Gid: gid0, Name: developers},
		backends.Group{Gid: gid1, Name: admins},
	}
	groups, _ = client.UserGroups(joe)
	if !reflect.DeepEqual(expectedGroups, groups) {
		t.Fatalf("expected joe have groups %v but has %v", expectedGroups, groups)
	}

	// Group Users
	users, _ := client.GroupUsers(admins)
	if len(users) < 1 {
		t.Fatalf("expected to have one admin, got %v", users)
	}
	client.AddUserToGroup(mike, admins)
	users, _ = client.GroupUsers(admins)
	expectedUsers := []backends.User{
		backends.User{Uid: uid0, Name: joe},
		backends.User{Uid: uid1, Name: mike},
	}
	if reflect.DeepEqual(expectedUsers, users) {
		t.Fatalf("expected admins have users %v but has %v", expectedUsers, users)
	}

	// Remove associations
	client.RemoveUserFromGroup(mike, admins)
	client.RemoveUserFromGroup(joe, admins)
	adminsG, _ := client.GroupUsers(admins)
	if len(adminsG) != 0 {
		t.Fatalf("expected to have no admin, got %v", adminsG)
	}
	users, _ = client.GroupUsers(admins)
	if len(users) != 0 {
		t.Fatalf("expected to have no admin, got %v", users)
	}
}

func TestDeleteGroupWithAssociations(t *testing.T) {
	client := newClient()
	defer client.Close()

	username := uniqName()
	client.CreateUser(username, "secret")

	group := uniqName()
	client.CreateGroup(group)
	defer client.DeleteGroup(group)

	_, lerr := client.LoginUser(username, "secret")
	if lerr != nil {
		t.Fatal("should be able to login", lerr)
	}
	client.SetUserData(username, "foo", "bar")
	client.AddUserToGroup(username, group)
	derr := client.DeleteUser(username)
	if derr != nil {
		t.Fatal("unable to delete the user", derr.Code)
	}
	_, lerr1 := client.LoginUser(username, "secret")
	if lerr1 == nil {
		t.Fatal("should not be able to login, but was able -> not deleted")
	}
}

func TestDeleteUserWithAssociations(t *testing.T) {
	client := newClient()
	defer client.Close()

	username := uniqName()
	client.CreateUser(username, "secret")

	group := uniqName()
	client.CreateGroup(group)

	_, lerr := client.LoginUser(username, "secret")
	if lerr != nil {
		t.Fatal("should be able to login", lerr)
	}
	client.AddUserToGroup(username, group)
	derr := client.DeleteGroup(group)
	if derr != nil {
		t.Fatal("unable to delete the group", derr.Code)
	}
}

func TestStats(t *testing.T) {
	client := newClient()
	defer client.Close()
	users, _ := client.Users()
	groups, _ := client.Groups()
	serverInstance.Stats.Reset()

	username := uniqName()
	client.CreateUser(username, "secret")
	userCount := int64(len(users) + 1)
	groupCount := int64(len(groups))
	defer client.DeleteUser(username)
	client.LoginUser(username, "secret") // Successfull login
	client.LoginUser("foobar", "123456") // Failed login
	client.LoginUser("foobar", "123456") // twice
	tempClient := newClient()            // Connects++
	// since the server and client are not in sync sleep
	time.Sleep(50 * time.Millisecond)

	stats, _ := client.Stats()
	expected := map[string]int64{
		"Connects":                             1,
		"Disconnects":                          0,
		"Active Connections":                   1,
		"Successfull logins":                   1,
		"Failed logins":                        2,
		"Unrestricted Commands":                0,
		"Restricted Commands":                  4,
		"Access denied on Restricted Commands": 0,
		"Users":  userCount,
		"Groups": groupCount,
	}
	if !reflect.DeepEqual(stats, expected) {
		t.Fatalf("expected %v to be %v", stats, expected)
	}

	tempClient.Close() // Disconnects++
	// since the server and client are not in sync sleep
	time.Sleep(50 * time.Millisecond)

	stats, _ = client.Stats()
	expected = map[string]int64{
		"Connects":                             1,
		"Disconnects":                          1,
		"Active Connections":                   0,
		"Successfull logins":                   1,
		"Failed logins":                        2,
		"Unrestricted Commands":                1,
		"Restricted Commands":                  5,
		"Access denied on Restricted Commands": 0,
		"Users":  userCount,
		"Groups": groupCount,
	}
	if !reflect.DeepEqual(stats, expected) {
		t.Fatalf("expected %v to be %v", stats, expected)
	}
}

func BenchmarkCreateUser(b *testing.B) {
	client := newClient()
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := uniqName()
		b.StartTimer()
		client.CreateUser(username, "secret")
		b.StopTimer()
		defer client.DeleteUser(username)
	}
}

func BenchmarkDeleteUser(b *testing.B) {
	client := newClient()
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := uniqName()
		client.CreateUser(username, "secret")
		b.StartTimer()
		client.DeleteUser(username)
		b.StopTimer()
	}
}
