package service

import (
	"bpl/auth"
	"bpl/repository"

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

func (s *UserService) GetOrCreateUserByDiscordId(discordId int64, discordName string) (*repository.User, error) {
	user, err := s.UserRepository.GetUserByDiscordId(discordId)
	if err != nil {
		user = &repository.User{DiscordID: discordId, DiscordName: discordName, Permissions: repository.Permissions{}}
		user, err = s.UserRepository.SaveUser(user)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
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
