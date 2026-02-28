package main

import (
	"net/http"
)

type healthResponse struct {
	Status  string `json:"status"`
	Env     string `json:"env"`
	Version string `json:"version"`
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	API health check
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	healthResponse	"API is live"
//	@Failure		500	{object}	error			"Something went worng"
//	@Security		ApiKeyAuth
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := healthResponse{
		Status:  "ok",
		Env:     app.config.env,
		Version: version,
	}
	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
