package models

type RegisterUserSchema struct {
	Password  string `requestsJson:"password"`
	Nickname  string `requestsJson:"nickname"`
	FirstName string `requestsJson:"first_name"`
	LastName  string `requestsJson:"last_name"`
}

type LoginUserSchema struct {
	Password string `requestsJson:"password"`
	Nickname string `requestsJson:"nickname"`
}

type CreateChatSchema struct {
	Title   string `requestsJson:"title"`
	AdminId uint   `requestsJson:"admin_id"`
	Members []uint `requestsJson:"members"`
}

type AddUserToChatSchema struct {
	UserId uint `requestsJson:"user_id"`
	ChatId uint `requestsJson:"chat_id"`
}
