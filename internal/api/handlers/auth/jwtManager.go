package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"merch/internal/entity"
	"strconv"
	"time"
)

type JWTManager struct {
	secret        string
	tokenDuration time.Duration
	issuer        string
}

func NewJWTManager(s, i string, d time.Duration) *JWTManager {
	return &JWTManager{
		secret:        s,
		tokenDuration: d,
		issuer:        i,
	}
}

func (m *JWTManager) Generate(user entity.User, customClaims map[string]string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"iss": m.issuer,
		"exp": jwt.NewNumericDate(now.Add(m.tokenDuration)),
		"iat": jwt.NewNumericDate(now),
	}

	for k, v := range customClaims {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) Parse(token string, withOUtValidation bool) (string, map[string]string) {
	originClaims, err := m.parseToken(token, withOUtValidation)
	if err != nil {
		return "", map[string]string{"error": err.Error()}
	}

	subject, err := originClaims.GetSubject()
	if err != nil {
		return "", map[string]string{"error": err.Error()}
	}

	result := make(map[string]string)
	for key, value := range originClaims {
		switch v := value.(type) {
		case string:
			result[key] = v
		case float64:
			result[key] = strconv.Itoa(int(v))
		default:
			result[key] = fmt.Sprint(v)
		}
	}

	return subject, result
}

func (m *JWTManager) parseToken(token string, withoutValidation bool) (jwt.MapClaims, error) {
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Name}),
		jwt.WithIssuer(m.issuer),
	}

	if withoutValidation {
		options = append(options, jwt.WithoutClaimsValidation())
	}

	parser := jwt.NewParser(options...)

	parsed, err := parser.Parse(token, func(t *jwt.Token) (interface{}, error) {
		//_, err := t.Claims.GetSubject()
		//if err != nil {
		//	return nil, err
		//}
		return []byte(m.secret), nil
	})

	if err != nil || parsed == nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	//TODO: this
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
