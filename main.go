package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	log "github.com/sirupsen/logrus"
)

const (
	clientID       = "e492eaede804e5ac89e2"
	clientSecret   = "3ce3a35daec317529b6f6157cd65456581b94039"
	accessTokenURI = "https://github.com/login/oauth/access_token"
	userURI        = "https://api.github.com/user"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("frontend/*")

	router.GET("/", IndexHandler)
	router.GET("/oauth/redirect", OAuthHandler)

	router.Run(":8080")
}

// IndexHandler index handler for /
func IndexHandler(c *gin.Context) {
	// 1. Request a user's GitHub identity
	// GET https://github.com/login/oauth/authorize
	c.HTML(http.StatusOK, "index.tmpl", nil)
}

// OAuthHandler oauth handler for /oauth/redirect
// https://developer.github.com/apps/building-oauth-apps/authorizing-oauth-apps/
func OAuthHandler(c *gin.Context) {
	// 2. Users are redirected back to your site by GitHub
	// GET http://localhost:8080/oauth/redirect
	oauthCode := c.Query("code")
	log.WithField("oauthCode", oauthCode).Infoln("oauthCode info")

	if oauthCode == "" {
		log.Error("OAuth failed: the oauth code from github invalid")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  "Invalid oauth code",
		})
		return
	}

	// 3. Exchange this code for an access token
	// POST https://github.com/login/oauth/access_token
	log.Infoln("start access token...")
	_, body, errs := gorequest.New().Post(accessTokenURI).
		Param("client_id", clientID).
		Param("client_secret", clientSecret).
		Param("code", oauthCode).
		Set("accept", "application/json").
		End()

	if len(errs) > 0 {
		log.WithField("errs", errs).Error("access token failed")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  errs,
		})
		return
	}

	log.WithField("body", body).Infoln("access token info")

	bodyMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &bodyMap); err != nil {
		log.WithField("body", body).WithError(err).Error("json unmarshal body failed")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  err,
		})
		return
	}

	accessToken, ok := bodyMap["access_token"].(string)
	if !ok {
		log.WithFields(log.Fields{
			"accessToken": accessToken,
			"ok":          ok,
		}).Error("parse access token failed")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  "parse access token failed",
		})
		return
	}

	// 4. Use the access token to access the API
	// GET https://api.github.com/user
	log.Infoln("start access user api...")
	_, body, errs = gorequest.New().Get(userURI).
		Set("Authorization", "token "+accessToken).
		Set("accept", "application/json").
		End()

	if len(errs) > 0 {
		log.WithField("errs", errs).Error("get user info failed")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  errs,
		})
		return
	}

	log.WithField("body", body).Infoln("user info")

	userBodyMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &userBodyMap); err != nil {
		log.WithField("body", body).WithError(err).Errorln("json unmarshal user body failed")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  err,
		})
		return
	}

	userName, ok := userBodyMap["name"].(string)
	if !ok {
		log.WithFields(log.Fields{
			"userName": userName,
			"ok":       ok,
		}).Error("parse user name failed")
		c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
			"code": http.StatusInternalServerError,
			"err":  "parse user name failed",
		})
		return
	}

	c.HTML(http.StatusOK, "welcome.tmpl", gin.H{
		"userName": userName,
	})
}
