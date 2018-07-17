package models

import (
	"pickup/core/redis"
	"pickup/settings"
	jwt "github.com/dgrijalva/jwt-go"
	"time"	
	"gopkg.in/mgo.v2/bson"
	"pickup/helpers"
)


func (backend *JWTAuthenticationBackend) GenerateToken(userUUID bson.ObjectId) (string, error) {

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(settings.Get().JWTExpirationDelta)).Unix(),
		IssuedAt: time.Now().Unix(),
		Subject: helpers.GetIDEncoded(userUUID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	
	tokenString, err := token.SignedString(backend.PrivateKey)
	if err != nil {
		panic(err)
		return "", err
	}
	return tokenString, nil
}

func (backend *JWTAuthenticationBackend) getTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

func (backend *JWTAuthenticationBackend) Logout(tokenString string, token *jwt.Token) (err error) {
	redisConn := redis.Connect()

	return redisConn.SetValue(tokenString, tokenString, backend.getTokenRemainingValidity(token.Claims.(jwt.StandardClaims).ExpiresAt))
}

func (backend *JWTAuthenticationBackend) IsInBlacklist(token string) bool {
	return false 
	redisConn := redis.Connect()
	redisToken, _ := redisConn.GetValue(token)

	if redisToken == nil {
		return false
	}

	return true
}
