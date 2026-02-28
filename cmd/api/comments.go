package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sharukh010/social/internal/store"
)

type CreateCommentPayload struct {
	UserID  int64  `json:"user_id" validate:"required"`
	Content string `json:"content" validate:"required,min=6,max=100"`
}

// CreateComment godoc
//
//	@Summary		Create a Comment
//	@Description	Creates a Post and return details
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Post ID"
//	@Param			comment	body		CreateCommentPayload	true	"Comment"
//	@Success		201		{object}	store.Comment			"Comment Created"
//	@Failure		400		{object}	error					"Invalid Comment Payload"
//	@Failure		404		{object}	error					"Post not found"
//	@Failure		500		{object}	error					"Something Went wrong"
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	post := getPostFromCtx(r)

	comment := &store.Comment{
		PostID:  post.ID,
		UserID:  payload.UserID,
		Content: payload.Content,
	}

	ctx := r.Context()

	if err := app.store.Comments.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeleteComment godoc
//
//	@Summary		Delete Comment
//	@Description	Delete Comment details by ID
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			postID		path		int		true	"Post ID"
//	@Param			commentID	path		int		true	"Comment ID"
//	@Success		204			{object}	nil		"Comment Deleted"
//	@Failure		404			{object}	error	"Comment Not found"
//	@Failure		500			{object}	error	"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID}/comments/{commentID} [delete]
func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "commentID")
	commentID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.store.Comments.Delete(ctx, commentID)

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
