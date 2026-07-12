package vendor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/adriein/soma/app/pkg/constants"
	"github.com/rotisserie/eris"
)

const GemniniURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.5-flash:generateContent"

type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Candidate struct {
	Content Content
}

type AIRes struct {
	Candidates []Candidate `json:"candidates"`
}

func (r *AIRes) Text() string {
	return r.Candidates[0].Content.Parts[0].Text
}

type AI interface {
	Ask(question string) (*AIRes, error)
}

type Gemini struct{}

func NewGemini() *Gemini {
	return &Gemini{}
}

func (g *Gemini) Ask(question string) (*AIRes, error) {
	apiKey := os.Getenv(constants.GeminiApiKey)

	body := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: question},
				},
			},
		},
	}

	jsonData, err := json.Marshal(body)

	if err != nil {
		return nil, eris.Wrap(err, "Gemini ask, error marshaling json")
	}

	req, err := http.NewRequest("POST", GemniniURL, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, eris.Wrap(err, "Gemini ask, error creating the request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, eris.Wrap(err, "Gemini ask, error doing the request")
	}

	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, eris.Wrap(err, "Gemini ask, error reading response")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, eris.New(fmt.Sprintf("Gemini response error, %s with http code %d", string(resBody), resp.StatusCode))
	}

	var geminiResp AIRes

	if err := json.Unmarshal(resBody, &geminiResp); err != nil {
		return nil, eris.Wrap(err, "Gemini ask, error unmarshalling the response")
	}

	return &geminiResp, nil
}
