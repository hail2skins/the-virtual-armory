package auth

import (
	"context"

	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/volatiletech/authboss/v3"
)

// GORMStorer is a storer implementation for Authboss using GORM
type GORMStorer struct{}

// NewGORMStorer creates a new GORMStorer
func NewGORMStorer() *GORMStorer {
	return &GORMStorer{}
}

// Load loads a user from the database
func (g *GORMStorer) Load(ctx context.Context, key string) (authboss.User, error) {
	db := database.GetDB()
	var user models.User

	result := db.Where("email = ?", key).First(&user)
	if result.Error != nil {
		return nil, authboss.ErrUserNotFound
	}

	return NewUserWrapper(&user), nil
}

// Save saves a user to the database
func (g *GORMStorer) Save(ctx context.Context, user authboss.User) error {
	db := database.GetDB()
	u := user.(*UserWrapper)

	result := db.Save(u.User)
	return result.Error
}

// New creates a new user
func (g *GORMStorer) New(ctx context.Context) authboss.User {
	return NewUserWrapper(&models.User{})
}

// Create creates a new user in the database
func (g *GORMStorer) Create(ctx context.Context, user authboss.User) error {
	db := database.GetDB()
	u := user.(*UserWrapper)

	result := db.Create(u.User)
	return result.Error
}
