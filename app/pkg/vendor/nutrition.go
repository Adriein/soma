package vendor

import (
	"maps"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	GetRequestTokenURL   = "https://authentication.fatsecret.com/oauth/request_token"
	AuthorizeTokenURL    = "https://authentication.fatsecret.com/oauth/authorize"
	GetAccessTokenURL    = "https://authentication.fatsecret.com/oauth/access_token"
	OAuthSignatureMethod = "HMAC-SHA1"
	OAuthVersion         = "1.0"
)

type NutritionDiary interface {
	GetToken() (*OAuth, error)
	AuthorizeToken(oauth *OAuth) (*string, error)
	VerifyToken(oauth *OAuth) (*OAuth, error)
}

type OAuth struct {
	OAuthToken       string
	OAuthTokenSecret string
	OauthVerifyCode  string
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

type FSTokenRes struct {
	OAuthToken       string `json:"oauth_token"`
	OAuthTokenSecret string `json:"oauth_token_secret"`
}

type FatSecret struct{}

func NewFatSecret() *FatSecret {
	return &FatSecret{}
}

func (fs *FatSecret) makeRequest(method string, requestURL string, params map[string]string, tokenSecret string) ([]byte, error) {
	clientID := os.Getenv(constants.FatSecretClientId)

	baseParams := map[string]string{
		"oauth_consumer_key":     clientID,
		"oauth_signature_method": OAuthSignatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            generateNonce(),
		"oauth_version":          OAuthVersion,
	}

	maps.Copy(baseParams, params)

	keys := make([]string, 0, len(baseParams))

	for k := range baseParams {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	paramPairs := make([]string, 0, len(keys))

	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", oauthEscape(k), oauthEscape(baseParams[k])))
	}

	normalizedParams := strings.Join(paramPairs, "&")

	signatureBaseString := fmt.Sprintf("%s&%s&%s",
		method,
		oauthEscape(requestURL),
		oauthEscape(normalizedParams),
	)

	consumerSecret := os.Getenv(constants.FatSecretApiKeyOauth1)

	var signingKey string
	if tokenSecret != "" {
		signingKey = fmt.Sprintf("%s&%s&", oauthEscape(consumerSecret), oauthEscape(tokenSecret))
	} else {
		signingKey = fmt.Sprintf("%s&", oauthEscape(consumerSecret))
	}

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	baseParams["oauth_signature"] = signature

	form := url.Values{}

	for k, v := range baseParams {
		form.Add(k, v)
	}

	var req *http.Request
	var err error

	if method == "GET" {
		req, err = http.NewRequest("GET", requestURL+"?"+form.Encode(), nil)
	} else {
		req, err = http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
		if err == nil {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	if err != nil {
		return nil, eris.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Add("Accept", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, eris.Wrap(err, "failed to execute HTTP request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, eris.Wrap(err, "failed to read response body")
	}

	return body, nil
}

func (fs *FatSecret) GetToken() (*OAuth, error) {
	params := map[string]string{
		"oauth_callback": "oob",
	}

	body, err := fs.makeRequest("GET", GetRequestTokenURL, params, "")
	if err != nil {
		return nil, err
	}

	var res FSTokenRes
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, eris.Wrap(err, "error unmarshaling FatSecret response body")
	}

	return &OAuth{
		OAuthToken:       res.OAuthToken,
		OAuthTokenSecret: res.OAuthTokenSecret,
	}, nil
}

func (fs *FatSecret) AuthorizeToken(oauth *OAuth) (*string, error) {
	params := map[string]string{
		"oauth_token": oauth.OAuthToken,
	}

	body, err := fs.makeRequest("POST", AuthorizeTokenURL, params, oauth.OAuthTokenSecret)
	if err != nil {
		return nil, err
	}

	var res FSTokenRes
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, eris.Wrap(err, "error unmarshaling FatSecret response body")
	}

	return &res.OAuthToken, nil
}

func (fs *FatSecret) VerifyToken(oauth *OAuth) (*OAuth, error) {
	params := map[string]string{
		"oauth_token":    oauth.OAuthToken,
		"oauth_verifier": oauth.OauthVerifyCode,
	}

	body, err := fs.makeRequest("POST", GetAccessTokenURL, params, oauth.OAuthTokenSecret)
	if err != nil {
		return nil, err
	}

	var res FSTokenRes
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, eris.Wrap(err, "error unmarshaling FatSecret response body")
	}

	return nil, nil
}

func generateNonce() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, 11)
	for i := range result {
		// Generate a random index between 0 and len(charset)-1
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

func oauthEscape(s string) string {
	escaped := url.QueryEscape(s)

	return strings.ReplaceAll(escaped, "+", "%20")
}
