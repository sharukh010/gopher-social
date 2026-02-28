package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
	Version   int       `json:"version"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count"`
}
type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content,title,user_id,tags)
		VALUES ($1,$2,$3,$4) RETURNING id,created_at,updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetByID(ctx context.Context, postID int64) (*Post, error) {
	query := `
	SELECT id,content,title,user_id,tags,version,
	created_at,updated_at FROM posts 
	WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(
		ctx,
		query,
		postID,
	).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	query := `
	DELETE FROM posts 
	WHERE id = $1 
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		postID,
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

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts 
		SET 
		title = $1,
		content = $2,
		tags = $3,
		updated_at = NOW(),
		version = version + 1 
		where id = $4 AND version = $5 
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Tags,
		post.ID,
		post.Version,
	).Scan(
		&post.Version,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil

}

func (s *PostStore) GetUserFeed(ctx context.Context, postID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	query := `
	select
	p.id,
	p.user_id,
	p.title,
	p.content,
	p.tags,
	p.version,
	p.created_at,
	p.updated_at,
	u.username,
	count(c.id) as comments_count
	from posts as p
	left join users as u on u.id = p.user_id
	left join comments as c on c.post_id = p.id
	left join followers as f on f.user_id = $1 and f.follower_id = p.user_id
	where (f.follower_id is not null or p.user_id = $1) and
	(p.title ILIKE '%' || $4 || '%' or p.content ILIKE '%' || $4 || '%') and
	(p.tags @> $5 or $5 = '{}' ) and
	(p.created_at between $6 and $7 or $6 IS NULL or $7 IS NULL)
	group by p.id,u.username
	order by p.created_at ` + fq.Sort + `
	limit $2 offset $3
	`
	feed := []PostWithMetadata{}
	var since *time.Time
	var until *time.Time
	var err error
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if fq.Since != "" {
		t, err := time.Parse(time.DateTime, fq.Since)
		if err != nil {
			return nil, err
		}
		since = &t
	}
	if fq.Until != "" {
		t, err := time.Parse(time.DateTime, fq.Until)
		if err != nil {
			return nil, err
		}
		until = &t
	}
	rows, err := s.db.QueryContext(
		ctx,
		query,
		postID,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(fq.Tags),
		since,
		until,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	defer rows.Close()
	for rows.Next() {
		post := PostWithMetadata{}

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			pq.Array(&post.Tags),
			&post.Version,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.User.Username,
			&post.CommentCount,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, post)
	}

	return feed, nil

}
