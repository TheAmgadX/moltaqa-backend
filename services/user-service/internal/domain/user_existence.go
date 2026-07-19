package domain

// UserExistence is used as a return value for UsersExist function to indicate whether a user exists or not.
type UserExistence struct {
	Id     string
	Exists bool
}

func NewUserExistence(id string, exists bool) UserExistence {
	return UserExistence{Id: id, Exists: exists}
}
