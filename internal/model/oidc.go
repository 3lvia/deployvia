package model

import (
	"context"
	"fmt"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type ValidatedClaims struct {
	RepositoryOwner string
	Repository      string
}

// Fetch the JWKS and use it to create a verification key function.
func createKeyFunc(ctx context.Context, jwksURL string) (jwt.Keyfunc, error) {
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc: %v", err)
	}

	return k.Keyfunc, nil
}

func verifyToken(tokenString string, keyFunc jwt.Keyfunc) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func validateClaims(claims jwt.MapClaims) (*ValidatedClaims, error) {
	iss, err := claims.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("error getting issuer: %s", err)
	}

	if iss != "https://token.actions.githubusercontent.com" {
		return nil, fmt.Errorf("invalid issuer: %s", iss)
	}

	aud, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("error getting audience: %s", err)
	}

	if len(aud) == 0 {
		return nil, fmt.Errorf("audience is empty")
	}

	if aud[0] != "https://github.com/3lvia" {
		return nil, fmt.Errorf("invalid audience: %s", aud[0])
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("error getting expiration time: %s", err)
	}

	if exp.IsZero() {
		return nil, fmt.Errorf("expiration time is zero")
	}

	if exp.Before(time.Now()) {
		return nil, fmt.Errorf("token is expired")
	}

	iat, err := claims.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("error getting issued at: %s", err)
	}

	if iat.IsZero() {
		return nil, fmt.Errorf("issued at time is zero")
	}

	if iat.After(time.Now()) {
		return nil, fmt.Errorf("token is not yet valid")
	}

	repositoryOwner, ok := claims["repository_owner"].(string)
	if !ok {
		return nil, fmt.Errorf("repository_owner claim is missing or not a string")
	}

	if repositoryOwner != "3lvia" {
		return nil, fmt.Errorf("repository owner %s is not valid", repositoryOwner)
	}

	repository, ok := claims["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("repository claim is missing or not a string")
	}

	return &ValidatedClaims{
		RepositoryOwner: repositoryOwner,
		Repository:      repository,
	}, nil
}

func ValidateToken(ctx context.Context, tokenString string, jwksURL string) (*ValidatedClaims, error) {
	keyFunc, err := createKeyFunc(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc: %v", err)
	}

	token, err := verifyToken(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	validatedClaims, err := validateClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to validate claims: %v", err)
	}

	return validatedClaims, nil
}
