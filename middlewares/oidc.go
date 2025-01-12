package middlewares

import (
	"context"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/storage/redis/v3"
	"go.oease.dev/goe/core"
	"go.oease.dev/goe/utils"
	"go.oease.dev/goe/webresult"
	"golang.org/x/oauth2"
	"runtime"
	"time"
)

type OIDCMiddleware struct {
	cfg             *OIDCMiddlewareConfig
	oauthStateStore *redis.Storage
	oauthConfig     *oauth2.Config
	oidcProvider    *oidc.Provider
}

type OIDCMiddlewareConfig struct {
	CallbackRedirectUri string
}

var defaultOIDCMiddlewareConfig = OIDCMiddlewareConfig{
	CallbackRedirectUri: "/auth/login?callback=true",
}

// OAuthClaimDataProcessor is a function type to process claim data from OIDC token, and return the data that will be stored in session as user info.
// The function can be used to check user permission, roles, or fetch user data from database. If the function returns an error, the login process will be failed.
type OAuthClaimDataProcessor func(claimData map[string]any) (any, error)

func NewOIDCMiddleware(config ...OIDCMiddlewareConfig) *OIDCMiddleware {
	cfg := defaultOIDCMiddlewareConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.CallbackRedirectUri == "" {
		cfg.CallbackRedirectUri = defaultOIDCMiddlewareConfig.CallbackRedirectUri
	}

	oidcAppId := core.UseGoeConfig().OIDC.AppId
	oidcAppSecret := core.UseGoeConfig().OIDC.AppSecret
	oidcIssuer := core.UseGoeConfig().OIDC.Issuer
	appScopes := core.UseGoeConfig().OIDC.AppScopes

	if oidcAppId == "" || oidcAppSecret == "" || oidcIssuer == "" {
		panic("OIDC config is not set properly!")
	}

	if len(appScopes) == 0 {
		appScopes = []string{"openid", "profile", "email"}
	}

	stateStore := redis.New(redis.Config{
		Host:     core.UseGoeConfig().Redis.Host,
		Port:     core.UseGoeConfig().Redis.Port,
		Username: core.UseGoeConfig().Redis.Username,
		Password: core.UseGoeConfig().Redis.Password,
		Database: core.RedisDBAuthOAuthState,
		PoolSize: 10 * runtime.GOMAXPROCS(0),
	})

	oidcProvider, err := oidc.NewProvider(context.Background(), oidcIssuer)
	if err != nil {
		panic(err)
	}
	oauthConfig := &oauth2.Config{
		ClientID:     oidcAppId,
		ClientSecret: oidcAppSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  cfg.CallbackRedirectUri,
		Scopes:       appScopes,
	}
	initSessionStore()
	return &OIDCMiddleware{
		cfg:             &cfg,
		oauthStateStore: stateStore,
		oauthConfig:     oauthConfig,
		oidcProvider:    oidcProvider,
	}
}

func (m *OIDCMiddleware) HandleLogin() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authRequestStateKey := utils.GenXid()
		loginUri := m.oauthConfig.AuthCodeURL(authRequestStateKey)
		if loginUri == "" {
			return webresult.SendFailed(ctx, "Failed to create auth request")
		}
		err := m.oauthStateStore.Set(authRequestStateKey, []byte(loginUri), time.Minute*5)
		if err != nil {
			return webresult.SystemBusy(err)
		}
		return webresult.SendSucceed(ctx, fiber.Map{
			"redirect": loginUri,
		})
	}
}

