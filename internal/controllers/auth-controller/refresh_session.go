package authcontroller

import (
	httputils "auth/internal/controllers/http-utils"
	ginutils "auth/internal/controllers/http-utils/gin-utils"
	"auth/internal/domain"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type refreshSessionRequestBody SessionDTO

type refreshSessionResponseBody SessionDTO

// @Summary		Refresh session
// @Description	Create a new access and refresh tokens pair from given access and refresh tokens.
// @Description	Provided refresh token is invalidated on success.
// @Description	Provided refresh token should have been issued with provided access token.
// @Description	Provided access token can be expired, but refresh token can't.
// @Tags			session
// @Accept			json
// @Produce		json
// @Param			access_and_refresh_tokens	body		refreshSessionRequestBody	yes	"Access and refresh tokens"
// @Success		201							{object}	refreshSessionResponseBody	"Success"
// @Failure		400							{object}	httputils.HTTPError			"Bad request"
// @Failure		401							{object}	httputils.HTTPError			"Unauthorized"
// @Failure		500							{object}	httputils.HTTPError			"Internal server error"
// @Router			/sessions/refresh [post]
func (controller *AuthController) refreshSession(c *gin.Context) {
	var reqBody refreshSessionRequestBody
	if err := c.BindJSON(&reqBody); err != nil {
		ginutils.BadRequest(c, err)
		return
	}
	refreshTokenDecoded, err := base64.StdEncoding.DecodeString(reqBody.RefreshToken)
	if err != nil {
		ginutils.BadRequest(c, errors.Wrap(err, "decode base64 refresh token value"))
		return
	}

	var unauthorizedError *domain.UnauthorizedError
	session, err := controller.authService.RefreshSession(
		&domain.Session{
			AccessTokenSigned: []byte(reqBody.AccessToken),
			RefreshTokenValue: refreshTokenDecoded,
		},
		httputils.GetRequestIP(c.Request))
	switch {
	case err == nil:
	case errors.As(err, &unauthorizedError):
		ginutils.UnauthorizedError(c, err)
		return
	default:
		ginutils.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, refreshSessionResponseBody{
		AccessToken: string(session.AccessTokenSigned),
		RefreshToken: base64.StdEncoding.
			EncodeToString(session.RefreshTokenValue),
	})
}
