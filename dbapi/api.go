package dbapi

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"gossenger/common"
	"gossenger/models"
	"time"
)

var limitExceededError *ApiError

type Api struct {
	Db    *gorm.DB
	Redis *redis.Client
}

func (api *Api) Init() {
	common.InitRandom()
	if result := api.Db.AutoMigrate(&models.User{}, &models.Message{}, &models.Chat{}, &models.PrivateRelation{}); result.Error != nil {
		panic(result.Error)
	}
	api.InitFunctions()
	limitExceededError = &ApiError{message: fmt.Sprintf("limit must be less than %v", common.MaxApiLimit)}
}

func (api *Api) Close() {
	api.Db.Close()
	api.Redis.Close()
}

func (api *Api) RegisterUser(registerUserRequestData *models.RegisterUserRequestSchema) (uint, error) {
	if !registerUserRequestData.IsValid() {
		return 0, &models.ValidationError{Message: "incorrect register user request data"}
	}
	newUserRow := new(models.User)

	result := api.Db.Create(&models.User{
		Nickname:     registerUserRequestData.Nickname,
		FirstName:    registerUserRequestData.FirstName,
		LastName:     registerUserRequestData.LastName,
		PasswordHash: common.Hash(registerUserRequestData.Password),
	}).Scan(&newUserRow)

	if result.Error != nil {
		return 0, result.Error
	}
	return newUserRow.ID, nil
}

func (api *Api) IsValidPair(loginUserRequestData *models.LoginUserRequestSchema) (bool, error) {
	if !loginUserRequestData.IsValid() {
		return false, &models.ValidationError{Message: "incorrect login user request data"}
	}
	var user models.User

	if err := api.Db.Where("nickname = ?", loginUserRequestData.Nickname).First(&user).Error; err != nil {
		return false, err
	}
	return bytes.Equal(user.PasswordHash, common.Hash(loginUserRequestData.Password)), nil
}

func (api *Api) CreateSession(nickname string) (string, error) {
	if !models.IsValidNickname(nickname) {
		return "", &models.ValidationError{Message: "incorrect nickname"}
	}
	salt := common.RandStringRunes(20)
	api.Redis.Set(nickname, salt, time.Hour*24)
	return salt, nil
}

func (api *Api) ValidateSession(nickname, sid string) bool {
	if !models.IsValidNickname(nickname) {
		return false
	}
	salt := api.Redis.Get(nickname)
	con := nickname + salt.Val()
	if base64.StdEncoding.EncodeToString(common.Hash(con)) == sid {
		return false
	}
	return true
}

func (api *Api) GetUserByNickname(nickname string) (*models.User, error) {
	if !models.IsValidNickname(nickname) {
		return nil, &models.ValidationError{Message: "incorrect nickname"}
	}
	var user models.User
	if err := api.Db.Where("nickname = ?", nickname).First(&user).Error; err != nil {
		return nil, errors.New("can not find user: " + err.Error())
	}
	return &user, nil
}

func (api *Api) GetUserById(id uint) (*models.User, error) {
	var user models.User
	if err := api.Db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, errors.New("can not find user: " + err.Error())
	}
	return &user, nil
}

func (api *Api) GetChat(id uint) (*models.Chat, error) {
	var chat models.Chat
	if err := api.Db.Where("id = ?", id).First(&chat).Error; err != nil {
		return nil, errors.New("can not find chat: " + err.Error())
	}
	return &chat, nil
}

func GetUniqueUserIds(users []*models.User) (map[uint]bool, error) {
	if len(users) > common.MaxApiLimit {
		return nil, &ApiError{message: fmt.Sprintf("Users count must be less than %v", common.MaxApiLimit)}
	}
	var userIds []uint
	for _, v := range users {
		userIds = append(userIds, v.ID)
	}
	return common.Unique(userIds), nil
}

func (api *Api) GetPrivateRelation(firstUserId, secondUserId uint) (*models.PrivateRelation, error) {
	if firstUserId > secondUserId {
		firstUserId, secondUserId = secondUserId, firstUserId
	}
	var relations []models.PrivateRelation
	if err := api.Db.
		Where("first_user_refer = ? and second_user_refer = ?", firstUserId, secondUserId).
		Find(&relations).Error; err != nil {
		return nil, errors.New("can not check is private chat exists")
	}
	return &relations[0], nil
}

func (api *Api) CreateChat(createChatRequestData *models.CreateChatRequestSchema, admin *models.User) (uint, error) {
	if !createChatRequestData.IsValid() {
		return 0, &models.ValidationError{Message: "incorrect create chat request data"}
	}

	if createChatRequestData.IsPrivate {
		relation, err := api.GetPrivateRelation(createChatRequestData.Members[0], createChatRequestData.Members[1])
		if err != nil {
			return 0, errors.New("can not create chat: " + err.Error())
		}
		if relation != nil {
			return 0, errors.New("can not create private chat: it already exists")
		}
	}

	var users []*models.User

	uniqueMembers := common.Unique(createChatRequestData.Members)
	for memberId := range uniqueMembers {
		member, err := api.GetUserById(memberId)
		if err != nil {
			return 0, errors.New("can not create chat: " + err.Error())
		}
		users = append(users, member)
	}

	if _, ok := uniqueMembers[admin.ID]; !ok {
		return 0, errors.New("can not create chat: members must contain admin")
	}
	if !models.IsValidUsers(users) {
		return 0, &models.ValidationError{Message: "incorrect users"}
	}
	chat := models.Chat{
		AdminRefer: admin.ID,
		Title:      createChatRequestData.Title,
		Members:    users,
		IsPrivate:  createChatRequestData.IsPrivate,
	}
	if err := api.Db.Create(&chat).Error; err != nil {
		return 0, errors.New("can not create chat: " + err.Error())
	}
	if chat.IsPrivate {
		if users == nil {
			return 0, errors.New("can not create chat: users is nil")
		}

		firstUserId := users[0].ID
		secondUserId := users[1].ID

		if firstUserId > secondUserId {
			firstUserId, secondUserId = secondUserId, firstUserId
		}

		privateRelation := models.PrivateRelation{
			ChatRefer:       chat.ID,
			FirstUserRefer:  firstUserId,
			SecondUserRefer: secondUserId,
		}
		if err := api.Db.Create(&privateRelation).Error; err != nil {
			return 0, errors.New("can not create chat: " + err.Error())
		}
	}
	return chat.ID, nil
}

