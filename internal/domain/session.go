package domain

type Session struct {
	AccessTokenSigned []byte
	RefreshTokenValue []byte
}
