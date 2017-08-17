package coinbase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
	// "github.com/fabioberger/coinbase-go/config"
)

var _ = hex.DecodeString

var _ = fmt.Println

const BaseURL = "https://api.coinbase.com/v2/"

// ApiKeyAuthentication Struct implements the Authentication interface and takes
// care of authenticating RPC requests for clients with a Key & Secret pair
type apiKeyAuthentication struct {
	Key     string
	Secret  string
	BaseUrl string
	Client  http.Client
}

// ApiKeyAuth instantiates ApiKeyAuthentication with the API key & secret
func apiKeyAuth(key string, secret string) *apiKeyAuthentication {
	a := apiKeyAuthentication{
		Key:     key,
		Secret:  secret,
		BaseUrl: BaseURL,
		Client: http.Client{
			Transport: &http.Transport{
				Dial: dialTimeout,
			},
		},
	}
	return &a
}

// API Key + Secret authentication requires a request header of the HMAC SHA-256
// signature of the "message" as well as an incrementing nonce and the API key
func (a apiKeyAuthentication) authenticate(req *http.Request, body []byte) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + req.Method + req.URL.Path

	sha := sha256.New
	h := hmac.New(sha, []byte(a.Secret))
	h.Write(append([]byte(message), body...))

	signature := fmt.Sprintf("%x", h.Sum(nil))

	req.Header.Set("CB-ACCESS-KEY", a.Key)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("CB-VERSION", "2017-04-08")

	return nil
}

func (a apiKeyAuthentication) getBaseUrl() string {
	return a.BaseUrl
}

func (a apiKeyAuthentication) getClient() *http.Client {
	return &a.Client
}

// dialTimeout is used to enforce a timeout for all http requests.
func dialTimeout(network, addr string) (net.Conn, error) {
	var timeout = time.Duration(2 * time.Second) //how long to wait when trying to connect to the coinbase
	return net.DialTimeout(network, addr, timeout)
}
