package rbac

func createBuiltInRoles(manager RoleManager) {
	allAdmin := []AdminPermission{
		ALLOW_CREATE_QUEUE,
		ALLOW_DELETE_QUEUE,
		ALLOW_CREATE_USER,
		ALLOW_DELETE_USER,
		ALLOW_CREATE_ROLE,
		ALLOW_DELETE_ROLE,
		ALLOW_ASSIGN_ROLE,
		ALLOW_UNASSIGN_ROLE,
		ALLOW_ADD_ROUTING_KEY,
		ALLOW_DELETE_ROUTING_KEY,
		ALLOW_CHANGE_PASSWORD,
	}

	anyAllow := map[string]bool{
		"*": true,
	}

	empty := map[string]bool{}

	// ROOT
	manager.AddRole(Role{
		Name:                  "ROOT",
		GetMessages:           anyAllow,
		SubmitMessage:         anyAllow,
		SubmitBatchedMessages: anyAllow,
		Ack:                   anyAllow,
		BatchAck:              anyAllow,
		AdminPermissions:      allAdmin,
	})

	// Admin
	manager.AddRole(Role{
		Name:                  "ADMIN",
		GetMessages:           empty,
		SubmitMessage:         empty,
		SubmitBatchedMessages: empty,
		Ack:                   empty,
		BatchAck:              empty,
		AdminPermissions:      allAdmin,
	})

	// Producer
	manager.AddRole(Role{
		Name:                  "PRODUCER",
		GetMessages:           empty,
		SubmitMessage:         anyAllow,
		SubmitBatchedMessages: anyAllow,
		Ack:                   empty,
		BatchAck:              empty,
		AdminPermissions:      []AdminPermission{},
	})

	// Consumer
	manager.AddRole(Role{
		Name:                  "CONSUMER",
		GetMessages:           anyAllow,
		SubmitMessage:         empty,
		SubmitBatchedMessages: empty,
		Ack:                   anyAllow,
		BatchAck:              anyAllow,
		AdminPermissions:      []AdminPermission{},
	})

	// Deny-All
	denyAdmin := []AdminPermission{
		DENY_CREATE_QUEUE,
		DENY_DELETE_QUEUE,
		DENY_CREATE_USER,
		DENY_DELETE_USER,
		DENY_CREATE_ROLE,
		DENY_DELETE_ROLE,
		DENY_ASSIGN_ROLE,
		DENY_UNASSIGN_ROLE,
		DENY_ADD_ROUTING_KEY,
		DENY_DELETE_ROUTING_KEY,
		DENY_CHANGE_PASSWORD,
	}

	anyDeny := map[string]bool{
		"*": false,
	}

	manager.AddRole(Role{
		Name:                  "DENY",
		GetMessages:           anyDeny,
		SubmitMessage:         anyDeny,
		SubmitBatchedMessages: anyDeny,
		Ack:                   anyDeny,
		BatchAck:              anyDeny,
		AdminPermissions:      denyAdmin,
	})
}
