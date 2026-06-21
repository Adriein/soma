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
	Authorize(oauth *OAuth) error
	GetDiaryEntry() ([]*DiaryMeal, error)
}

type OAuth struct {
	OAuthToken       string
	OAuthTokenSecret string
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

type FSUnauthorizedTokenRes struct {
	OAuthToken       string `json:"oauth_token"`
	OAuthTokenSecret string `json:"oauth_token_secret"`
}

type FatSecret struct{}

func NewFatSecret() *FatSecret {
	return &FatSecret{}
}

func (fs *FatSecret) GetUnauthorizedToken() error {
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

	apiKey := os.Getenv(constants.FatSecretApiKey)

	signingKey := fmt.Sprintf("%s&", oauthEscape(apiKey))

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params["oauth_signature"] = signature

	form := url.Values{}

	for k, v := range params {
		form.Add(k, v)
	}

	resp, err := http.PostForm(GetRequestTokenURL, form)

	if err != nil {
		return eris.Wrap(err, "FatSecret get unauthorized token error making post request")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return eris.Wrap(err, "FatSecret get unauthorized token error reading body stream")
	}

	var res FSUnauthorizedTokenRes

	if err := json.Unmarshal(body, &res); err != nil {
		return eris.Wrap(err, "FatSecret get unauthorized token error parsing body")
	}

	fmt.Printf("HTTP Status: %s\n", resp.Status)

	return nil
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
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}
