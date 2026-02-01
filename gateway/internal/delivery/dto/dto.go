package dto

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type VerifyResp struct {
	Role string `json:"role"`
}
type LoginResp struct {
	AccessToken string `json:"access_token"`
}
