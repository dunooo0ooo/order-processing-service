package dto

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type CreateUserResponse struct {
	IsSuccess bool `json:"isSuccess"`
}

type VerifyUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyUserResponse struct {
	Role string `json:"role"`
}
