package repository

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Permission string

const (
	PermissionAdmin       Permission = "admin"
	PermissionCommandTeam Permission = "command_team"
)

type Permissions []Permission

func (p *Permissions) Scan(src interface{}) error {
	x := make(pq.StringArray, 0)
	x.Scan(src)
	permissions := make(Permissions, len(x))
	for i, perm := range x {
		permissions[i] = Permission(perm)
	}
	*p = permissions
	return nil
}

func (p Permissions) Value() (driver.Value, error) {
	permissions := make(pq.StringArray, len(p))
	for i, perm := range p {
		permissions[i] = string(perm)
	}
	return permissions.Value()
}

type User struct {
	ID                int         `gorm:"primaryKey autoIncrement"`
	AccountName       string      `gorm:"null"`
	DiscordID         int64       `gorm:"null"`
	DiscordName       string      `gorm:"null"`
	PoeToken          string      `gorm:"null"`
	PoeTokenExpiresAt int64       `gorm:"null"`
	Permissions       Permissions `gorm:"type:text[];not null;default:'{}'"`
}

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetUserById(userId int) (*User, error) {
	var user User
	query := r.DB

	result := query.First(&user, userId)
	if result.Error != nil {
		return nil, fmt.Errorf("user with id %d not found", userId)
	}
	return &user, nil
}

func (r *UserRepository) GetUserByDiscordId(discordId int64) (*User, error) {
	var user User
	query := r.DB

	result := query.First(&user, "discord_id = ?", discordId)
	if result.Error != nil {
		return nil, fmt.Errorf("user with discord id %d not found", discordId)
	}
	return &user, nil
}

func (r *UserRepository) SaveUser(user *User) (*User, error) {
	result := r.DB.Save(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}
	return user, nil
}

func (r *UserRepository) GetUsers() ([]*User, error) {
	var users []*User
	query := r.DB

	result := query.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users: %v", result.Error)
	}
	return users, nil
}
