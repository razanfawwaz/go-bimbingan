package model

type StudentForm struct {
	Name          string `form:"name" validate:"required"`
	NPM           string `form:"npm" validate:"required,numeric"`
	FieldInterest string `form:"field_interest" validate:"required"`
	ProjectTitle  string `form:"project_title" validate:"required"`
	Batch         string `form:"batch" validate:"required"`
	Token         string `form:"token" validate:"required"`
	ProjectLink   string `form:"project_link" validate:"required,url`
	ProfileLink   string `form:"profile_link" validate:"required,url`
	IsGraduated   string `form:"is_graduated" validate:"required,oneof=graduated not_graduated"`
}
