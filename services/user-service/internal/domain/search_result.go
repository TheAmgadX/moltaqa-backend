package domain

type UserSearchResult struct {
	Users        []UserSummary
	TotalResults uint32
}
