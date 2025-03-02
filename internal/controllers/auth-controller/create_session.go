package authcontroller

import (
	httputils "auth/internal/controllers/http-utils"
	ginutils "auth/internal/controllers/http-utils/gin-utils"
	slogutils "auth/internal/utils/slog-utils"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const userIDParamName = "userID"

type createSessionResponseBody SessionDTO

//	@Summary		Create new session
//	@Description	Create new access and refresh tokens given user ID
//	@Tags			session
//	@Produce		json
//	@Param			userID	query		string						yes	"User ID"
//	@Success		201		{object}	createSessionResponseBody	"Success"
//	@Failure		400		{object}	httputils.HTTPError			"Bad request"
//	@Failure		500		{object}	httputils.HTTPError			"Internal server error"
//	@Router			/sessions [post]
func (controller *AuthController) createSession(c *gin.Context) {
	userID, err := uuid.Parse(c.Query(userIDParamName))
	if err != nil {
		ginutils.BadRequest(c, errors.Wrap(err, "parse userID"))
		return
	}

	session, err := controller.authService.
		CreateSession(
			userID,
			httputils.GetRequestIP(c.Request))
	if err != nil {
		slogutils.Error("create session", err)
		ginutils.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, createSessionResponseBody{
		AccessToken: string(session.AccessTokenSigned),
		RefreshToken: base64.StdEncoding.
			EncodeToString(session.RefreshTokenValue),
	})
}
