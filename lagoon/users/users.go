package users

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/amazeeio/lagoon-cli/api"
	"github.com/amazeeio/lagoon-cli/graphql"
	"github.com/amazeeio/lagoon-cli/output"
)

// AddUser function
func AddUser(user api.User) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `mutation addUser ($firstname: String, $lastname: String, $email: String!) {
				addUser(input:{firstName: $firstname, lastName: $lastname, email: $email}) {
					id
				}
			}`,
		Variables: map[string]interface{}{
			"firstname": user.FirstName,
			"lastname":  user.LastName,
			"email":     user.Email,
		},
		MappedResult: "addUser",
	}
	returnResult, err := lagoonAPI.Request(customReq)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

// AddSSHKeyToUser function
func AddSSHKeyToUser(user api.User, sshKey api.SSHKey) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `mutation addSshKey ($email: String!, $keyname: String!, $keyvalue: String!, $keytype: SshKeyType!) {
				addSshKey(input:{
					user: {
						email: $email
					}
					name: $keyname
					keyValue: $keyvalue
					keyType: $keytype
				}) {
					id
				}
			}`,
		Variables: map[string]interface{}{
			"email":    user.Email,
			"keyname":  sshKey.Name,
			"keyvalue": sshKey.KeyValue,
			"keytype":  sshKey.KeyType,
		},
		MappedResult: "addSshKey",
	}
	returnResult, err := lagoonAPI.Request(customReq)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

// DeleteSSHKey function
func DeleteSSHKey(keyName string) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `mutation deleteSshKey ($keyname: String!) {
				deleteSshKey(input:{name: $keyname})
		  	}`,
		Variables: map[string]interface{}{
			"keyname": keyName,
		},
		MappedResult: "deleteSshKey",
	}
	returnResult, err := lagoonAPI.Request(customReq)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

// DeleteUser function
func DeleteUser(user api.User) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `mutation deleteUser ($email: String!) {
				deleteUser(input:{user: {email: $email}})
			}`,
		Variables: map[string]interface{}{
			"email": user.Email,
		},
		MappedResult: "deleteUser",
	}
	returnResult, err := lagoonAPI.Request(customReq)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

// ModifyUser function
func ModifyUser(user api.User, patch api.User) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `mutation updateUser ($email: String!, $patch: UpdateUserPatchInput!) {
				updateUser(input:{
					user:{email: $email}
						patch: $patch
				}){
					id
				}
			}`,
		Variables: map[string]interface{}{
			"patch": patch,
			"email": user.Email,
		},
		MappedResult: "updateUser",
	}
	returnResult, err := lagoonAPI.Request(customReq)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

// ListUsers function
func ListUsers(groupName string) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `query allGroups ($name: String) {
				allGroups (name: $name) {
			  		name
			  		id
			  		members{
						user{
							id
							email
							firstName
							lastName
						}
						role
			  		}
				}
		  	}`,
		Variables: map[string]interface{}{
			"name": groupName,
		},
		MappedResult: "allGroups",
	}
	listUsers, err := lagoonAPI.Request(customReq)
	returnResult, err := processUserList(listUsers)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

func processUserList(listUsers []byte) ([]byte, error) {
	var groupMembers GroupMembers
	err := json.Unmarshal([]byte(listUsers), &groupMembers)
	if err != nil {
		return []byte(""), errors.New("no data returned from lagoon") // @TODO could be a permissions thing when no data is returned
	}
	// process the data for output
	data := []output.Data{}
	userDataStep1 := UsersData{}
	userDataStep2 := UsersData{}

	// initial sort to change group members to members with groups
	for _, group := range groupMembers {
		for _, member := range group.Members {
			userDataStep1.AddItem(UserData{ID: member.User.ID, Email: member.User.Email, FirstName: member.User.FirstName, LastName: member.User.LastName})
		}
	}
	// second sort to append the groups to the user data
	for _, usersData := range userDataStep1.User {
		userGroups := []UserGroup{}
		for _, group := range groupMembers {
			for _, member := range group.Members {
				if member.User.Email == usersData.Email {
					userGroups = append(userGroups, UserGroup{Name: group.Name, Role: member.Role})
				}
			}
		}
		usersData.Groups = userGroups
		userDataStep2.User = append(userDataStep2.User, usersData)
	}
	// finally display the re-sorted users with group information
	for _, i := range distinctObjects(userDataStep2.User) {
		for _, group := range i.Groups {
			userID := returnNonEmptyString(i.ID)
			userEmail := returnNonEmptyString(strings.Replace(i.Email, " ", "_", -1)) //remove spaces to make friendly for parsing with awk
			userFirstName := returnNonEmptyString(strings.Replace(i.FirstName, " ", "_", -1))
			userLastName := returnNonEmptyString(strings.Replace(i.LastName, " ", "_", -1))
			userGroup := returnNonEmptyString(strings.Replace(group.Name, " ", "_", -1))
			userRole := returnNonEmptyString(strings.Replace(group.Role, " ", "_", -1))
			data = append(data, []string{
				userID,
				userEmail,
				userFirstName,
				userLastName,
				userGroup,
				userRole,
			})
		}
	}
	dataMain := output.Table{
		Header: []string{"ID", "Name", "FirstName", "LastName", "Group", "Role"},
		Data:   data,
	}
	return json.Marshal(dataMain)
}

