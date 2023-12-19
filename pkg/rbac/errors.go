package rbac

import "fmt"

type RoleDoesNotExistError struct {
	Name string
}

func (e *RoleDoesNotExistError) Error() string {
	return fmt.Sprintf("role '%s' does not exist", e.Name)
}
