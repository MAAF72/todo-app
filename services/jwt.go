package services

import (
	"time"
	"todo-app/config"

	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/jwt"
)

var JWTServices JWT

type JWT struct {
	Signer           *jwt.Signer
	Verifier         *jwt.Verifier
	VerifyMiddleware context.Handler
	Signature        jwt.Alg
}

func (j *JWT) Init(conf config.JWTConfiguration) {
	j.Signature = jwt.HS256
	j.Signer = jwt.NewSigner(jwt.HS256, conf.Secret_Key, time.Duration(conf.Token_Duration)*time.Minute)
	j.Verifier = jwt.NewVerifier(jwt.HS256, conf.Secret_Key).WithDefaultBlocklist()
	j.VerifyMiddleware = j.Verifier.Verify(func() interface{} { return new(AuthClaims) })
}

type AuthClaims struct {
	Id int64
}
