package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"secretlinks/middleware"
	"secretlinks/storage"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Create(key string, link storage.Link, unique bool) bool {
	args := m.Called(key, link, unique)
	return args.Bool(0)
}

func (m *MockStorage) Get(key string) (storage.Link, bool) {
	args := m.Called(key)
	return args.Get(0).(storage.Link), args.Bool(1)
}

func (m *MockStorage) Update(key string, link storage.Link) {
	m.Called(key, link)
}

func (m *MockStorage) Delete(key string) {
	m.Called(key)
}

func (m *MockStorage) Cleanup() {
	m.Called()
}

func TestCreateHandler_Success(t *testing.T) {

	mockStorage := new(MockStorage)
	mockStorage.On("Create", mock.AnythingOfType("string"), mock.AnythingOfType("storage.Link"), true).Return(true).Once()

	form := url.Values{}
	form.Add("secret", "my secret")
	form.Add("expiration", "33")
	form.Add("maxviews", "7")

	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Contains(t, w.Body.String(), "http://")
	mockStorage.AssertExpectations(t)

	call := mockStorage.Calls[0]
	link := call.Arguments[1].(storage.Link)
	assert.Equal(t, middleware.EncryptText("my secret"), link.Secret)
	assert.Equal(t, 7, link.MaxViews)
	assert.WithinDuration(t, time.Now().Add(33*time.Minute), link.ExpiresAt, 2*time.Second)
}

func TestCreateHandler_DefaultValues(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("Create", mock.Anything, mock.Anything, true).Return(true)

	form := url.Values{}
	form.Add("secret", "new secret")

	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertExpectations(t)

	call := mockStorage.Calls[0]
	link := call.Arguments[1].(storage.Link)
	assert.Equal(t, middleware.EncryptText("new secret"), link.Secret)
	assert.Equal(t, 1, link.MaxViews)
	assert.WithinDuration(t, time.Now().Add(60*time.Minute), link.ExpiresAt, 2*time.Second)
}

func TestCreateHandler_EmptyValues(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("Create", mock.Anything, mock.Anything, true).Return(true)

	form := url.Values{}

	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusNotAcceptable, w.Code)
	assert.Contains(t, w.Body.String(), "Expected 'secret' value")
}

func TestCreateHandler_WrongValues(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("Create", mock.Anything, mock.Anything, true).Return(true)

	form := url.Values{}
	form.Add("sEcRet", "my secret")

	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusNotAcceptable, w.Code)
	assert.Contains(t, w.Body.String(), "Expected 'secret' value")
}
func TestCreateHandler_InvalidMethod(t *testing.T) {
	mockStorage := new(MockStorage)

	req := httptest.NewRequest("GET", "/create", nil)
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestCreateHandler_InvalidExpiration(t *testing.T) {
	mockStorage := new(MockStorage)

	form := url.Values{}
	form.Add("secret", "my secret")
	form.Add("expiration", "text")

	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusNotAcceptable, w.Code)
	assert.Contains(t, w.Body.String(), "Expected int value")
}

func TestCreateHandler_InvalidMaxViews(t *testing.T) {
	mockStorage := new(MockStorage)

	form := url.Values{}
	form.Add("secret", "my secret")
	form.Add("maxviews", "text")

	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusNotAcceptable, w.Code)
	assert.Contains(t, w.Body.String(), "Expected int value")
}

func TestCreateHandler_KeyGenerationRetry(t *testing.T) {
	mockStorage := new(MockStorage)

	mockStorage.On("Create", mock.Anything, mock.Anything, true).
		Return(false).Twice()
	mockStorage.On("Create", mock.Anything, mock.Anything, true).
		Return(true).Once()

	form := url.Values{"secret": []string{"retry test"}}
	req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler := CreateHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertNumberOfCalls(t, "Create", 3)
}

func TestRedirectHandler_Success(t *testing.T) {

	mockStorage := new(MockStorage)
	mockStorage.On("Get", "valid_key").Return(storage.Link{
		Secret:    middleware.EncryptText("secret_msg"),
		ExpiresAt: time.Now().Add(time.Hour),
		MaxViews:  3,
		Views:     0,
	}, true).Once()
	mockStorage.On("Update", "valid_key", mock.AnythingOfType("storage.Link")).Run(func(args mock.Arguments) {
		link := args.Get(1).(storage.Link)
		assert.Equal(t, 1, link.Views, "should increment view count")
		assert.Equal(t, middleware.EncryptText("secret_msg"), link.Secret, "should save secret_msg")
	}).Return().Once()

	form := url.Values{}
	form.Add("secret", "my secret")

	req := httptest.NewRequest("GET", "/valid_key", strings.NewReader(form.Encode()))
	w := httptest.NewRecorder()
	handler := RedirectHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestRedirectHandler_WrongValue(t *testing.T) {

	mockStorage := new(MockStorage)
	mockStorage.On("Get", "wrong_key").Return(storage.Link{}, false).Once()

	form := url.Values{}
	form.Add("secret", "my secret")

	req := httptest.NewRequest("GET", "/wrong_key", strings.NewReader(form.Encode()))
	w := httptest.NewRecorder()
	handler := RedirectHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestRedirectHandler_InvalidExpiration(t *testing.T) {

	mockStorage := new(MockStorage)
	mockStorage.On("Get", "valid_key_InvalidExpiration").Return(storage.Link{
		Secret:    middleware.EncryptText("secret_msg"),
		ExpiresAt: time.Now().Add(-time.Hour),
		MaxViews:  5,
		Views:     2,
	}, true).Once()
	mockStorage.On("Delete", "valid_key_InvalidExpiration").Return()

	form := url.Values{}
	form.Add("secret", "my secret")

	req := httptest.NewRequest("GET", "/valid_key_InvalidExpiration", strings.NewReader(form.Encode()))
	w := httptest.NewRecorder()
	handler := RedirectHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
	mockStorage.AssertExpectations(t)
	assert.Contains(t, w.Body.String(), "Link expired")
}

func TestRedirectHandler_InvalidMaxViews(t *testing.T) {

	mockStorage := new(MockStorage)
	mockStorage.On("Get", "valid_key_InvalidMaxViews").Return(storage.Link{
		Secret:    middleware.EncryptText("secret_msg"),
		ExpiresAt: time.Now().Add(time.Hour),
		MaxViews:  5,
		Views:     5,
	}, true).Once()
	mockStorage.On("Delete", "valid_key_InvalidMaxViews").Return()

	form := url.Values{}
	form.Add("secret", "my secret")

	req := httptest.NewRequest("GET", "/valid_key_InvalidMaxViews", strings.NewReader(form.Encode()))
	w := httptest.NewRecorder()
	handler := RedirectHandler(mockStorage)
	handler(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
	mockStorage.AssertExpectations(t)
	assert.Contains(t, w.Body.String(), "Link expired")
}
