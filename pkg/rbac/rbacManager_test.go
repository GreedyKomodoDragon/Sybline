package rbac_test

import (
	"sybline/pkg/rbac"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRole_Success(t *testing.T) {
	jsonRole := `{"role": "TestRole", "actions": {"GetMessages": "allow:entity1,deny:entity2"}}`
	rm := rbac.NewRoleManager()
	_, err := rm.CreateRole(jsonRole)

	require.NoError(t, err)
}

func TestCreateRole_Success_Admin_Permission(t *testing.T) {
	jsonRole := `{"role": "TestRole", "actions": {"GetMessages": "allow:entity1,deny:entity2", "CreateUser": "deny"}}`
	rm := rbac.NewRoleManager()
	_, err := rm.CreateRole(jsonRole)

	assert.NoError(t, err)
}

func TestCreateRole_InvalidInput_No_Name(t *testing.T) {
	invalidJSONRole := `{"invalid_field": "TestRole", "actions": {"GetMessages": "allow:entity1,deny:entity2"}}`
	rm := rbac.NewRoleManager()
	_, err := rm.CreateRole(invalidJSONRole)

	require.Error(t, err)
}

func Test_Builtin_Root_Can_Do_All(t *testing.T) {
	rm := rbac.NewRoleManager()

	ok, err := rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.Error(t, err)

	require.NoError(t, rm.AssignRole("sybline", "ROOT"))

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_BATCH_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.ACK_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.BATCH_ACK_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.GET_MESSAGES_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_QUEUE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_QUEUE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ASSIGN_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_UNASSIGN_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_USER)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_USER)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ADD_ROUTING_KEY)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROUTING_KEY)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CHANGE_PASSWORD)
	require.NoError(t, err)
	require.True(t, ok, "should be true")
}

func Test_Builtin_Admin(t *testing.T) {
	rm := rbac.NewRoleManager()

	ok, err := rm.HasPermission("sybline", "routeName", rbac.ALLOW_ASSIGN_ROLE)
	require.Error(t, err)

	require.NoError(t, rm.AssignRole("sybline", "ADMIN"))

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false: rbac.SUBMIT_MESSAGE_ACTION")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_BATCH_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.ACK_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.BATCH_ACK_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.GET_MESSAGES_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_QUEUE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_QUEUE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ASSIGN_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_UNASSIGN_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROLE)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_USER)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_USER)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ADD_ROUTING_KEY)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROUTING_KEY)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CHANGE_PASSWORD)
	require.NoError(t, err)
	require.True(t, ok, "should be true")
}

func Test_Builtin_Consumer(t *testing.T) {
	rm := rbac.NewRoleManager()

	ok, err := rm.HasPermission("sybline", "routeName", rbac.ACK_ACTION)
	require.Error(t, err)

	require.NoError(t, rm.AssignRole("sybline", "CONSUMER"))

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_BATCH_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.ACK_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.BATCH_ACK_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.GET_MESSAGES_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_QUEUE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_QUEUE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ASSIGN_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_UNASSIGN_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_USER)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_USER)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ADD_ROUTING_KEY)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROUTING_KEY)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CHANGE_PASSWORD)
	require.NoError(t, err)
	require.False(t, ok, "should be false")
}

func Test_Builtin_DENY_All(t *testing.T) {

	rm := rbac.NewRoleManager()

	ok, err := rm.HasPermission("sybline", "routeName", rbac.ACK_ACTION)
	require.False(t, ok, "should be false")
	require.Error(t, err)

	require.NoError(t, rm.AssignRole("sybline", "ROOT"))
	require.NoError(t, rm.AssignRole("sybline", "DENY"))

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_BATCH_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.ACK_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "queueName", rbac.BATCH_ACK_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasPermission("sybline", "routeName", rbac.GET_MESSAGES_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_QUEUE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_QUEUE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ASSIGN_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_UNASSIGN_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROLE)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CREATE_USER)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_USER)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_ADD_ROUTING_KEY)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_DELETE_ROUTING_KEY)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	ok, err = rm.HasAdminPermission("sybline", rbac.ALLOW_CHANGE_PASSWORD)
	require.NoError(t, err)
	require.False(t, ok, "should be false")
}

func Test_Builtin_Unassigned(t *testing.T) {
	rm := rbac.NewRoleManager()

	ok, err := rm.HasPermission("sybline", "routeName", rbac.ACK_ACTION)
	require.False(t, ok, "should be false")
	require.Error(t, err)

	require.NoError(t, rm.AssignRole("sybline", "ROOT"))
	require.NoError(t, rm.AssignRole("sybline", "DENY"))

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.NoError(t, err)
	require.False(t, ok, "should be false")

	require.NoError(t, rm.UnassignRole("sybline", "DENY"))

	ok, err = rm.HasPermission("sybline", "routeName", rbac.SUBMIT_MESSAGE_ACTION)
	require.NoError(t, err)
	require.True(t, ok, "should be true")
}