// ListUserSSHKeys function
func ListUserSSHKeys(email string, allUsers bool) ([]byte, error) {
	// set up a lagoonapi client
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	}
	customReq := api.CustomRequest{
		Query: `query allGroups {
				allGroups{
			  		name
			  		id
			  		members{
						user{
							id
							email
							firstName
							lastName
							sshKeys{
								name
								keyType
								keyValue
							}
						}
						role
			  		}
				}
		  	}`,
		Variables:    map[string]interface{}{},
		MappedResult: "allGroups",
	}
	listUsers, err := lagoonAPI.Request(customReq)
	returnedKeys, err := processReturnedUserKeysList(listUsers)
	var returnResult []byte
	if allUsers {
		returnResult, err = processAllUserKeysList(returnedKeys)
		if err != nil {
			return []byte(""), err
		}
	} else {
		returnResult, err = processUserKeysList(returnedKeys, email)
		if err != nil {
			return []byte(""), err
		}
	}
	return returnResult, nil
}

func processReturnedUserKeysList(listUsers []byte) ([]ExtendedSSHKey, error) {
	var groupMembers GroupMembers
	userDataStep1 := []ExtendedSSHKey{}
	err := json.Unmarshal([]byte(listUsers), &groupMembers)
	if err != nil {
		return userDataStep1, errors.New("no data returned from lagoon") // @TODO could be a permissions thing when no data is returned
	}
	// initial sort to change group members to members with groups
	for _, group := range groupMembers {
		for _, member := range group.Members {
			for _, key := range member.User.SSHKeys {
				userDataStep1 = append(userDataStep1, ExtendedSSHKey{SSHKey: &key, Email: member.User.Email})
			}
		}
	}
	return userDataStep1, nil
}

func processAllUserKeysList(listUsers []ExtendedSSHKey) ([]byte, error) {
	// second sort to append the groups to the user data
	data := []output.Data{}
	for _, usersData := range distinctKeys(listUsers) {
		userEmail := returnNonEmptyString(strings.Replace(usersData.Email, " ", "_", -1)) //remove spaces to make friendly for parsing with awk
		keyName := returnNonEmptyString(strings.Replace(usersData.SSHKey.Name, " ", "_", -1))
		keyValue := returnNonEmptyString(strings.Replace(usersData.SSHKey.KeyValue, " ", "_", -1))
		keyType := returnNonEmptyString(strings.Replace(string(usersData.SSHKey.KeyType), " ", "_", -1))
		data = append(data, []string{
			userEmail,
			keyName,
			keyType,
			keyValue,
		})
	}
	dataMain := output.Table{
		Header: []string{"Email", "Name", "Type", "Value"},
		Data:   data,
	}
	return json.Marshal(dataMain)
}

func processUserKeysList(listUsers []ExtendedSSHKey, email string) ([]byte, error) {
	// second sort to append the groups to the user data
	data := []output.Data{}
	for _, usersData := range distinctKeys(listUsers) {
		if usersData.Email == email {
			userEmail := returnNonEmptyString(strings.Replace(usersData.Email, " ", "_", -1)) //remove spaces to make friendly for parsing with awk
			keyName := returnNonEmptyString(strings.Replace(usersData.SSHKey.Name, " ", "_", -1))
			keyValue := returnNonEmptyString(strings.Replace(usersData.SSHKey.KeyValue, " ", "_", -1))
			keyType := returnNonEmptyString(strings.Replace(string(usersData.SSHKey.KeyType), " ", "_", -1))
			data = append(data, []string{
				userEmail,
				keyName,
				keyType,
				keyValue,
			})
		}
	}
	dataMain := output.Table{
		Header: []string{"Email", "Name", "Type", "Value"},
		Data:   data,
	}
	return json.Marshal(dataMain)
}

func returnNonEmptyString(value string) string {
	if len(value) == 0 {
		value = "-"
	}
	return value
}

type ExtendedSSHKey struct {
	*api.SSHKey
	Email string `json:"email"`
}
type UserGroup struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Role string `json:"role"`
}

type UsersData struct {
	User []UserData
}

type UserData struct {
	ID        string       `json:"id"`
	Email     string       `json:"email"`
	FirstName string       `json:"firstName"`
	LastName  string       `json:"lastName"`
	SSHKeys   []api.SSHKey `json:"sshKeys"`
	Groups    []UserGroup  `json:"groups"`
}

type GroupMembers []struct {
	ID      string   `json:"id"`
	Members []Member `json:"members"`
	Name    string   `json:"name"`
}

type Member struct {
	Role string `json:"role"`
	User struct {
		ID        string       `json:"id"`
		Email     string       `json:"email"`
		FirstName string       `json:"firstName"`
		SSHKeys   []api.SSHKey `json:"sshKeys"`
		LastName  string       `json:"lastName"`
	} `json:"user"`
}

func (ud *UsersData) AddItem(userData UserData) {
	ud.User = append(ud.User, userData)
}

func distinctObjects(objs []UserData) (distinctedObjs []UserData) {
	var output []UserData
	for i := range objs {
		if output == nil || len(output) == 0 {
			output = append(output, objs[i])
		} else {
			founded := false
			for j := range output {
				if output[j].ID == objs[i].ID {
					founded = true
				}
			}
			if !founded {
				output = append(output, objs[i])
			}
		}
	}
	return output
}

func distinctKeys(objs []ExtendedSSHKey) (distinctedObjs []ExtendedSSHKey) {
	var output []ExtendedSSHKey
	for i := range objs {
		if output == nil || len(output) == 0 {
			output = append(output, objs[i])
		} else {
			founded := false
			for j := range output {
				if output[j].Email == objs[i].Email {
					founded = true
				}
			}
			if !founded {
				output = append(output, objs[i])
			}
		}
	}
	return output
}
