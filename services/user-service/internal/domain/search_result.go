package domain

type UserSearchResult struct {
	Users   []UserSummary
	HasMore bool
}
