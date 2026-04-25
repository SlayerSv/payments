package interceptors

type ctxKey string

const (
	JWTKey ctxKey = "jwt_token"
	UserID ctxKey = "user_id"
)
