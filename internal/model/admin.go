package model

type UpdateStatusRequest struct {
	Id     string `form:"id" validate:"required"`
	Status string `form:"status" validate:"required,oneof=approved rejected"`
}

type UpdateStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
