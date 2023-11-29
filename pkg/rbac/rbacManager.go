package rbac

import (
	"encoding/json"
	"fmt"
	"strings"
)

type AdminPermission uint32

const (
	ALLOW_CREATE_QUEUE = iota
	DENY_CREATE_QUEUE
	ALLOW_DELETE_QUEUE
	DENY_DELETE_QUEUE
	ALLOW_CREATE_USER
	DENY_CREATE_USER
	ALLOW_DELETE_USER
	DENY_DELETE_USER
	ALLOW_CREATE_ROLE
	DENY_CREATE_ROLE
	ALLOW_DELETE_ROLE
	DENY_DELETE_ROLE
	ALLOW_ASSIGN_ROLE
	DENY_ASSIGN_ROLE
	ALLOW_UNASSIGN_ROLE
	DENY_UNASSIGN_ROLE
)

type Action uint32

const (
	GET_MESSAGES_ACTION = iota
	SUBMIT_MESSAGE_ACTION
	SUBMIT_BATCH_ACTION
	ACK_ACTION
	BATCH_ACK_ACTION
)

type Role struct {
	Name                  string
	GetMessages           map[string]bool
	SubmitMessage         map[string]bool
	SubmitBatchedMessages map[string]bool
	Ack                   map[string]bool
	BatchAck              map[string]bool
	AdminPermissions      []AdminPermission
}

type RoleManager interface {
	CreateRole(jsonRole string) error
	DeleteRole(role string) error
	AssignRole(name, role string) error
	UnassignRole(name, role string) error
	HasAdminPermission(username string, permission AdminPermission) (bool, error)
	HasPermission(username string, entity string, permission Action) (bool, error)
}

func NewRoleManager() RoleManager {
	return &roleManager{
		roles: map[string]Role{},
		users: map[string][]*Role{},
	}
}

type roleManager struct {
	roles map[string]Role
	users map[string][]*Role
}

func (r *roleManager) CreateRole(jsonRole string) error {
	var roleData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonRole), &roleData); err != nil {
		return err
	}

	roleName, ok := roleData["role"]
	if !ok {
		return fmt.Errorf("missing role name")
	}

	nameStr, ok := roleName.(string)
	if !ok || len(nameStr) == 0 {
		return fmt.Errorf("role name must have length >1")
	}

	if _, ok := r.roles[nameStr]; ok {
		return fmt.Errorf("role '%s' already exists", nameStr)
	}

	role := Role{
		Name:                  nameStr,
		GetMessages:           make(map[string]bool),
		SubmitMessage:         make(map[string]bool),
		SubmitBatchedMessages: make(map[string]bool),
		Ack:                   make(map[string]bool),
		BatchAck:              make(map[string]bool),
		AdminPermissions:      []AdminPermission{},
	}

	act, ok := roleData["actions"]
	if !ok {
		return fmt.Errorf("missing actions")
	}

	actions, ok := act.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid actions")
	}

	for action, permissions := range actions {
		perm, ok := permissions.(string)
		if !ok {
			return fmt.Errorf("invalid permission")
		}

		switch action {
		case "GetMessages":
			mp, err := parsePermissions(perm)
			if err != nil {
				return err
			}

			role.GetMessages = mp
		case "SubmitMessage":
			mp, err := parsePermissions(perm)
			if err != nil {
				return err
			}

			role.SubmitMessage = mp
		case "SubmitBatchedMessages":
			mp, err := parsePermissions(perm)
			if err != nil {
				return err
			}

			role.SubmitBatchedMessages = mp
		case "BatchAck":
			mp, err := parsePermissions(perm)
			if err != nil {
				return err
			}

			role.BatchAck = mp

		case "CreateQueue":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CREATE_QUEUE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CREATE_QUEUE)
			}

		case "DeleteQueue":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_QUEUE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_QUEUE)
			}
		case "CreateUser":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CREATE_USER)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CREATE_USER)
			}
		case "DeleteUser":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_USER)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_USER)
			}
		case "CreateRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CREATE_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CREATE_ROLE)
			}
		case "DeleteRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_ROLE)
			}
		case "AssignRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_ASSIGN_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_ASSIGN_ROLE)
			}
		case "UnAssignRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_UNASSIGN_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_UNASSIGN_ROLE)
			}
		}
	}

	r.roles[nameStr] = role

	return nil
}

