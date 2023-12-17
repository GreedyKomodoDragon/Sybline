package rbac

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
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
	ALLOW_CHANGE_PASSWORD
	DENY_CHANGE_PASSWORD
	ALLOW_ADD_ROUTING_KEY
	DENY_ADD_ROUTING_KEY
	ALLOW_DELETE_ROUTING_KEY
	DENY_DELETE_ROUTING_KEY
)

const (
	ALL string = "*"
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
	CreateRole(jsonRole string) (*Role, error)
	AddRole(role Role) error
	DeleteRole(role string) error
	AssignRole(name, role string) error
	UnassignRole(name, role string) error
	HasAdminPermission(username string, permission AdminPermission) (bool, error)
	HasPermission(username string, entity string, permission Action) (bool, error)
	RoleExists(role string) bool
}

func NewRoleManager() RoleManager {
	// admin role
	manager := &roleManager{
		roles: map[string]Role{},
		users: map[string][]*Role{},
		lock:  &sync.RWMutex{},
	}

	createBuiltInRoles(manager)

	return manager
}

type roleManager struct {
	roles map[string]Role
	users map[string][]*Role
	lock  *sync.RWMutex
}

func (r *roleManager) CreateRole(jsonRole string) (*Role, error) {
	var roleData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonRole), &roleData); err != nil {
		return nil, err
	}

	roleName, ok := roleData["role"]
	if !ok {
		return nil, fmt.Errorf("missing role name")
	}

	nameStr, ok := roleName.(string)
	if !ok || len(nameStr) == 0 {
		return nil, fmt.Errorf("role name must have length >1")
	}

	r.lock.RLock()
	defer r.lock.RUnlock()
	if _, ok := r.roles[nameStr]; ok {
		return nil, fmt.Errorf("role '%s' already exists", nameStr)
	}

	role := &Role{
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
		return nil, fmt.Errorf("missing actions")
	}

	actions, ok := act.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid actions")
	}

	for action, permissions := range actions {
		perm, ok := permissions.(string)
		if !ok {
			return nil, fmt.Errorf("invalid permission")
		}

		switch action {
		case "GetMessages":
			mp, err := parsePermissions(perm)
			if err != nil {
				return nil, err
			}

			role.GetMessages = mp
		case "SubmitMessage":
			mp, err := parsePermissions(perm)
			if err != nil {
				return nil, err
			}

			role.SubmitMessage = mp
		case "SubmitBatchedMessages":
			mp, err := parsePermissions(perm)
			if err != nil {
				return nil, err
			}

			role.SubmitBatchedMessages = mp
		case "BatchAck":
			mp, err := parsePermissions(perm)
			if err != nil {
				return nil, err
			}

			role.BatchAck = mp

		case "CreateQueue":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CREATE_QUEUE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CREATE_QUEUE)
			}

		case "DeleteQueue":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_QUEUE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_QUEUE)
			}
		case "CreateUser":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CREATE_USER)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CREATE_USER)
			}
		case "DeleteUser":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_USER)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_USER)
			}
		case "CreateRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CREATE_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CREATE_ROLE)
			}
		case "DeleteRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_ROLE)
			}
		case "AssignRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_ASSIGN_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_ASSIGN_ROLE)
			}
		case "UnAssignRole":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_UNASSIGN_ROLE)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_UNASSIGN_ROLE)
			}
		case "ChangePassword":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_CHANGE_PASSWORD)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_CHANGE_PASSWORD)
			}
		case "AddRoutingKey":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_ADD_ROUTING_KEY)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_ADD_ROUTING_KEY)
			}
		case "DeleteRoutingKey":
			result, err := parseAdminPermissions(perm)
			if err != nil {
				return nil, err
			}

			if result {
				role.AdminPermissions = append(role.AdminPermissions, ALLOW_DELETE_ROUTING_KEY)
			} else {
				role.AdminPermissions = append(role.AdminPermissions, DENY_DELETE_ROUTING_KEY)
			}
		default:
			return nil, fmt.Errorf("invalid field found: %s", perm)
		}
	}

	return role, nil
}

func (r *roleManager) AddRole(role Role) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.roles[role.Name]; ok {
		return fmt.Errorf("role name already exists")
	}

	r.roles[role.Name] = role
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
	r.lock.Lock()
	defer r.lock.Unlock()

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
	r.lock.Lock()
	defer r.lock.Unlock()

	role, ok := r.roles[roleName]
	if !ok {
		return &RoleDoesNotExistError{
			Name: roleName,
		}
	}

	_, ok = r.users[username]
	if !ok {
		r.users[username] = []*Role{
			&role,
		}

		return nil
	}

	r.users[username] = append(r.users[username], &role)

	return nil
}

func (r *roleManager) UnassignRole(username, roleName string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.roles[roleName]; !ok {
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
			r.users[username] = remove(roles, i)
			return nil
		}
	}

	return fmt.Errorf("user with name '%s' does not have role '%s' to be unassigned", username, roleName)
}
func (r *roleManager) HasAdminPermission(username string, permission AdminPermission) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	roles, ok := r.users[username]
	if !ok {
		return false, fmt.Errorf("user with name '%s' does not exist or have any roles", username)
	}

	// Assumes passed in is allow therefore deny is always one on
	permissionDeny := permission + 1
	foundAllow := false

	for _, rol := range roles {
		for _, permissions := range rol.AdminPermissions {
			if permissions == permissionDeny {
				return false, nil
			}

			if permissions == permission {
				foundAllow = true
			}
		}
	}

	return foundAllow, nil
}

func (r *roleManager) HasPermission(username string, entity string, permission Action) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	roles, ok := r.users[username]
	if !ok {
		return false, fmt.Errorf("user with name '%s' does not exist or have any roles", username)
	}

	hasPerm := false

	for _, rol := range roles {
		switch permission {
		case GET_MESSAGES_ACTION:
			value, ok := rol.GetMessages[ALL]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

			value, ok = rol.GetMessages[entity]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

		case ACK_ACTION:
			value, ok := rol.Ack[ALL]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

			value, ok = rol.Ack[entity]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}
		case BATCH_ACK_ACTION:
			value, ok := rol.BatchAck[ALL]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

			value, ok = rol.BatchAck[entity]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

		case SUBMIT_MESSAGE_ACTION:
			value, ok := rol.SubmitMessage[ALL]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

			value, ok = rol.SubmitMessage[entity]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

		case SUBMIT_BATCH_ACTION:
			value, ok := rol.SubmitBatchedMessages[ALL]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}

			value, ok = rol.SubmitBatchedMessages[entity]
			if ok {
				if value {
					hasPerm = true
					continue
				}

				return false, nil
			}
		}

	}
	return hasPerm, nil
}

func (r *roleManager) RoleExists(role string) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	_, ok := r.roles[role]
	return ok
}

func remove(s []*Role, i int) []*Role {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
