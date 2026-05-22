package models

type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name,omitzero"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type RegisterRequest struct {
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
}
