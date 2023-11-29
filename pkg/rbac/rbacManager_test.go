package rbac_test

import (
	"sybline/pkg/rbac"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRole_Success(t *testing.T) {
	// Your test logic for successful role creation goes here
	// For example:
	jsonRole := `{"role": "TestRole", "actions": {"GetMessages": "allow:entity1,deny:entity2"}}`
	rm := rbac.NewRoleManager()
	err := rm.CreateRole(jsonRole)

	require.NoError(t, err)
	// Additional assertions as needed
}

func TestCreateRole_Success_Admin_Permission(t *testing.T) {
	// Your test logic for successful role creation goes here
	// For example:
	jsonRole := `{"role": "TestRole", "actions": {"GetMessages": "allow:entity1,deny:entity2", "CreateUser": "deny"}}`
	rm := rbac.NewRoleManager()
	err := rm.CreateRole(jsonRole)

	assert.NoError(t, err)
	// Additional assertions as needed
}

func TestCreateRole_InvalidInput_No_Name(t *testing.T) {
	// Your test logic for invalid role creation input goes here
	// For example:
	invalidJSONRole := `{"invalid_field": "TestRole", "actions": {"GetMessages": "allow:entity1,deny:entity2"}}`
	rm := rbac.NewRoleManager()
	err := rm.CreateRole(invalidJSONRole)

	require.Error(t, err)
	// Additional assertions as needed
}