func parseAdminPermissions(permString string) (bool, error) {
	if permString == "allow" {
		return true, nil

	}

	if permString == "deny" {
		return false, nil
	}

	return false, fmt.Errorf("action can only be: deny or allow")
}

func parsePermissions(permString string) (map[string]bool, error) {
	perms := strings.Split(permString, ",")

	permissionMap := map[string]bool{}

	for _, perm := range perms {
		parts := strings.Split(perm, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid number of parts")
		}

		if len(parts[1]) == 0 {
			return nil, fmt.Errorf("invalid entity length")
		}

		if parts[0] == "allow" {
			permissionMap[parts[1]] = true
			continue
		} else if parts[0] != "deny" {
			return nil, fmt.Errorf("action can only be: deny or allow")
		}

		permissionMap[parts[1]] = false
	}

	return permissionMap, nil
}

func (r *roleManager) DeleteRole(role string) error {
	if _, ok := r.roles[role]; !ok {
		return &RoleDoesNotExistError{
			Name: role,
		}
	}

	delete(r.roles, role)

	for k, v := range r.users {
		for i, rol := range v {
			if rol.Name == role {
				r.users[k] = remove(v, i)
				break
			}
		}
	}

	return nil
}

func (r *roleManager) AssignRole(username, roleName string) error {
	role, ok := r.roles[roleName]
	if !ok {
		return &RoleDoesNotExistError{
			Name: roleName,
		}
	}

	roles, ok := r.users[username]
	if !ok {
		r.users[username] = []*Role{
			&role,
		}

		return nil
	}

	roles = append(roles, &role)

	return nil
}

func (r *roleManager) UnassignRole(username, roleName string) error {
	_, ok := r.roles[roleName]
	if !ok {
		return &RoleDoesNotExistError{
			Name: roleName,
		}
	}

	roles, ok := r.users[username]
	if !ok {
		return fmt.Errorf("user with name '%s' does not exist or have any roles", username)
	}

	for i, rol := range roles {
		if rol.Name == roleName {
			roles = remove(roles, i)
			return nil
		}
	}

	return fmt.Errorf("user with name '%s' does not have role '%s' to be unassigned", username, roleName)
}
func (r *roleManager) HasAdminPermission(username string, permission AdminPermission) (bool, error) {
	roles, ok := r.users[username]
	if !ok {
		return false, fmt.Errorf("user with name '%s' does not exist or have any roles", username)
	}

	// Assumes passed in is allow therefore deny is always one on
	permissionDeny := permission + 1
	foundAllow := false

	for _, rol := range roles {
		for _, permissions := range rol.AdminPermissions {
			if permissions == permission {
				foundAllow = true
				continue
			}

			if permission == permissionDeny {
				return false, nil
			}
		}
	}

	return foundAllow, nil
}
func (r *roleManager) HasPermission(username string, entity string, permission Action) (bool, error) {
	roles, ok := r.users[username]
	if !ok {
		return false, fmt.Errorf("user with name '%s' does not exist or have any roles", username)
	}

	for _, rol := range roles {
		switch permission {
		case GET_MESSAGES_ACTION:
			if value, ok := rol.GetMessages[entity]; ok && value {
				return true, nil
			}

		case ACK_ACTION:
			if value, ok := rol.Ack[entity]; ok && value {
				return true, nil
			}
		case BATCH_ACK_ACTION:
			if value, ok := rol.BatchAck[entity]; ok && value {
				return true, nil
			}

		case SUBMIT_MESSAGE_ACTION:
			if value, ok := rol.SubmitMessage[entity]; ok && value {
				return true, nil
			}

		case SUBMIT_BATCH_ACTION:
			if value, ok := rol.SubmitBatchedMessages[entity]; ok && value {
				return true, nil
			}
		}

	}
	return false, nil
}

func remove(s []*Role, i int) []*Role {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
