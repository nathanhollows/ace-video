package models

import (
	"context"
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/nathanhollows/ace-video/sessions"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	baseModel

	ID       string `bun:",pk,type:varchar(36)" json:"user_id"`
	Email    string `bun:",unique,pk" json:"email"`
	Password string `bun:",type:varchar(255)" json:"password"`
}

type Users []*User

// NewUser creates a new user
func NewUser(email, password string) (user *User) {
	user = &User{}
	user.ID = uuid.New().String()
	user.Email = email
	user.SetPassword(password)
	return user
}

// Save the user to the database
func (u *User) Save() error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(u).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// AuthenticateUser checks the user's credentials and returns the user if they are valid
func AuthenticateUser(email, password string) (*User, error) {
	// Find the user by email
	user, err := FindUserByEmail(email)
	if err != nil {
		log.Error("Error finding user: ", err)
		return nil, err
	}

	// Check the password
	if !user.checkPassword(password) {
		log.Error("Invalid password")
		return nil, errors.New("invalid password")
	} else {
		return user, nil
	}
}

// FindUserByEmail finds a user by their email address
func FindUserByEmail(email string) (*User, error) {
	ctx := context.Background()
	// Find the user by email
	user := &User{}
	err := db.NewSelect().
		Model(user).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindUserByID finds a user by their user id
func FindUserByID(userID string) (*User, error) {
	ctx := context.Background()
	// Find the user by user id
	user := &User{}
	err := db.NewSelect().
		Model(user).
		Where("User.id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CheckPassword checks if the given password is correct
func (u *User) checkPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		log.Error("Error comparing password: ", err)
		return false
	}
	return true
}

// SetPassword sets the user's password
func (u *User) SetPassword(password string) {
	u.Password = hashAndSalt(password)
}

// hashAndSalt hashes and salts the given password
func hashAndSalt(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error hashing password: ", err)
	}

	return string(hash)
}

// FindUserBySession finds the user by the session
func FindUserBySession(r *http.Request) (*User, error) {
	// Get the session
	session, err := sessions.Get(r, "admin")
	if err != nil {
		return nil, err
	}

	// Get the user id from the session
	userID, ok := session.Values["user_id"].(string)
	if !ok {
		return nil, errors.New("User not found")
	}

	// Find the user by the user id
	user, err := FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CheckAnyUsers checks if any users exist
func CheckAnyUsers() (bool, error) {
	ctx := context.Background()
	// Check if any users exist
	user := &User{}
	count, err := db.NewSelect().
		Model(user).
		Count(ctx)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
