package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	TraceID string      `json:"trace_id"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}

type SuccessResult struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type ErrorResult struct {
	Errors []FieldError `json:"errors,omitempty"`
}

type Pagination struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalPages   int   `json:"total_pages"`
	TotalRecords int64 `json:"total_records"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		return traceID.(string)
	}
	return "00000000-0000-0000-0000-000000000000"
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		TraceID: getTraceID(c),
		Status:  "success",
		Message: message,
		Result: SuccessResult{
			Data: data,
		},
	})
}

func SuccessWithPagination(c *gin.Context, message string, data interface{}, pagination *Pagination) {
	c.JSON(http.StatusOK, Envelope{
		TraceID: getTraceID(c),
		Status:  "success",
		Message: message,
		Result: SuccessResult{
			Data:       data,
			Pagination: pagination,
		},
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{
		TraceID: getTraceID(c),
		Status:  "success",
		Message: message,
		Result: SuccessResult{
			Data: data,
		},
	})
}

func BadRequest(c *gin.Context, message string, errors []FieldError) {
	c.JSON(http.StatusBadRequest, Envelope{
		TraceID: getTraceID(c),
		Status:  "error",
		Message: message,
		Result: ErrorResult{
			Errors: errors,
		},
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Envelope{
		TraceID: getTraceID(c),
		Status:  "error",
		Message: message,
		Result:  ErrorResult{},
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Envelope{
		TraceID: getTraceID(c),
		Status:  "error",
		Message: message,
		Result:  ErrorResult{},
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Envelope{
		TraceID: getTraceID(c),
		Status:  "error",
		Message: message,
		Result:  ErrorResult{},
	})
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Envelope{
		TraceID: getTraceID(c),
		Status:  "error",
		Message: message,
		Result:  ErrorResult{},
	})
}
