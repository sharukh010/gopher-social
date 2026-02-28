package store

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  password `json:"-"`
	CreatedAt string `json:"created_at"`
}

type password struct {
	// text *string 
	hash []byte
}

func (p *password) Set(text string) error {
	hash,err := bcrypt.GenerateFromPassword([]byte(text),bcrypt.DefaultCost)
	if err != nil {
		return err 
	}
	// p.text = &text 
	p.hash = hash 
	return nil 
}


func (p *password) Compare(text string) error {
	if err := bcrypt.CompareHashAndPassword(p.hash,[]byte(text)); err != nil {
		return err 
	}
	return nil 
}
type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `
	INSERT INTO USERS
	(username,email,password)
	VALUES ($1,$2,$3)
	RETURNING id,created_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*User, error) {
	query := `
	SELECT id,username,email,password,created_at 
	FROM users 
	WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil

}

func (s *UserStore) Follow(ctx context.Context, followerUserID, followingUserID int64) error {
	query := `
	INSERT into followers 
	(user_id,follower_id)
	Values
	($1,$2)
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		followerUserID,
		followingUserID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) UnFollow(ctx context.Context, followerUserID, followingUserID int64) error {
	query := `
	DELETE FROM followers 
	WHERE user_id = $1 and follower_id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		followerUserID,
		followingUserID,
	)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context,user *User,token string) error {
	// transaction wrapper 
		// create a user 
		// create a invite
	return nil 
}