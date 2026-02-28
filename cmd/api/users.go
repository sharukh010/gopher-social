package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sharukh010/social/internal/store"
)

type userKey string

const userCtx userKey = "user"

const userURLParam = "userID"

type FollowRequest struct {
	UserID int64 `json:"user_id"`
}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int			true	"User ID"
//	@Success		200	{object}	store.User	"User Details"
//	@Failure		404	{object}	error		"User not found"
//	@Failure		500	{object}	error		"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		Follow a user
//	@Description	Follow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"User ID"
//	@Param			id	body		FollowRequest	true	"User ID"
//	@Success		204	{object}	nil				"Followed User"
//	@Failure		400	{object}	error			"Invalid Follow Payload"
//	@Failure		404	{object}	error			"User not found"
//	@Failure		500	{object}	error			"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)
	payload := &FollowRequest{}
	//TODO: revert back to auth userID from ctx
	err := readJSON(w, r, payload)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if followerUser.ID == payload.UserID {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()
	err = app.store.Users.Follow(ctx, followerUser.ID, payload.UserID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// UnFollowUser godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"User ID"
//	@Param			id	body		FollowRequest	true	"User ID"
//	@Success		204	{object}	nil				"Unfollowed User"
//	@Failure		400	{object}	error			"Invalid Unfollow Payload"
//	@Failure		404	{object}	error			"User not found"
//	@Failure		500	{object}	error			"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unFollowerUser := getUserFromCtx(r)
	//TODO: revert back to auth userID from ctx
	payload := &FollowRequest{}
	err := readJSON(w, r, payload)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.store.Users.UnFollow(ctx, unFollowerUser.ID, payload.UserID)

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

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paramID := chi.URLParam(r, userURLParam)
		userID, err := strconv.ParseInt(paramID, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userID)
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

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
