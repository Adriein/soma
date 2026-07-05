package vendor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"maps"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adriein/soma/app/pkg/constants"
	"github.com/rotisserie/eris"
)

const (
	GetRequestTokenURL               = "https://authentication.fatsecret.com/oauth/request_token"
	AuthorizeTokenURL                = "https://authentication.fatsecret.com/oauth/authorize"
	GetAccessTokenURL                = "https://authentication.fatsecret.com/oauth/access_token"
	OAuthSignatureMethod             = "HMAC-SHA1"
	OAuthVersion                     = "1.0"
	FSGetTokenResOAuthTokenKey       = "oauth_token"
	FSGetTokenResOAuthTokenSecretKey = "oauth_token_secret"
)

type NutritionDiary interface {
	GetToken() (*OAuth, error)
	AuthorizeToken(oauth *OAuth) (*string, error)
	VerifyToken(oauth *OAuth) (*OAuth, error)
}

type OAuth struct {
	OAuthToken       string
	OAuthTokenSecret string
	OauthVerifyCode  int
}

type DiaryMeal struct {
	ID                 int    `json:"food_entry_id"`
	Calcium            string `json:"calcium"`
	Calories           string `json:"calories"`
	Carbohydrate       string `json:"carbohydrate"`
	Cholesterol        string `json:"cholesterol"`
	DateInt            string `json:"date_int"`
	Fat                string `json:"fat"`
	Fiber              string `json:"fiber"`
	Description        string `json:"food_entry_description"`
	Name               string `json:"food_entry_name"`
	FoodID             string `json:"food_id"`
	Iron               string `json:"iron"`
	Meal               string `json:"meal"`
	MonounsaturatedFat string `json:"monounsaturated_fat"`
	NumberOfUnits      string `json:"number_of_units"`
	PolyunsaturatedFat string `json:"polyunsaturated_fat"`
	Potassium          string `json:"potassium"`
	Protein            string `json:"protein"`
	SaturatedFat       string `json:"saturated_fat"`
	ServingID          string `json:"serving_id"`
	Sodium             string `json:"sodium"`
	Sugar              string `json:"sugar"`
	VitaminA           string `json:"vitamin_a"`
	VitaminC           string `json:"vitamin_c"`
}

type FSDiaryEntry struct {
	Entries struct {
		Meals []DiaryMeal `json:"food_entry"`
	} `json:"food_entries"`
}

type FatSecret struct{}

func NewFatSecret() *FatSecret {
	return &FatSecret{}
}

func (fs *FatSecret) makeRequestAuth(method string, requestURL string, extraParams map[string]string, tokenSecret string) ([]byte, error) {
	params := fs.mergeWithBasicParams(extraParams)

	keys := make([]string, 0, len(params))

	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	paramPairs := make([]string, 0, len(keys))

	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", fs.oauthEscape(k), fs.oauthEscape(params[k])))
	}

	normalizedParams := strings.Join(paramPairs, "&")

	signatureBaseString := fmt.Sprintf("%s&%s&%s",
		method,
		fs.oauthEscape(requestURL),
		fs.oauthEscape(normalizedParams),
	)

	signingKey := fs.buildSigningKey(tokenSecret)

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params["oauth_signature"] = signature

	form := url.Values{}

	for k, v := range params {
		form.Add(k, v)
	}

	req, err := fs.buildRequest(method, requestURL, form)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to execute HTTP request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to read response body")
	}

	return body, nil
}

func (fs *FatSecret) GetToken() (*OAuth, error) {
	params := map[string]string{
		"oauth_callback": "oob",
	}

	body, err := fs.makeRequestAuth(http.MethodGet, GetRequestTokenURL, params, "")

	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(body))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to parse query string")
	}

	oauthToken := values.Get(FSGetTokenResOAuthTokenKey)
	oauthTokenSecret := values.Get(FSGetTokenResOAuthTokenSecretKey)

	return &OAuth{
		OAuthToken:       oauthToken,
		OAuthTokenSecret: oauthTokenSecret,
	}, nil
}

func (fs *FatSecret) AuthorizeToken(oauth *OAuth) (*string, error) {
	url := fmt.Sprintf("%s?oauth_token=%s", AuthorizeTokenURL, oauth.OAuthToken)

	return &url, nil
}

func (fs *FatSecret) VerifyToken(oauth *OAuth) (*OAuth, error) {
	params := map[string]string{
		"oauth_token":    oauth.OAuthToken,
		"oauth_verifier": strconv.Itoa(oauth.OauthVerifyCode),
	}

	body, err := fs.makeRequestAuth(http.MethodGet, GetAccessTokenURL, params, oauth.OAuthTokenSecret)

	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(body))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to parse query string")
	}

	oauthToken := values.Get(FSGetTokenResOAuthTokenKey)
	oauthTokenSecret := values.Get(FSGetTokenResOAuthTokenSecretKey)

	return &OAuth{
		OAuthToken:       oauthToken,
		OAuthTokenSecret: oauthTokenSecret,
		OauthVerifyCode:  oauth.OauthVerifyCode,
	}, nil
}

func (fs *FatSecret) generateNonce() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, 11)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

func (fs *FatSecret) oauthEscape(s string) string {
	escaped := url.QueryEscape(s)

	return strings.ReplaceAll(escaped, "+", "%20")
}

func (fs *FatSecret) mergeWithBasicParams(params map[string]string) map[string]string {
	clientID := os.Getenv(constants.FatSecretClientId)

	baseParams := map[string]string{
		"oauth_consumer_key":     clientID,
		"oauth_signature_method": OAuthSignatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            fs.generateNonce(),
		"oauth_version":          OAuthVersion,
	}

	maps.Copy(baseParams, params)

	return baseParams
}

func (fs *FatSecret) buildSigningKey(tokenSecret string) string {
	consumerSecret := os.Getenv(constants.FatSecretApiKeyOauth1)

	if tokenSecret != "" {
		return fmt.Sprintf("%s&%s", fs.oauthEscape(consumerSecret), fs.oauthEscape(tokenSecret))
	}

	return fmt.Sprintf("%s&", fs.oauthEscape(consumerSecret))
}

func (fs *FatSecret) buildRequest(method string, requestURL string, form url.Values) (*http.Request, error) {
	if method == http.MethodGet {
		req, err := http.NewRequest(http.MethodGet, requestURL+"?"+form.Encode(), nil)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to create HTTP request")
		}

		return req, nil
	}

	req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to create HTTP request")
	}

	return req, nil
}
