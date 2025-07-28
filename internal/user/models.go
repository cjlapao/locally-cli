package user

type CreateUserRequest struct {
	Name     string `json:"name" yaml:"name" validate:"required"`
	Username string `json:"username" yaml:"username" validate:"required"`
	Password string `json:"password" yaml:"password" validate:"required,password_complexity"`
	Email    string `json:"email" yaml:"email" validate:"required,email"`
	Role     string `json:"role" yaml:"role" validate:"required"`
}

type CreateUserResponse struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" yaml:"name"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Email    string `json:"email" yaml:"email"`
	Role     string `json:"role" yaml:"role"`
}

type UpdateUserResponse struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
}

type UpdateUserPasswordRequest struct {
	Password string `json:"password" yaml:"password" validate:"required,password_complexity"`
}
