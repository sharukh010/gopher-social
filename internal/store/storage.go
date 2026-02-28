package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context,*sql.Tx, *User) error
		GetByID(context.Context, int64) (*User, error)
		Follow(context.Context, int64, int64) error
		UnFollow(context.Context, int64, int64) error
		CreateAndInvite(context.Context,*User,string,time.Duration) error
		createUserInvitation(context.Context,*sql.Tx,string,time.Duration,int64) error 
	}
	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostID(context.Context, int64) ([]Comment, error)
		Delete(context.Context, int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentStore{db},
	}
}

func withTx(db *sql.DB,ctx context.Context,fn func(*sql.Tx) error) error {
	tx,err := db.BeginTx(ctx,nil)
	if err != nil {
		return err
	}
	if err:= fn(tx); err != nil {
		_  = tx.Rollback()
		return err 
	}
	return tx.Commit()
	
}