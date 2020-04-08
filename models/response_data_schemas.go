package models

type ChatResponseSchema struct {
	ChatId uint   `json:"chat_id"`
	Title  string `json:"title"`
}

type ListChatsResponseSchema []ChatResponseSchema

type MessageResponseSchema struct {
	Text     string `json:"text"`
	SenderId uint   `json:"sender_id"`
	Time     int64  `json:"time"`
}

type ListMessagesResponseSchema []MessageResponseSchema

type ChatMemberResponseSchema struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ListChatMembersResponseSchema []ChatMemberResponseSchema

type FastMessageResponseSchema struct {
	Text     string `json:"text"`
	SenderId uint   `json:"sender_id"`
	Time     int64  `json:"time"`
	ChatId   uint   `json:"chat_id"`
}