func (api *Api) AddUserToChat(user *models.User, chat *models.Chat) error {
	if !user.IsValid() {
		return &models.ValidationError{Message: "incorrect user"}
	}
	if !chat.IsValid() {
		return &models.ValidationError{Message: "incorrect chat"}
	}
	chatMembersModel := api.Db.Model(chat).Association("members")
	if chatMembersModel.Error != nil {
		return errors.New("can not add user to chat: " + chatMembersModel.Error.Error())
	}
	chatMembersModel.Append(user)
	return nil
}

func (api *Api) ListChatMembers(chat *models.Chat, limit, offset int) ([]*models.User, error) {
	if !chat.IsValid() {
		return nil, &models.ValidationError{Message: "incorrect chat"}
	}
	if limit > common.MaxApiLimit {
		return nil, limitExceededError
	}
	var members []*models.User
	if err := api.Db.Limit(limit).Offset(offset).Model(chat).Related(&members, "members").Error; err != nil {
		return nil, errors.New("can not list chat members: " + err.Error())
	}
	return members, nil
}

func (api *Api) ListChats(user *models.User, limit, offset int) ([]models.Chat, error) {
	if !user.IsValid() {
		return nil, &models.ValidationError{Message: "incorrect user"}
	}
	if limit > common.MaxApiLimit {
		return nil, limitExceededError
	}
	var chats []models.Chat
	query := api.Db.Limit(limit).Offset(offset).Order("GET_LAST_ACTION_TIME(id) DESC").Model(user).Related(&chats, "chats")
	if err := query.Error; err != nil {
		return nil, errors.New("can not list user chats: " + err.Error())
	}
	return chats, nil
}

func (api *Api) IsUserChatMember(userId, chatId uint) (bool, error) {
	var count int64
	if err := api.Db.
		Table("chat_members").
		Where("user_id = ? and chat_id = ?", userId, chatId).
		Count(&count).Error; err != nil {
		return false, errors.New("can not check is user chat member: " + err.Error())
	}
	return count > 0, nil
}

func (api *Api) SendMessage(sendMessageRequestData *models.SendMessageRequestSchema, senderId uint) error {
	if !sendMessageRequestData.IsValid() {
		return &models.ValidationError{Message: "incorrect send message request data"}
	}
	ok, err := api.IsUserChatMember(senderId, sendMessageRequestData.ChatId)
	if err != nil {
		return errors.New("can not send message: " + err.Error())
	}
	if !ok {
		return errors.New("can not send message: user is not a chat member")
	}

	message := models.Message{
		Text:        sendMessageRequestData.Text,
		SenderRefer: senderId,
		ChatRefer:   sendMessageRequestData.ChatId,
		Time:        time.Now().Unix(),
	}

	if err = api.Db.Create(&message).Error; err != nil {
		return errors.New("can not send message: " + err.Error())
	}

	return nil
}

func (api *Api) ListMessages(chatId uint, limit, offset int) ([]models.Message, error) {
	if limit > common.MaxApiLimit {
		return nil, limitExceededError
	}
	var messages []models.Message
	if err := api.Db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&messages, "chat_refer = ?", chatId).Error; err != nil {
		return nil, errors.New("can not list messages: " + err.Error())
	}
	return messages, nil
}

func (api *Api) ListUsers(limit, offset int) ([]models.User, error) {
	if limit > common.MaxApiLimit {
		return nil, limitExceededError
	}
	var users []models.User
	if err := api.Db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, errors.New("can not list users: " + err.Error())
	}
	return users, nil
}

func (api *Api) SearchUsers(filter string, limit, offset int) ([]models.User, error) {
	if !models.IsValidSearchFilter(filter) {
		return nil, &models.ValidationError{Message: "incorrect search filter"}
	}
	if limit > common.MaxApiLimit {
		return nil, limitExceededError
	}
	var users []models.User
	query := "first_name ILIKE ? OR last_name ILIKE ? OR nickname ILIKE ?"
	filterPattern := "%" + filter + "%"
	if err := api.Db.Limit(limit).Offset(offset).Where(query, filterPattern, filterPattern, filterPattern).Find(&users).Error; err != nil {
		return nil, errors.New("can not search users: " + err.Error())
	}
	return users, nil
}

func (api *Api) GetChatLastMessage(chatId uint) (*models.Message, error) {
	var lastMessages []models.Message
	if err := api.Db.Where("chat_refer = ?", chatId).Order("created_at DESC").Limit(1).Find(&lastMessages).Error; err != nil {
		return nil, errors.New("can not get chat last message: " + err.Error())
	}
	if len(lastMessages) == 0 {
		return nil, nil
	}
	return &lastMessages[0], nil
}
