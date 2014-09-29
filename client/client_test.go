package client

import (
	"fmt"
	"testing"
	"time"
)

func uniqUsername() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}

func TestConnect(t *testing.T) {
	client, err := Dial("localhost:35786")
	defer client.Close()
	if err != nil {
		t.Fatal("client was unable to connect to server", err)
	}
}

func TestConnectTls(t *testing.T) {
	client, err := Dial("localhost:35786")
	defer client.Close()
	if err != nil {
		t.Fatal("client was unable to connect to server", err)
	}
	aerr := client.StartTlsWithCert("../config/cert.pem")
	if aerr != nil {
		t.Fatal("unable to establish tls", aerr)
	}
	username := uniqUsername()
	defer client.DeleteUser(username)
	id, serr := client.CreateUser(username, "secret")
	if id <= 0 {
		t.Fatal("user not created, expected id bigger than 0 got", id, serr)
	}
}

func TestCreateUser(t *testing.T) {
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username := uniqUsername()
	defer client.DeleteUser(username)
	id, err := client.CreateUser(username, "secret")
	if id <= 0 {
		t.Fatal("user not created, expected id bigger than 0 got", id, err)
	}
}

func TestDisableUser(t *testing.T) {
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username := uniqUsername()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	err := client.DisableUser(username)
	if err != nil {
		t.Fatal("unable to diable user", err)
	}
}

func TestEnableUser(t *testing.T) {
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username := uniqUsername()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	err := client.EnableUser(username)
	if err != nil {
		t.Fatal("unable to diable user", err)
	}
}

func TestSetGetUserData(t *testing.T) {
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username := uniqUsername()
	defer client.DeleteUser(username)
	client.CreateUser(username, "secret")
	client.SetUserData(username, "firstname", "Tester")
	firstname, err := client.GetUserData(username, "firstname")
	if firstname != "Tester" {
		t.Fatal("firstname should have been 'Tester' but was", firstname, err)
	}
}

func TestLoginUser(t *testing.T) {
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username := uniqUsername()
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

func TestDeleteUser(t *testing.T) {
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username := uniqUsername()
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
	client, _ := Dial("localhost:35786")
	defer client.Close()
	username0 := uniqUsername()
	uid0, _ := client.CreateUser(username0, "secret")
	username1 := uniqUsername()
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