func (m *OIDCMiddleware) HandleLoginCallback(claimDataProcFunc ...OAuthClaimDataProcessor) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		sess := UseSession(ctx)
		if sess == nil {
			return webresult.SystemBusy(errors.New("session not configured"))
		}

		if ctx.Query("error") != "" {
			errMsg := ctx.Query("error_description")
			if errMsg == "" {
				errMsg = "OAuth provider returned unknown error"
			}
			return webresult.SendFailed(ctx, errMsg)
		}

		// Check if user is already logged in
		if !sess.Fresh() && sess.Get("user") != nil {
			//// User is already logged in, return user info from session data
			//userData := sess.Get("user")
			//if userData == nil {
			//	return webresult.SendFailed(ctx, "user data not found in session")
			//}
			//userInfo := make(map[string]any)
			//err := json.Unmarshal(userData.([]byte), &userInfo)
			//if err != nil {
			//	return webresult.SystemBusy(err)
			//}
			//return webresult.SendSucceed(ctx, userInfo)
			// user already logged in, but reset session to force re-login and update user information
			_ = sess.Reset()
		}

		// User is not logged in, handle sign-in callback

		// check state
		state := ctx.Query("state")
		if state == "" {
			return webresult.SendFailed(ctx, "Invalid state")
		}
		stateBytes, err := m.oauthStateStore.Get(state)
		if err != nil {
			return webresult.SystemBusy(err)
		}
		if stateBytes == nil {
			return webresult.SendFailed(ctx, "Login request expired or invalid")
		}

		//state check passed, verify login callback
		code := ctx.Query("code")
		if code == "" {
			return webresult.SendFailed(ctx, "Invalid OAuth code")
		}

		// Exchange code for tokens
		token, err := m.oauthConfig.Exchange(ctx.Context(), code)
		if err != nil {
			return webresult.SystemBusy(err)
		}
		if !token.Valid() {
			return webresult.SendFailed(ctx, "Invalid OAuth token, or token expired")
		}

		rawIdToken, ok := token.Extra("id_token").(string)
		if !ok {
			return webresult.SendFailed(ctx, "Failed to fetch ID Token")
		}

		verifiedIdToken, err := m.oidcProvider.Verifier(&oidc.Config{
			ClientID:                   m.oauthConfig.ClientID,
			SkipClientIDCheck:          false,
			SkipExpiryCheck:            false,
			SkipIssuerCheck:            false,
			InsecureSkipSignatureCheck: false,
		}).Verify(ctx.Context(), rawIdToken)
		if err != nil {
			return webresult.SystemBusy(err)
		}

		//err = verifiedIdToken.VerifyAccessToken(token.AccessToken)
		//if err != nil {
		//	return webresult.SystemBusy(err)
		//}

		claimData := make(map[string]any)
		err = verifiedIdToken.Claims(&claimData)
		if err != nil {
			return webresult.SystemBusy(err)
		}
		var sessionUserData any
		// Process claim data
		if len(claimDataProcFunc) > 0 {
			// Process claim data with custom processor
			sessionUserData, err = claimDataProcFunc[0](claimData)
			if err != nil {
				return webresult.SendFailed(ctx, err.Error())
			}
		} else {
			// Process claim data with default processor
			sessionUserData, err = defaultClaimDataProcessor(claimData)
			if err != nil {
				return webresult.SendFailed(ctx, err.Error())
			}
		}

		//save user info to session
		codedUserInfo, err := json.Marshal(sessionUserData)
		if err != nil {
			return webresult.SystemBusy(err)
		}

		sid := sess.ID()
		sess.Set("sid", sid)
		sess.Set("user", codedUserInfo)
		sess.Set("ip", ctx.IP())
		sess.Set("ua", string(ctx.Request().Header.UserAgent()))
		if token.AccessToken != "" {
			sess.Set("access_token", token.AccessToken)
		}

		// Fiber v3 beta.4 using session handler, manually saving no longer needed
		//// Save session
		//if err := sess.Save(); err != nil {
		//	return webresult.SystemBusy(err)
		//}

		// login success, invalid old state and redirect to home page
		err = m.oauthStateStore.Delete(state)
		if err != nil {
			return webresult.SystemBusy(err)
		}

		return webresult.SendSucceed(ctx, sessionUserData)
	}
}

var defaultClaimDataProcessor = func(claimData map[string]any) (any, error) {
	core.UseGoeContainer().GetLogger().Debug("Using default claim data processor.")
	// Default claim data processor, means that the claim data will be stored in session as is
	return claimData, nil
}
