package api

import (
	"gitlab.cern.ch/lb-experts/goermis/auth"
)

//currentUser creates a lightweight user profile
var currentUser User

//User describes the profile of a user
type User struct {
	Username  string
	Superuser bool
	Pwn       []string
}

//SetUser creates a mini profile
func SetUser(username string) {
	currentUser.Username = username

	if currentUser.Username != "" {
		var d auth.Group
		currentUser.Superuser = d.CheckCud(currentUser.Username)
		currentUser.Pwn = auth.GetPwn(currentUser.Username)
	} else {
		currentUser.Superuser = false
		currentUser.Pwn = []string{}

	}

}

//GetUserProfile returns the profile of current user
func GetUserProfile() User {
	return currentUser
}

//GetUsersHostgroups returns current users hostgroups
func GetUsersHostgroups() []string {
	return currentUser.Pwn
}

//GetUsername returns current users hostgroups
func GetUsername() string {
	return currentUser.Username
}

//IsSuperuser returns current users hostgroups
func IsSuperuser() bool {
	return currentUser.Superuser
}
