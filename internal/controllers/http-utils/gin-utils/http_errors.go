package ginutils

import (
	httputils "auth/internal/controllers/http-utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Error(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, httputils.HTTPError{Message: err.Error()})
}

func BadRequest(ctx *gin.Context, err error) {
	Error(ctx, http.StatusBadRequest, err)
}

func BindJSONError(ctx *gin.Context, err error) {
	BadRequest(ctx, errors.Wrap(err, "parse and validate JSON body"))
}

func BindQueryError(ctx *gin.Context, err error) {
	BadRequest(ctx, errors.Wrap(err, "parse and validate query params"))
}

func InternalError(ctx *gin.Context) {
	Error(ctx, http.StatusInternalServerError, errors.New(""))
}

func UnauthorizedError(ctx *gin.Context, err error) {
	Error(ctx, http.StatusUnauthorized, err)
}
