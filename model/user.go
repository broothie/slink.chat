package model

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const TypeUser Type = "user"

var userContextKey struct{}

type User struct {
	ID        string    `firestore:"id" json:"userID"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`

	Screenname     string `firestore:"screenname" json:"screenname"`
	PasswordDigest []byte `firestore:"password_digest" json:"-"`
}

func (User) Type() Type {
	return TypeUser
}

func (u *User) UpdatePassword(password string) error {
	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "failed to generate password digest")
	}

	u.PasswordDigest = passwordDigest
	return nil
}

func (u *User) PasswordsMatch(password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(u.PasswordDigest, []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}

		return false, errors.Wrap(err, "failed to compare passwords")
	}

	return true, nil
}

func (u *User) OnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, userContextKey, *u)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}
