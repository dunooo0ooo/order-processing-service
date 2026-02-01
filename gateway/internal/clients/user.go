package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserClient struct {
	base string
	c    *http.Client
}

func NewUserClient(base string) *UserClient {
	return &UserClient{
		base: base,
		c:    &http.Client{},
	}
}

func (u *UserClient) Register(r *http.Request) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, u.base+"/users", r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header.Clone()
	return u.c.Do(req)
}

func (u *UserClient) Verify(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := u.c.Post(u.base+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("verify failed: %s", string(b))
	}

	var v struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	return v.Role, nil
}
