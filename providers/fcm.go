package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

const (
	fcmBaseURL = "https://fcm.googleapis.com"
	fcmScope   = "https://www.googleapis.com/auth/firebase.messaging"
)

// FCMProvider sends pushes using the Firebase Cloud Messaging HTTP v1 API.
type FCMProvider struct {
	*BaseProvider
	projectID   string
	tokenSource oauth2.TokenSource
}

func NewFCMProvider(config map[string]interface{}) Provider {
	base := ""
	if p, ok := config["provider"]; ok {
		if s, ok := p.(string); ok {
			base = s
		}
	}
	if base == "" {
		base = fcmBaseURL
	}

	tokenSource, projectID, err := newFCMTokenSource(config)
	if err != nil || tokenSource == nil {
		tokenSource = nil
	}
	if projectID == "" {
		projectID = argString(config, "project_id", "projectId", "project")
	}

	return &FCMProvider{
		BaseProvider: NewBaseProvider(base),
		projectID:    projectID,
		tokenSource:  tokenSource,
	}
}

func (f *FCMProvider) SendPush(token string, payload map[string]interface{}) (*PushResponse, error) {
	ctx := context.Background()
	if f.tokenSource == nil {
		return nil, errors.New("fcm: access token source is not configured")
	}
	if f.projectID == "" {
		return nil, errors.New("fcm: project id is not configured")
	}

	tok, err := f.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("fcm: failed to obtain access token: %w", err)
	}

	body, err := json.Marshal(buildFCMV1Message(token, payload))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.BaseURL+"/v1/projects/"+f.projectID+"/messages:send", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}

	var sendResp fcmSendResponse
	if err := json.Unmarshal(respBody, &sendResp); err == nil && sendResp.MessageID != "" {
		return &PushResponse{Status: "delivered", Reason: "ok"}, nil
	}

	var fcmErr fcmErrorResponse
	if err := json.Unmarshal(respBody, &fcmErr); err != nil {
		fcmErr = fcmErrorResponse{}
	}
	reason := http.StatusText(resp.StatusCode)
	if fcmErr.Error.Message != "" {
		reason = fcmErr.Error.Message
	}
	return &PushResponse{Status: "failed", Reason: reason}, nil
}

func newFCMTokenSource(config map[string]interface{}) (oauth2.TokenSource, string, error) {
	if tokStr := argString(config, "access_token", "accessToken", "token"); tokStr != "" {
		src := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: tokStr,
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		})
		return src, argString(config, "project_id", "projectId", "project"), nil
	}

	saStr := argString(config, "service_account", "serviceAccount", "credentials", "credential")
	if saStr == "" {
		return nil, "", errors.New("fcm: either service_account or access_token must be provided")
	}

	var saJSON []byte
	if !strings.HasPrefix(saStr, "{") {
		data, err := os.ReadFile(saStr)
		if err != nil {
			return nil, "", err
		}
		saJSON = data
	} else {
		saJSON = []byte(saStr)
	}

	var saInfo struct {
		Type         string `json:"type"`
		ProjectID    string `json:"project_id"`
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
		ClientEmail  string `json:"client_email"`
		TokenURI     string `json:"token_uri"`
	}
	if err := json.Unmarshal(saJSON, &saInfo); err != nil {
		return nil, "", err
	}

	tokenURL := saInfo.TokenURI
	if tokenURL == "" {
		tokenURL = "https://oauth2.googleapis.com/token"
	}
	cfg := &jwt.Config{
		Email:        saInfo.ClientEmail,
		PrivateKey:   []byte(saInfo.PrivateKey),
		PrivateKeyID: saInfo.PrivateKeyID,
		Scopes:       []string{fcmScope},
		TokenURL:     tokenURL,
		Expires:      time.Hour,
	}
	src := cfg.TokenSource(context.Background())
	return src, saInfo.ProjectID, nil
}

func argString(args map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := args[k]; ok {
			switch val := v.(type) {
			case string:
				if val != "" {
					return val
				}
			case float64:
				if val != float64(0) {
					return strconv.FormatFloat(val, 'f', -1, 64)
				}
			}
		}
	}
	return ""
}

type fcmV1Message struct {
	Message fcmV1MessageBody `json:"message"`
}

type fcmV1MessageBody struct {
	Token        string             `json:"token,omitempty"`
	Notification *fcmV1Notification `json:"notification,omitempty"`
	Data         map[string]string  `json:"data,omitempty"`
}

type fcmV1Notification struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type fcmSendResponse struct {
	Name      string `json:"name"`
	MessageID string `json:"message_id"`
}

type fcmErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func buildFCMV1Message(token string, payload map[string]interface{}) *fcmV1Message {
	msg := &fcmV1Message{Message: fcmV1MessageBody{Token: token}}

	title := argString(payload, "title")
	body := argString(payload, "body")
	if notif, ok := payload["notification"].(map[string]interface{}); ok {
		if t, ok := notif["title"].(string); ok && t != "" {
			title = t
		}
		if b, ok := notif["body"].(string); ok && b != "" {
			body = b
		}
	}

	if title != "" || body != "" {
		msg.Message.Notification = &fcmV1Notification{Title: title, Body: body}
	}

	data := map[string]string{}
	for k, v := range payload {
		if k == "notification" || k == "title" || k == "body" {
			continue
		}
		switch val := v.(type) {
		case string:
			data[k] = val
		case float64:
			data[k] = strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			data[k] = strconv.FormatBool(val)
		}
	}
	if len(data) > 0 {
		msg.Message.Data = data
	}

	return msg
}
