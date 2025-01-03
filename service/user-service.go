package service

import (
	"bpl/auth"
	"bpl/repository"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserService struct {
	UserRepository *repository.UserRepository
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		UserRepository: repository.NewUserRepository(db),
	}
}

func (s *UserService) GetUserByDiscordId(discordId int64) (*repository.User, error) {
	return s.UserRepository.GetUserByDiscordId(discordId)
}

func (s *UserService) GetUserByPoEAccount(poeAccount string) (*repository.User, error) {
	return s.UserRepository.GetUserByPoEAccount(poeAccount)
}

func (s *UserService) GetUserByTwitchId(twitchId string) (*repository.User, error) {
	return s.UserRepository.GetUserByTwitchId(twitchId)
}

func (s *UserService) SaveUser(user *repository.User) (*repository.User, error) {
	return s.UserRepository.SaveUser(user)
}

func (s *UserService) GetUsers() ([]*repository.User, error) {
	return s.UserRepository.GetUsers()
}

func (s *UserService) GetUserById(id int) (*repository.User, error) {
	return s.UserRepository.GetUserById(id)
}

func (s *UserService) GetUserFromAuthCookie(c *gin.Context) (*repository.User, error) {
	cookie, err := c.Cookie("auth")
	if err != nil {
		return nil, err
	}
	return s.GetUserFromToken(cookie)
}

func (s *UserService) GetUserFromToken(tokenString string) (*repository.User, error) {
	token, err := auth.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims := &auth.Claims{}
	if token.Valid {
		claims.FromJWTClaims(token.Claims)
		if err := claims.Valid(); err != nil {
			return nil, err
		}
		return s.GetUserById(claims.UserID)
	}
	return nil, jwt.ErrInvalidKey
}

func (s *UserService) ChangePermissions(userId int, permissions []repository.Permission) error {
	user, err := s.GetUserById(userId)
	if err != nil {
		return err
	}
	user.Permissions = permissions
	_, err = s.UserRepository.SaveUser(user)
	return err
}

func (s *UserService) RemoveProvider(user *repository.User, provider repository.OauthProvider) (*repository.User, error) {
	numberOfProviders := 0
	if user.DiscordID != nil {
		numberOfProviders++
	}
	if user.TwitchID != nil {
		numberOfProviders++
	}
	if user.POEAccount != nil {
		numberOfProviders++
	}
	if numberOfProviders < 2 {
		return nil, fmt.Errorf("cannot remove last provider")
	}

	switch provider {
	case repository.OauthProviderDiscord:
		user.DiscordID = nil
		user.DiscordName = nil
	case repository.OauthProviderTwitch:
		user.TwitchID = nil
		user.TwitchName = nil
	case repository.OauthProviderPoE:
		user.POEAccount = nil
		user.PoeToken = nil
		user.PoeTokenExpiresAt = nil
	default:
		return nil, fmt.Errorf("unknown provider")
	}
	return s.UserRepository.SaveUser(user)
}
