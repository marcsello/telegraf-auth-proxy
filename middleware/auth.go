package middleware

import (
	"encoding/csv"
	httpAuth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

const BasicAuthUserKey = "bu"

func GetBasicAuthUserFromCtx(ctx *gin.Context) (string, bool) {
	bu, ok := ctx.Get(BasicAuthUserKey)
	if !ok {
		return "", false
	}
	var username string
	username, ok = bu.(string)
	if !ok {
		return "", false
	}
	return username, true
}

func loadBasicAuthCredentials(htpasswdPath string) (map[string]string, error) {
	// Adopted from here: https://github.com/abbot/go-http-auth/blob/master/users.go
	var err error
	var f *os.File
	f, err = os.Open(htpasswdPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ':'
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	var records [][]string
	records, err = reader.ReadAll()
	if err != nil {
		return nil, err
	}

	users := make(map[string]string)
	for _, record := range records {
		users[record[0]] = record[1]
	}

	return users, nil
}

func BasicAuthMiddleware(htpasswdPath string) gin.HandlerFunc {

	users, err := loadBasicAuthCredentials(htpasswdPath)
	if err != nil {
		panic(err)
	}

	return func(ctx *gin.Context) {
		username, password, ok := ctx.Request.BasicAuth()
		if !ok {
			// no credentials provided
			ctx.Header("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var storedHash string
		storedHash, ok = users[username]
		if !ok {
			// invalid user
			ctx.Header("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !httpAuth.CheckSecret(password, storedHash) {
			// invalid password
			ctx.Header("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// everything is cool and good, set the context value
		ctx.Set(BasicAuthUserKey, username)
	}
}
