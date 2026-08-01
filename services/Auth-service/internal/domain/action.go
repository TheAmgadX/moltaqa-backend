package domain

type AuthAction string

const (
	ActionLogin   AuthAction = "Login"
	ActionRestore AuthAction = "Restore"
	ActionDelete  AuthAction = "Delete"
)
