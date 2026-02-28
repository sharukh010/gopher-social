package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
	INSERT INTO comments 
	(post_id,user_id,content)
	VALUES 
	($1,$2,$3) RETURNING id,created_at 
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	comment.User.ID = comment.UserID
	if err != nil {
		return err
	}
	return nil
}
func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
	SELECT c.id,c.post_id,c.user_id,c.content,c.created_at,users.username,
	users.id FROM 
	comments c JOIN users on users.id = c.user_id 
	Where c.post_id = $1
	order by c.created_at DESC
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var c Comment
		c.User = User{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.ID); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (s *CommentStore) Delete(ctx context.Context, commentID int64) error {
	query := `
	DELETE from posts 
	WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		commentID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
