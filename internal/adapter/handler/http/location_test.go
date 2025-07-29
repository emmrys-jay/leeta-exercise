package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"leeta/internal/adapter/config"
	"leeta/internal/adapter/storage/postgres"
	"leeta/internal/adapter/storage/postgres/repository"
	"leeta/internal/core/domain"
	"leeta/internal/core/port"
	"leeta/internal/core/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *postgres.DB
var testHandler *LocationHandler
var testService port.LocationService

// setupTestDB initializes the test database and service
func setupTestDB(t *testing.T) {
	// Create test database configuration
	dbConfig := &config.DatabaseConfiguration{
		Protocol: "postgres",
		Host:     "localhost",
		Port:     "5433",
		User:     "postgres",
		Password: "postgres",
		Name:     "leeta_test",
	}

	// Connect to test database
	ctx := context.Background()
	var err error
	testDB, err = postgres.New(ctx, dbConfig)
	assert.NoError(t, err, "Failed to connect to test database")

	// Run migrations
	err = testDB.Migrate()
	assert.NoError(t, err, "Failed to run database migrations")

	// Create repository and service
	repo := repository.NewLocationRepository(testDB)
	testService = service.NewLocationService(repo)

	// Create handler
	validate := validator.New()
	testHandler = NewLocationHandler(testService, validate)
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T) {
	if testDB != nil {
		testDB.Close()
	}
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.Exec(ctx, "DELETE FROM locations")
	require.NoError(t, err, "Failed to cleanup test data")
}

func TestMain(m *testing.M) {
	// Setup test database
	setupTestDB(&testing.T{})

	// Run tests
	code := m.Run()

	// Cleanup
	teardownTestDB(&testing.T{})

	os.Exit(code)
}

func TestLocationHandler_RegisterLocation(t *testing.T) {
	cleanupTestData(t)

	t.Run("Success - Register new location", func(t *testing.T) {
		requestBody := domain.RegisterLocationRequest{
			Name:      "Test Location",
			Latitude:  40.7128,
			Longitude: -74.0060,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/locations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testHandler.RegisterLocation(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var res response
		err = json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.Equal(t, "Location created successfully", res.Message)
		assert.NotNil(t, res.Data)

		// Verify data structure
		data := res.Data.(domain.Location)
		assert.Equal(t, "Test Location", data.Name)
		assert.Equal(t, "test-location", data.Slug)
		assert.Equal(t, 40.7128, data.Latitude)
		assert.Equal(t, -74.0060, data.Longitude)
		assert.NotEmpty(t, data.ID)
		assert.NotEmpty(t, data.CreatedAt)
	})

	t.Run("Error - Missing required fields", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"latitude":  40.7128,
			"longitude": -74.0060,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/locations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testHandler.RegisterLocation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res errorResponse
		err = json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "Name")
	})

	t.Run("Error - Invalid latitude", func(t *testing.T) {
		requestBody := domain.RegisterLocationRequest{
			Name:      "Invalid Location",
			Latitude:  100.0, // Invalid latitude
			Longitude: -74.0060,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/locations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testHandler.RegisterLocation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res errorResponse
		err = json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "Latitude")
	})

	t.Run("Error - Duplicate location", func(t *testing.T) {
		// First, create a location
		requestBody := domain.RegisterLocationRequest{
			Name:      "Duplicate Test",
			Latitude:  40.7128,
			Longitude: -74.0060,
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/locations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testHandler.RegisterLocation(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Try to create the same location again
		req2 := httptest.NewRequest(http.MethodPost, "/location", bytes.NewBuffer(body))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()

		testHandler.RegisterLocation(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)

		var res errorResponse
		err = json.Unmarshal(w2.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
	})
}

func TestLocationHandler_GetLocation(t *testing.T) {
	cleanupTestData(t)

	res := createTestLocationViaHTTP(t, "Get Test Location", 40.7128, -74.0060)
	locationID := res.Data.(domain.Location).ID

	t.Run("Success - Get location by name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/Get Test Location", nil)
		w := httptest.NewRecorder()

		// Set up chi context with URL parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "Get Test Location")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		testHandler.GetLocation(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.NotNil(t, res.Data)

		data := res.Data.(domain.Location)
		assert.Equal(t, locationID, data.ID)
		assert.Equal(t, "Get Test Location", data.Name)
		assert.Equal(t, "get-test-location", data.Slug)
	})

	t.Run("Success - Get location by slug", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/get-test-location", nil)
		w := httptest.NewRecorder()

		// Set up chi context with URL parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "get-test-location")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		testHandler.GetLocation(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.NotNil(t, res.Data)

		data := res.Data.(domain.Location)
		assert.Equal(t, locationID, data.ID)
		assert.Equal(t, "Get Test Location", data.Name)
	})

	t.Run("Error - Empty location name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/", nil)
		w := httptest.NewRecorder()

		// Set up chi context with empty URL parameter
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		testHandler.GetLocation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
	})

	t.Run("Error - Location not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/non-existent", nil)
		w := httptest.NewRecorder()

		// Set up chi context with URL parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "non-existent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		testHandler.GetLocation(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var res errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
	})
}

