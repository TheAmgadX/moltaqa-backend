package domain

type UserSummary struct {
	Id              string
	Username        string
	DisplayName     string
	PhoneNumber     string
	ProfileImageURL string
	AccountBadge    AccountBadgeType
}
