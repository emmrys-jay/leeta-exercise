package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"leeta/internal/adapter/logger"
	"leeta/internal/core/domain"
	"leeta/internal/core/port"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// LocationHandler represents the HTTP handler for location-related requests
type LocationHandler struct {
	svc      port.LocationService
	validate *validator.Validate
}

// NewLocationHandler creates a new LocationHandler instance
func NewLocationHandler(svc port.LocationService, vld *validator.Validate) *LocationHandler {
	return &LocationHandler{
		svc,
		vld,
	}
}

// RegisterUser godoc
//
//	@Summary		Register a new location
//	@Description	register a new location with all required details
//	@Tags			Location
//	@Accept			json
//	@Produce		json
//	@Param			domain.RegisterLocationRequest	body		domain.RegisterLocationRequest	true	"Location"
//	@Success		201								{object}	response						"Location created successfully"
//	@Failure		400								{object}	errorResponse					"Validation error"
//	@Failure		409								{object}	errorResponse					"Conflict error"
//	@Failure		500								{object}	errorResponse					"Internal server error"
//	@Router			/location [post]
func (ch *LocationHandler) RegisterLocation(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.FromCtx(r.Context()).Error("Error decoding json body", zap.Error(err))
		handleError(w, domain.ErrInternal)
		return
	}

	if err := ch.validate.Struct(&req); err != nil {
		validationError(w, err)
		return
	}

	result, cerr := ch.svc.RegisterLocation(r.Context(), &req)
	if cerr != nil {
		handleError(w, cerr)
		return
	}

	handleSuccessWithMessage(w, http.StatusCreated, result, "Location created successfully")
}

// GetLocation godoc
//
//	@Summary		Get a location by name
//	@Description	fetch a location through name
//	@Tags			Location
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string			true	"Location name"
//	@Success		200		{object}	response		"Success"
//	@Failure		400		{object}	errorResponse	"Validation error"
//	@Failure		500		{object}	errorResponse	"Internal server error"
//	@Router			/location/{name} [get]
//	@Security		BearerAuth
func (ch *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		handleError(w, domain.NewBadRequestCError("Invalid location name"))
		return
	}

	result, cerr := ch.svc.GetLocation(r.Context(), name)
	if cerr != nil {
		handleError(w, cerr)
		return
	}

	handleSuccess(w, http.StatusOK, result)
}

// ListLocations godoc
//
//	@Summary		List all locations
//	@Description	list all registered active locations
//	@Tags			Location
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response		"Success"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/locations [get]
//	@Security		BearerAuth
func (ch *LocationHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	results, cerr := ch.svc.ListLocations(r.Context())
	if cerr != nil {
		handleError(w, cerr)
		return
	}

	handleSuccess(w, http.StatusOK, results)
}

// Delete Location godoc
//
//	@Summary		Delete a location by name
//	@Description	delete a location through name
//	@Tags			Location
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string			true	"Location name"
//	@Success		200		{object}	response		"Success"
//	@Failure		400		{object}	errorResponse	"Validation error"
//	@Failure		404		{object}	errorResponse	"Not found error"
//	@Failure		500		{object}	errorResponse	"Internal server error"
//	@Router			/location/{name} [delete]
//	@Security		BearerAuth
func (ch *LocationHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		handleError(w, domain.NewBadRequestCError("Invalid location name"))
		return
	}

	cerr := ch.svc.DeleteLocation(r.Context(), name)
	if cerr != nil {
		handleError(w, cerr)
		return
	}

	handleSuccessWithMessage(w, http.StatusOK, nil, "Deleted location successfully")
}

// GetNearestLocation godoc
//
//	@Summary		Get the nearest location to the longitude and latitude
//	@Description	get the nearest location to the longitude and latitude
//	@Tags			Location
//	@Accept			json
//	@Produce		json
//	@Param			lat		query		float64	true	"Latitude"
//	@Param			lng		query		float64	true	"Longitude"
//	@Success		200		{object}	response		"Success"
//	@Failure		400		{object}	errorResponse	"Validation error"
//	@Failure		404		{object}	errorResponse	"Not found error"
//	@Failure		500		{object}	errorResponse	"Internal server error"
//	@Router			/location/nearest [get]
//	@Security		BearerAuth
func (ch *LocationHandler) GetNearestLocation(w http.ResponseWriter, r *http.Request) {
	latitude, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		handleError(w, domain.NewBadRequestCError("Invalid latitude"))
		return
	}

	longitude, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if err != nil {
		handleError(w, domain.NewBadRequestCError("Invalid longitude"))
		return
	}

	result, cerr := ch.svc.GetNearestLocation(r.Context(), latitude, longitude)
	if cerr != nil {
		handleError(w, cerr)
		return
	}

	handleSuccess(w, http.StatusOK, result)
}
