package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func StartJob(url string, body map[string]string) ([]byte, error) {
	postBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