func TestLocationHandler_ListLocations(t *testing.T) {
	cleanupTestData(t)

	// Create multiple test locations
	locations := []domain.RegisterLocationRequest{
		{
			Name:      "Location 1",
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
		{
			Name:      "Location 2",
			Latitude:  34.0522,
			Longitude: -118.2437,
		},
		{
			Name:      "Location 3",
			Latitude:  51.5074,
			Longitude: -0.1278,
		},
	}

	for _, loc := range locations {
		createTestLocationViaHTTP(t, loc.Name, loc.Latitude, loc.Longitude)
	}

	t.Run("Success - List all locations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/locations", nil)
		w := httptest.NewRecorder()

		testHandler.ListLocations(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.NotNil(t, res.Data)

		data := res.Data.([]domain.Location)
		assert.Len(t, data, 3)

		// Verify all locations are present
		names := make(map[string]bool)
		for _, item := range data {
			names[item.Name] = true
		}

		assert.True(t, names["Location 1"])
		assert.True(t, names["Location 2"])
		assert.True(t, names["Location 3"])
	})

	t.Run("Success - Empty list", func(t *testing.T) {
		cleanupTestData(t)

		req := httptest.NewRequest(http.MethodGet, "/locations", nil)
		w := httptest.NewRecorder()

		testHandler.ListLocations(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)

		data := res.Data.([]domain.Location)
		assert.Len(t, data, 0)
	})
}

func TestLocationHandler_DeleteLocation(t *testing.T) {
	cleanupTestData(t)

	res := createTestLocationViaHTTP(t, "Delete Test Location", 40.7128, -74.0060)
	locationName := res.Data.(domain.Location).Name

	t.Run("Success - Delete location by name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/location/"+locationName, nil)
		w := httptest.NewRecorder()

		// Set up chi context with URL parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "Delete Test Location")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		testHandler.DeleteLocation(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.Equal(t, "Deleted location successfully", res.Message)

		// Verify location is actually deleted
		getReq := httptest.NewRequest(http.MethodGet, "/location/"+locationName, nil)
		getW := httptest.NewRecorder()

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("name", locationName)
		getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rctx2))

		testHandler.GetLocation(getW, getReq)
		assert.Equal(t, http.StatusNotFound, getW.Code)

		var res2 errorResponse
		err = json.Unmarshal(getW.Body.Bytes(), &res2)
		require.NoError(t, err)

		assert.False(t, res2.Success)
	})

	t.Run("Error - Empty location name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/location/", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		testHandler.DeleteLocation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
	})
}

func TestLocationHandler_GetNearestLocation(t *testing.T) {
	cleanupTestData(t)

	// Create multiple test locations
	locations := []domain.RegisterLocationRequest{
		{
			Name:      "New York",
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
		{
			Name:      "Los Angeles",
			Latitude:  34.0522,
			Longitude: -118.2437,
		},
		{
			Name:      "London",
			Latitude:  51.5074,
			Longitude: -0.1278,
		},
	}

	for _, loc := range locations {
		response := createTestLocationViaHTTP(t, loc.Name, loc.Latitude, loc.Longitude)
		assert.True(t, response.Success)
	}

	t.Run("Success - Find nearest location to New York", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/nearest?lat=40.7589&lng=-73.9851", nil)
		w := httptest.NewRecorder()

		testHandler.GetNearestLocation(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.Equal(t, "Success", res.Message)
		assert.NotNil(t, res.Data)

		data := res.Data.(domain.NearestLocation)
		assert.Equal(t, "New York", data.Name)
		assert.NotEmpty(t, data.Distance)
		assert.Contains(t, data.Distance, "meters")
	})

	t.Run("Success - Find nearest location to Los Angeles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/nearest?lat=34.0522&lng=-118.2437", nil)
		w := httptest.NewRecorder()

		testHandler.GetNearestLocation(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.Equal(t, "Success", res.Message)
		assert.NotNil(t, res.Data)

		data := res.Data.(domain.NearestLocation)
		assert.Equal(t, "Los Angeles", data.Name)
		assert.NotEmpty(t, data.Distance)
	})

	t.Run("Error - Invalid latitude", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/nearest?lat=invalid&lng=-74.0060", nil)
		w := httptest.NewRecorder()

		testHandler.GetNearestLocation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
		assert.Equal(t, "Invalid latitude", res.Message)
	})

	t.Run("Error - Invalid longitude", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/location/nearest?lat=40.7128&lng=invalid", nil)
		w := httptest.NewRecorder()

		testHandler.GetNearestLocation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
		assert.Equal(t, "Invalid longitude", res.Message)
	})

	t.Run("Error - No locations found", func(t *testing.T) {
		cleanupTestData(t)

		req := httptest.NewRequest(http.MethodGet, "/location/nearest?lat=40.7128&lng=-74.0060", nil)
		w := httptest.NewRecorder()

		testHandler.GetNearestLocation(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var res errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "not found")
	})
}

// Helper function to create a test location via HTTP
func createTestLocationViaHTTP(t *testing.T, name string, lat, lng float64) response {
	requestBody := domain.RegisterLocationRequest{
		Name:      name,
		Latitude:  lat,
		Longitude: lng,
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/locations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.RegisterLocation(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var res response
	err = json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)

	return res
}
