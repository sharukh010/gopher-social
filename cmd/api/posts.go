package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sharukh010/social/internal/store"
)

type postKey string

const postCtx postKey = "post"

const postURLParam = "postID"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string   `json:"title" validate:"omitempty,max=100"`
	Content *string   `json:"content" validate:"omitempty,max=1000"`
	Tags    *[]string `json:"tags" validate:"omitempty"`
}

// CreatePost godoc
//
//	@Summary		Create a Post
//	@Description	Creates a Post and return Post details
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			post	body		CreatePostPayload	true	"Post details"
//	@Success		201		{object}	store.Post			"Post Created"
//	@Failure		400		{object}	error				"Invalid Post Payload"
//	@Failure		500		{object}	error				"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/posts/ [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	user_Id := 1
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		UserID:  int64(user_Id),
		Tags:    payload.Tags,
	}
	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// GetPost godoc
//
//	@Summary		Fetch Post
//	@Description	Fetch Post details by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int			true	"Post ID"
//	@Success		200	{object}	store.Post	"Post Details"
//	@Failure		404	{object}	error		"Post Not found"
//	@Failure		500	{object}	error		"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	ctx := r.Context()

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Delete Post
//	@Description	Delete Post details by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int		true	"Post ID"
//	@Success		204	{object}	nil		"Post Deleted"
//	@Failure		404	{object}	error	"Post not found"
//	@Failure		500	{object}	error	"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	ctx := r.Context()
	err := app.store.Posts.Delete(ctx, post.ID)

	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}

	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePost godoc
//
//	@Summary		Update Post
//	@Description	Update Post details by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			post	body		UpdatePostPayload	true	"Updated Post details"
//	@Success		201		{object}	store.Post			"Post Updated"
//	@Failure		400		{object}	error				"Invalid Post Payload"
//	@Failure		404		{object}	error				"Post not found"
//	@Failure		500		{object}	error				"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload UpdatePostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := getPostFromCtx(r)

	// the reason why we use != nil is some time "" can be valid
	// it is valid incase you want to delete title or contents or tags

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Tags != nil {
		post.Tags = *payload.Tags
	}

	ctx := r.Context()
	if err := app.store.Posts.Update(ctx, post); err != nil {
		envelop := map[string]string{
			"error": "the post has been modified by other user, try again",
		}
		switch err {
		case store.ErrNotFound:
			writeJSON(w, http.StatusConflict, envelop)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}

	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, postURLParam)
		postID, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, postID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, postCtx, post)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
