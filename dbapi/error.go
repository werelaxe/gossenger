package dbapi

type ApiError struct {
	message string
}

func (apiError *ApiError) Error() string {
	return "Api error: " + apiError.message
}
