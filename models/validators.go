package models

import (
	"fmt"
	"regexp"
)

var PasswordPattern = regexp.MustCompile(fmt.Sprintf("^[[:print:]]{%v,%v}$", MinPasswordLength, MaxPasswordLength))
var NicknamePattern = regexp.MustCompile(fmt.Sprintf("[[:alnum:]]{%v,%v}$", MinNicknameLength, MaxNicknameLength))
var FirstNamePattern = regexp.MustCompile(fmt.Sprintf("[[:alpha:]]{%v,%v}$", MinFirstNameLength, MaxFirstNameLength))
var LastNamePattern = regexp.MustCompile(fmt.Sprintf("[[:alpha:]]{%v,%v}$", MinLastNameLength, MaxLastNameLength))

var ChatTitlePattern = regexp.MustCompile(fmt.Sprintf("[[:print:]]{%v,%v}", MinChatTitleLength, MaxChatTitleLength))

var SearchFilterPattern = regexp.MustCompile(fmt.Sprintf("[[:alnum:]]{%v,%v}", MinSearchFilterLength, MaxSearchFilterLength))

type ValidationError struct {
	Message string
}

func (validationError *ValidationError) Error() string {
	return "Validation error: " + validationError.Message
}

func IsValidNickname(nickname string) bool {
	return NicknamePattern.MatchString(nickname)
}

func IsValidPassword(password string) bool {
	return PasswordPattern.MatchString(password)
}

func IsValidFirstName(firstName string) bool {
	return FirstNamePattern.MatchString(firstName)
}

func IsValidLastName(lastName string) bool {
	return LastNamePattern.MatchString(lastName)
}

func IsValidSearchFilter(filter string) bool {
	return SearchFilterPattern.MatchString(filter)
}

func (registerUserRequestData *RegisterUserRequestSchema) IsValid() bool {
	return IsValidPassword(registerUserRequestData.Password) &&
		IsValidNickname(registerUserRequestData.Nickname) &&
		IsValidFirstName(registerUserRequestData.FirstName) &&
		IsValidLastName(registerUserRequestData.LastName)
}

func (loginUserRequestData *LoginUserRequestSchema) IsValid() bool {
	return IsValidPassword(loginUserRequestData.Password) &&
		IsValidNickname(loginUserRequestData.Nickname)
}

func (createChatRequestData *CreateChatRequestSchema) IsValid() bool {
	if createChatRequestData.IsPrivate {
		return len(createChatRequestData.Members) == 2
	}
	return IsValidChatTitle(createChatRequestData.Title) &&
		len(createChatRequestData.Members) >= MinChatMembersCount &&
		len(createChatRequestData.Members) <= MaxChatMembersCount
}

func (sendMessageRequestData *SendMessageRequestSchema) IsValid() bool {
	return len(sendMessageRequestData.Text) >= MinMessageLength &&
		len(sendMessageRequestData.Text) <= MaxMessageLength
}

func IsValidChatTitle(title string) bool {
	return ChatTitlePattern.MatchString(title)
}

func (user *User) IsValid() bool {
	return IsValidNickname(user.Nickname) &&
		IsValidFirstName(user.FirstName) &&
		IsValidLastName(user.LastName)
}

func IsValidUsers(users []*User) bool {
	for _, user := range users {
		if !user.IsValid() {
			return false
		}
	}
	return true
}

func (chat *Chat) IsValid() bool {
	return IsValidChatTitle(chat.Title) &&
		IsValidUsers(chat.Members)
}
