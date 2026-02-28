package main

import (
	"net/http"

	"github.com/sharukh010/social/internal/store"
)

// GetUserFeed godoc
//
//	@Summary		Fetch User Feed
//	@Description	Fetch User Feed by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int							false	"Limit"
//	@Param			offset	query		int							false	"Offset"
//	@Param			sort	query		string						false	"Sort"
//	@Param			tags	query		string						false	"Tags"
//	@Success		200		{object}	[]store.PostWithMetadata	"User Feed"
//	@Failure		400		{object}	error						"Invalid Feed payload"
//	@Failure		404		{object}	error						"Feed not found"
//	@Failure		500		{object}	error						"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(104), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
