package vendor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

func (fs *FatSecret) GetToken() (*OAuth, error) {
	clientID := os.Getenv(constants.FatSecretClientId)

	params := map[string]string{
		"oauth_consumer_key":     clientID,
		"oauth_signature_method": OAuthSignatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            generateNonce(),
		"oauth_version":          OAuthVersion,
		"oauth_callback":         "oob",
	}

	var keys []string

	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var paramPairs []string

	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", oauthEscape(k), oauthEscape(params[k])))
	}

	normalizedParams := strings.Join(paramPairs, "&")

	signatureBaseString := fmt.Sprintf("%s&%s&%s",
		"POST",
		oauthEscape(GetRequestTokenURL),
		oauthEscape(normalizedParams),
	)

	apiKey := os.Getenv(constants.FatSecretApiKeyOauth1)

	signingKey := fmt.Sprintf("%s&", apiKey)

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	form := url.Values{}

	for k, v := range params {
		form.Add(k, v)
	}

	form.Add("oauth_signature", oauthEscape(signature))

	reqURL, err := url.Parse(GetRequestTokenURL)

	if err != nil {
		// Handle error
	}

	// 3. Encode and attach the query parameters
	// This automatically handles the '?' and URL-encoding for you
	reqURL.RawQuery = form.Encode()

	// 4. Execute the GET request using the updated string URL
	resp, err := http.Get(reqURL.String())

	if err != nil {
		return nil, eris.Wrap(err, "Error doing post request to obtain FatSecret unauthorized token")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, eris.Wrap(err, "Error reading body stream of FatSecret response")
	}

	var res FSTokenRes

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, eris.Wrap(err, "Error unmarshaling FatSecret response body")
	}

	return &OAuth{
		OAuthToken:       res.OAuthToken,
		OAuthTokenSecret: res.OAuthTokenSecret,
	}, nil
}

func (fs *FatSecret) AuthorizeToken(oauth *OAuth) (*string, error) {
	clientID := os.Getenv(constants.FatSecretClientId)

	params := map[string]string{
		"oauth_consumer_key":     clientID,
		"oauth_signature_method": OAuthSignatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            generateNonce(),
		"oauth_version":          OAuthVersion,
		"oauth_token":            oauth.OAuthToken,
	}

	var keys []string

	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var paramPairs []string

	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", oauthEscape(k), oauthEscape(params[k])))
	}

	normalizedParams := strings.Join(paramPairs, "&")

	signatureBaseString := fmt.Sprintf("%s&%s&%s",
		"GET",
		oauthEscape(GetRequestTokenURL),
		oauthEscape(normalizedParams),
	)

	apiKey := os.Getenv(constants.FatSecretApiKeyOauth1)

	signingKey := fmt.Sprintf("%s&%s&", oauthEscape(apiKey), oauthEscape(oauth.OAuthTokenSecret))

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params["oauth_signature"] = signature

	form := url.Values{}

	for k, v := range params {
		form.Add(k, v)
	}

	resp, err := http.PostForm(AuthorizeTokenURL, form)

	if err != nil {
		return nil, eris.Wrap(err, "Error doing Authorize post request")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, eris.Wrap(err, "Error reading body stream of FatSecret response")
	}

	var res FSTokenRes

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, eris.Wrap(err, "Error unmarshaling FatSecret response body")
	}

	return &signature, nil
}

func (fs *FatSecret) VerifyToken(oauth *OAuth) (*OAuth, error) {
	clientID := os.Getenv(constants.FatSecretClientId)

	params := map[string]string{
		"oauth_consumer_key":     clientID,
		"oauth_signature_method": OAuthSignatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            generateNonce(),
		"oauth_version":          OAuthVersion,
		"oauth_token":            oauth.OAuthToken,
		"oauth_verifier":         oauth.OauthVerifyCode,
	}

	var keys []string

	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var paramPairs []string

	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", oauthEscape(k), oauthEscape(params[k])))
	}

	normalizedParams := strings.Join(paramPairs, "&")

	signatureBaseString := fmt.Sprintf("%s&%s&%s",
		"GET",
		oauthEscape(GetRequestTokenURL),
		oauthEscape(normalizedParams),
	)

	apiKey := os.Getenv(constants.FatSecretApiKeyOauth1)

	signingKey := fmt.Sprintf("%s&%s&", oauthEscape(apiKey), oauthEscape(oauth.OAuthTokenSecret))

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params["oauth_signature"] = oauthEscape(signature)

	form := url.Values{}

	for k, v := range params {
		form.Add(k, v)
	}

	resp, err := http.PostForm(GetAccessTokenURL, form)

	if err != nil {
		return nil, eris.Wrap(err, "Error doing Verify post request")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, eris.Wrap(err, "Error reading body stream of FatSecret response")
	}

	var res FSTokenRes

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, eris.Wrap(err, "Error unmarshaling FatSecret response body")
	}

	return nil, nil
}

func generateNonce() string {
	b := make([]byte, 16)

	_, err := rand.Read(b)

	if err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(b)
}

func oauthEscape(s string) string {
	return strings.NewReplacer("+", "%20", "*", "%2A", "%7E", "~").Replace(url.QueryEscape(s))
}
