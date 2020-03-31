package dbapi

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"messenger/models"
	"messenger/utils"
	"time"
)

type Api struct {
	Db    *gorm.DB
	Redis *redis.Client
}

func (api *Api) Init() {
	utils.InitRandom()
	if result := api.Db.AutoMigrate(&models.User{}, &models.Message{}, &models.Chat{}); result.Error != nil {
		panic(result.Error)
	}
}

func (api *Api) Close() {
	api.Db.Close()
	api.Redis.Close()
}

func (api *Api) RegisterUser(nickname, firstName, lastName, password string) error {
	result := api.Db.Create(&models.User{
		Nickname:     nickname,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: utils.Hash(password),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (api *Api) IsValidPair(nickname, password string) (bool, error) {
	var user models.User

	if err := api.Db.Where("nickname = ?", nickname).First(&user).Error; err != nil {
		return false, err
	}
	return bytes.Equal(user.PasswordHash, utils.Hash(password)), nil
}

func (api *Api) CreateSession(nickname string) string {
	salt := utils.RandStringRunes(20)
	api.Redis.Set(nickname, salt, time.Second*600)
	return salt
}

func (api *Api) ValidateSession(nickname, sid string) bool {
	salt := api.Redis.Get(nickname)
	con := nickname + salt.Val()
	if base64.StdEncoding.EncodeToString(utils.Hash(con)) == sid {
		return false
	}
	return true
}

func (api *Api) GetUserByNickname(nickname string) (*models.User, error) {
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

func GetUniqueUserIds(users []*models.User) map[uint]bool {
	var userIds []uint
	for _, v := range users {
		userIds = append(userIds, v.ID)
	}
	return utils.Unique(userIds)
}

func (api *Api) CreateChat(title string, admin *models.User, users []*models.User) error {
	uniqueUserIds := GetUniqueUserIds(users)

	if len(uniqueUserIds) < 2 {
		return errors.New("can not create chat: members must contain at least two unique users")
	}

	if _, ok := uniqueUserIds[admin.ID]; !ok {
		return errors.New("can not create chat: members must contain admin")
	}

	chat := models.Chat{
		AdminRefer: admin.ID,
		Title:      title,
		Members:    users,
	}
	if err := api.Db.Create(&chat).Error; err != nil {
		return errors.New("can not create chat: " + err.Error())
	}
	return nil
}

func (api *Api) AddUserToChat(user *models.User, chat *models.Chat) error {
	chatMembersModel := api.Db.Model(chat).Association("members")
	if chatMembersModel.Error != nil {
		return errors.New("can not add user to chat: " + chatMembersModel.Error.Error())
	}
	chatMembersModel.Append(user)
	return nil
}

func (api *Api) ListChatMembers(chat *models.Chat) ([]*models.User, error) {
	var members []*models.User
	if err := api.Db.Model(chat).Related(&members, "members").Error; err != nil {
		return nil, errors.New("can not list chat members: " + err.Error())
	}
	return members, nil
}

func (api *Api) ListUserChats(user *models.User) ([]*models.Chat, error) {
	var chats []*models.Chat
	if err := api.Db.Model(user).Related(&chats, "chats").Error; err != nil {
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

func (api *Api) SendMessage(messageText string, senderId, chatId uint) error {
	ok, err := api.IsUserChatMember(senderId, chatId)
	if err != nil {
		return errors.New("can not send message: " + err.Error())
	}
	if !ok {
		return errors.New("can not send message: user is not a chat member")
	}

	message := models.Message{
		Text:        messageText,
		SenderRefer: senderId,
		ChatRefer:   chatId,
		Time:        time.Now().Unix(),
	}

	if err = api.Db.Create(&message).Error; err != nil {
		return errors.New("can not send message: " + err.Error())
	}

	return nil
}

func (api *Api) ListMessages(chatId uint) ([]models.Message, error) {
	var messages []models.Message
	if err := api.Db.Find(&messages, "chat_refer = ?", chatId).Error; err != nil {
		return nil, errors.New("can not list messages: " + err.Error())
	}
	return messages, nil
}

func (api *Api) ListUsers() (*[]models.User, error) {
	var users []models.User
	if err := api.Db.Find(&users).Error; err != nil {
		return nil, errors.New("can not list users: " + err.Error())
	}
	return &users, nil
}
