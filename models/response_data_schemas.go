package models

type ChatResponseSchema struct {
	ChatId               uint   `json:"chat_id"`
	Title                string `json:"title"`
	PreviewMessageSender uint   `json:"preview_message_sender"`
	PreviewMessageText   string `json:"preview_message_text"`
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

type FastChatCreatingResponseSchema struct {
	Title                string `json:"title"`
	ID                   uint   `json:"id"`
	PreviewMessageSender uint   `json:"preview_message_sender"`
	PreviewMessageText   string `json:"preview_message_text"`
}

type UserResponseSchema struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ListUsersResponseSchema []UserResponseSchema
