package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var cognitoJwksURL = "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_POLxfwFzO/.well-known/jwks.json"

// JWK representa uma chave pública do JWKS
type JWK struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		Alg string `json:"alg"`
		Use string `json:"use"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// AuthMiddleware valida o token JWT do Cognito e injeta o userID no contexto
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token não fornecido"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato do token inválido"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid não encontrado no header do token")
			}

			return getKeyFromJWK(kid)
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Falha ao extrair claims do token"})
			c.Abort()
			return
		}

		userID, ok := claims["username"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token sem informações de usuário"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
func decodePublicKey(key jwkKey) (*rsa.PublicKey, error) {
	// 1) Decodifica o campo 'n' (Base64 URL) em bytes
	nDecoded, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar 'n': %w", err)
	}

	// 2) Decodifica o campo 'e' (Base64 URL) em bytes
	eDecoded, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar 'e': %w", err)
	}

	// Converte em big.Int
	n := big.NewInt(0).SetBytes(nDecoded)
	e := big.NewInt(0).SetBytes(eDecoded)

	// Normalmente 'e' é pequeno, ex: 65537 (0x10001)
	if e.BitLen() > 32 {
		return nil, fmt.Errorf("valor de 'e' muito grande")
	}

	// Cria a chave pública
	pubKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return pubKey, nil
}

func getKeyFromJWK(kid string) (interface{}, error) {
	resp, err := http.Get(cognitoJwksURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar JWKS: %v", err)
	}
	defer resp.Body.Close()

	var jwk JWK
	if err := json.NewDecoder(resp.Body).Decode(&jwk); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JWKS: %v", err)
	}

	// Procura a chave correspondente ao "kid" do token
	for _, key := range jwk.Keys {
		if key.Kid == kid {
			// Decodifica de Base64 para *rsa.PublicKey
			pubKey, err := decodePublicKey(key)
			if err != nil {
				return nil, err
			}
			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("chave pública não encontrada para kid: %s", kid)
}
