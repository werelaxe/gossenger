package models

type RegisterUserSchema struct {
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginUserSchema struct {
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type CreateChatSchema struct {
	Title   string `json:"title"`
	AdminId uint   `json:"admin_id"`
	Members []uint `json:"members"`
}
