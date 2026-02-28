package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
	"github.com/sharukh010/social/internal/store"
)

type RegisterUserPayload struct {
	UserName string `json:"username" validate:"requried,max=100"`
	Email string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// RegisterUser godoc
//
//	@Summary		Register a User
//	@Description	Register a User
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			post	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	store.User			"User Registered"
//	@Failure		400		{object}	error				"Invalid User Payload"
//	@Failure		500		{object}	error				"Something went wrong"
//	@Security		ApiKeyAuth
//	@Router			/authenticate/user [user]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request){
	payload := &RegisterUserPayload{}
	if err := readJSON(w,r,payload); err != nil {
		app.badRequestResponse(w,r,err)
		return 
	}
	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w,r,err)
		return 
	}

	user := &store.User{
		Username: payload.UserName,
		Email: payload.Email,
	}
	//hasing the password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w,r,err)
		return 
	}
	
	ctx := r.Context()

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])
	if err := app.store.Users.CreateAndInvite(ctx,user,hashToken,app.config.mail.exp); err != nil {
		app.internalServerError(w,r,err)
		return 
	}
	
	// store the user 
	if err := app.jsonResponse(w,http.StatusCreated,nil); err != nil {
		app.internalServerError(w,r,err)
		return 
	}
}