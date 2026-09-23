package telegram

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildSendMessageBody(t *testing.T) {
	body, err := buildSendMessageBody(
		"-123456",
		"backup progress",
	)

	if err != nil {
		t.Fatalf(
			"buildSendMessageBody() unexpected error = %v",
			err,
		)
	}

	var got SendMessageRequest

	err = json.Unmarshal(body, &got)
	if err != nil {
		t.Fatalf(
			"failed to unmarshal body: %v",
			err,
		)
	}

	want := SendMessageRequest{
		ChatID: "-123456",
		Text:   "backup progress",
	}

	if got != want {
		t.Errorf(
			"buildSendMessageBody() = %+v, want %+v",
			got,
			want,
		)
	}
}

func TestBuildSendMessageBodyEscapesText(t *testing.T) {
	text := "backup \"alice\"\nTransferred: 50%"

	body, err := buildSendMessageBody(
		"-123456",
		text,
	)

	if err != nil {
		t.Fatalf(
			"buildSendMessageBody() unexpected error = %v",
			err,
		)
	}

	var got SendMessageRequest

	err = json.Unmarshal(body, &got)
	if err != nil {
		t.Fatalf(
			"failed to unmarshal body: %v",
			err,
		)
	}

	if got.Text != text {
		t.Errorf(
			"Text = %q, want %q",
			got.Text,
			text,
		)
	}
}

func TestBuildSendMessageRequest(t *testing.T) {
	req, err := buildSendMessageRequest(
		"test-token",
		"-123456",
		"backup progress",
	)

	if err != nil {
		t.Fatalf(
			"buildSendMessageRequest() unexpected error = %v",
			err,
		)
	}

	if req.Method != http.MethodPost {
		t.Errorf(
			"Method = %q, want %q",
			req.Method,
			http.MethodPost,
		)
	}

	wantURL := "https://api.telegram.org/bottest-token/sendMessage"

	if req.URL.String() != wantURL {
		t.Errorf(
			"URL = %q, want %q",
			req.URL.String(),
			wantURL,
		)
	}

	contentType := req.Header.Get("Content-Type")

	if contentType != "application/json" {
		t.Errorf(
			"Content-Type = %q, want %q",
			contentType,
			"application/json",
		)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf(
			"failed to read request body: %v",
			err,
		)
	}

	var got SendMessageRequest

	err = json.Unmarshal(body, &got)
	if err != nil {
		t.Fatalf(
			"failed to unmarshal request body: %v",
			err,
		)
	}

	want := SendMessageRequest{
		ChatID: "-123456",
		Text:   "backup progress",
	}

	if got != want {
		t.Errorf(
			"request body = %+v, want %+v",
			got,
			want,
		)
	}
}

func TestBuildSendMessageRequestReturnsErrorForInvalidURL(t *testing.T) {
	_, err := buildSendMessageRequest(
		"invalid\ntoken",
		"-123456",
		"backup progress",
	)

	if err == nil {
		t.Fatal(
			"buildSendMessageRequest() error = nil, want error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"failed create new request to telegram",
	) {
		t.Errorf(
			"error = %q, want request creation error",
			err,
		)
	}
}

func TestDoRequestSuccess(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)

			_, err := w.Write(
				[]byte(`{"ok":true}`),
			)
			if err != nil {
				t.Fatal(err)
			}
		}),
	)
	defer server.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		server.URL,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	body, err := doRequest(
		server.Client(),
		req,
	)

	if err != nil {
		t.Fatalf(
			"doRequest() unexpected error = %v",
			err,
		)
	}

	want := `{"ok":true}`

	if string(body) != want {
		t.Errorf(
			"doRequest() body = %q, want %q",
			body,
			want,
		)
	}
}

func TestDoRequestReturnsErrorForBadStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusBadRequest)

			_, err := w.Write(
				[]byte(`{"ok":false,"description":"bad request"}`),
			)
			if err != nil {
				t.Fatal(err)
			}
		}),
	)
	defer server.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		server.URL,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	body, err := doRequest(
		server.Client(),
		req,
	)

	if err == nil {
		t.Fatal(
			"doRequest() error = nil, want error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"400 Bad Request",
	) {
		t.Errorf(
			"doRequest() error = %q, want 400 Bad Request",
			err,
		)
	}

	wantBody := `{"ok":false,"description":"bad request"}`

	if string(body) != wantBody {
		t.Errorf(
			"doRequest() body = %q, want %q",
			body,
			wantBody,
		)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	return f(req)
}

func TestSendMessageSuccess(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(
			req *http.Request,
		) (*http.Response, error) {

			if req.Method != http.MethodPost {
				t.Errorf(
					"Method = %q, want %q",
					req.Method,
					http.MethodPost,
				)
			}

			wantURL :=
				"https://api.telegram.org/bottest-token/sendMessage"

			if req.URL.String() != wantURL {
				t.Errorf(
					"URL = %q, want %q",
					req.URL.String(),
					wantURL,
				)
			}

			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf(
					"Content-Type = %q, want application/json",
					req.Header.Get("Content-Type"),
				)
			}

			requestBody, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}

			var request SendMessageRequest

			err = json.Unmarshal(requestBody, &request)
			if err != nil {
				t.Fatal(err)
			}

			wantRequest := SendMessageRequest{
				ChatID: "-123456",
				Text:   "backup progress",
			}

			if request != wantRequest {
				t.Errorf(
					"request = %+v, want %+v",
					request,
					wantRequest,
				)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body: io.NopCloser(
					strings.NewReader(
						`{"ok":true,"result":{"message_id":42}}`,
					),
				),
			}, nil
		}),
	}

	messageID, err := SendMessage(
		client,
		"test-token",
		"-123456",
		"backup progress",
	)

	if err != nil {
		t.Fatalf(
			"SendMessage() unexpected error = %v",
			err,
		)
	}

	want := 42

	if messageID != want {
		t.Errorf(
			"SendMessage() = %d, want %d",
			messageID,
			want,
		)
	}
}

func TestSendMessageReturnsHTTPError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(
			req *http.Request,
		) (*http.Response, error) {

			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "400 Bad Request",
				Header:     make(http.Header),
				Body: io.NopCloser(
					strings.NewReader(
						`{"ok":false,"description":"chat not found"}`,
					),
				),
			}, nil
		}),
	}

	messageID, err := SendMessage(
		client,
		"test-token",
		"-123456",
		"backup progress",
	)

	if err == nil {
		t.Fatal("SendMessage() error = nil, want error")
	}

	if messageID != 0 {
		t.Errorf(
			"SendMessage() = %d, want 0",
			messageID,
		)
	}

	if !strings.Contains(
		err.Error(),
		"failed to send telegram message",
	) {
		t.Errorf(
			"SendMessage() error = %q, want send message error",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"400 Bad Request",
	) {
		t.Errorf(
			"SendMessage() error = %q, want 400 Bad Request",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"chat not found",
	) {
		t.Errorf(
			"SendMessage() error = %q, want telegram response body",
			err,
		)
	}
}

func TestSendMessageReturnsNetworkError(t *testing.T) {
	networkErr := errors.New("network unavailable")

	client := &http.Client{
		Transport: roundTripFunc(func(
			req *http.Request,
		) (*http.Response, error) {
			return nil, networkErr
		}),
	}

	messageID, err := SendMessage(
		client,
		"test-token",
		"-123456",
		"backup progress",
	)

	if err == nil {
		t.Fatal("SendMessage() error = nil, want error")
	}

	if messageID != 0 {
		t.Errorf(
			"SendMessage() = %d, want 0",
			messageID,
		)
	}

	if !errors.Is(err, networkErr) {
		t.Errorf(
			"SendMessage() error does not wrap network error: %v",
			err,
		)
	}
}

func TestSendMessageReturnsBuildRequestError(t *testing.T) {
	client := &http.Client{}

	messageID, err := SendMessage(
		client,
		"invalid\ntoken",
		"-123456",
		"backup progress",
	)

	if err == nil {
		t.Fatal("SendMessage() error = nil, want error")
	}

	if messageID != 0 {
		t.Errorf(
			"SendMessage() = %d, want 0",
			messageID,
		)
	}

	if !strings.Contains(
		err.Error(),
		"failed to build send message request",
	) {
		t.Errorf(
			"SendMessage() error = %q, want build request error",
			err,
		)
	}
}

func TestSendMessageReturnsParseResponseError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(
			req *http.Request,
		) (*http.Response, error) {

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body: io.NopCloser(
					strings.NewReader(
						`{"ok":true,"result":{}}`,
					),
				),
			}, nil
		}),
	}

	messageID, err := SendMessage(
		client,
		"test-token",
		"-123456",
		"backup progress",
	)

	if err == nil {
		t.Fatal("SendMessage() error = nil, want error")
	}

	if messageID != 0 {
		t.Errorf(
			"SendMessage() = %d, want 0",
			messageID,
		)
	}

	if !strings.Contains(
		err.Error(),
		"failed to parse telegram send message response",
	) {
		t.Errorf(
			"SendMessage() error = %q, want parse response error",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"does not contain message_id",
	) {
		t.Errorf(
			"SendMessage() error = %q, want missing message_id error",
			err,
		)
	}
}

func TestParseMessageID(t *testing.T) {
	body := []byte(
		`{"ok":true,"result":{"message_id":42}}`,
	)

	got, err := parseMessageID(body)

	if err != nil {
		t.Fatalf(
			"parseMessageID() unexpected error = %v",
			err,
		)
	}

	want := 42

	if got != want {
		t.Errorf(
			"parseMessageID() = %d, want %d",
			got,
			want,
		)
	}
}

func TestParseMessageIDReturnsErrorForInvalidJSON(t *testing.T) {
	body := []byte(`{"ok":`)

	messageID, err := parseMessageID(body)

	if err == nil {
		t.Fatal(
			"parseMessageID() error = nil, want error",
		)
	}

	if messageID != 0 {
		t.Errorf(
			"parseMessageID() = %d, want 0",
			messageID,
		)
	}

	var syntaxErr *json.SyntaxError

	if !errors.As(err, &syntaxErr) {
		t.Errorf(
			"parseMessageID() error does not wrap *json.SyntaxError: %v",
			err,
		)
	}
}

func TestParseMessageIDReturnsErrorWhenResponseNotOK(t *testing.T) {
	body := []byte(
		`{"ok":false}`,
	)

	messageID, err := parseMessageID(body)

	if err == nil {
		t.Fatal(
			"parseMessageID() error = nil, want error",
		)
	}

	if messageID != 0 {
		t.Errorf(
			"parseMessageID() = %d, want 0",
			messageID,
		)
	}

	if !strings.Contains(
		err.Error(),
		"telegram response is not ok",
	) {
		t.Errorf(
			"parseMessageID() error = %q, want not ok error",
			err,
		)
	}
}

func TestParseMessageIDReturnsErrorWhenResultMissing(t *testing.T) {
	body := []byte(
		`{"ok":true}`,
	)

	messageID, err := parseMessageID(body)

	if err == nil {
		t.Fatal(
			"parseMessageID() error = nil, want error",
		)
	}

	if messageID != 0 {
		t.Errorf(
			"parseMessageID() = %d, want 0",
			messageID,
		)
	}

	if !strings.Contains(
		err.Error(),
		"does not contain result",
	) {
		t.Errorf(
			"parseMessageID() error = %q, want missing result error",
			err,
		)
	}
}

func TestParseMessageIDReturnsErrorWhenMessageIDMissing(t *testing.T) {
	body := []byte(
		`{"ok":true,"result":{}}`,
	)

	messageID, err := parseMessageID(body)

	if err == nil {
		t.Fatal(
			"parseMessageID() error = nil, want error",
		)
	}

	if messageID != 0 {
		t.Errorf(
			"parseMessageID() = %d, want 0",
			messageID,
		)
	}

	if !strings.Contains(
		err.Error(),
		"does not contain message_id",
	) {
		t.Errorf(
			"parseMessageID() error = %q, want missing message_id error",
			err,
		)
	}
}

func TestBuildEditMessageTextBody(t *testing.T) {
	body, err := buildEditMessageTextBody("-123456", "Скопировано 25%", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got EditMessageTextRequest

	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	want := EditMessageTextRequest{
		ChatID:    "-123456",
		MessageID: 42,
		Text:      "Скопировано 25%",
	}

	if got != want {
		t.Fatalf("unexpected body: got %+v, want %+v", got, want)
	}
}

func TestBuildEditMessageTextBodyWithSpecialCharacters(t *testing.T) {
	text := "backup \"photos\"\nTransferred: 25%"

	body, err := buildEditMessageTextBody("-123456", text, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got EditMessageTextRequest

	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if got.Text != text {
		t.Fatalf("unexpected text: got %q, want %q", got.Text, text)
	}
}

func TestBuildEditMessageTextRequest(t *testing.T) {
	req, err := buildEditMessageTextRequest(
		"test-token",
		"-123456",
		"Скопировано 25%",
		42,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Method != http.MethodPost {
		t.Fatalf("unexpected method: got %q, want %q", req.Method, http.MethodPost)
	}

	wantURL := "https://api.telegram.org/bottest-token/editMessageText"

	if req.URL.String() != wantURL {
		t.Fatalf(
			"unexpected URL: got %q, want %q",
			req.URL.String(),
			wantURL,
		)
	}

	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"unexpected Content-Type: got %q, want %q",
			got,
			"application/json",
		)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	var got EditMessageTextRequest

	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	want := EditMessageTextRequest{
		ChatID:    "-123456",
		MessageID: 42,
		Text:      "Скопировано 25%",
	}

	if got != want {
		t.Fatalf("unexpected request body: got %+v, want %+v", got, want)
	}
}

func TestBuildEditMessageTextRequestReturnsError(t *testing.T) {
	_, err := buildEditMessageTextRequest(
		"invalid\ntoken",
		"-123456",
		"text",
		42,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed create new edit message request to telegram") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEditMessageTextSuccess(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Errorf(
					"unexpected method: got %q, want %q",
					req.Method,
					http.MethodPost,
				)
			}

			wantURL := "https://api.telegram.org/bottest-token/editMessageText"

			if req.URL.String() != wantURL {
				t.Errorf(
					"unexpected URL: got %q, want %q",
					req.URL.String(),
					wantURL,
				)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			var got EditMessageTextRequest

			if err := json.Unmarshal(body, &got); err != nil {
				return nil, err
			}

			want := EditMessageTextRequest{
				ChatID:    "-123456",
				MessageID: 42,
				Text:      "Скопировано 25%",
			}

			if got != want {
				t.Errorf(
					"unexpected request body: got %+v, want %+v",
					got,
					want,
				)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(
					`{"ok":true,"result":{"message_id":42}}`,
				)),
				Header: make(http.Header),
			}, nil
		}),
	}

	err := EditMessageText(
		client,
		"test-token",
		"-123456",
		"Скопировано 25%",
		42,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEditMessageTextReturnsHTTPError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "400 Bad Request",
				Body: io.NopCloser(strings.NewReader(
					`{"ok":false,"description":"Bad Request: message to edit not found"}`,
				)),
				Header: make(http.Header),
			}, nil
		}),
	}

	err := EditMessageText(
		client,
		"test-token",
		"-123456",
		"text",
		42,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to edit telegram message") {
		t.Fatalf("error does not contain EditMessageText context: %v", err)
	}

	if !strings.Contains(err.Error(), "400 Bad Request") {
		t.Fatalf("error does not contain HTTP status: %v", err)
	}

	if !strings.Contains(err.Error(), "message to edit not found") {
		t.Fatalf("error does not contain telegram response body: %v", err)
	}
}

func TestEditMessageTextReturnsNetworkError(t *testing.T) {
	networkErr := errors.New("network unavailable")

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, networkErr
		}),
	}

	err := EditMessageText(
		client,
		"test-token",
		"-123456",
		"text",
		42,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, networkErr) {
		t.Fatalf(
			"expected wrapped network error, got: %v",
			err,
		)
	}

	if !strings.Contains(err.Error(), "failed to edit telegram message") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEditMessageTextReturnsBuildRequestError(t *testing.T) {
	client := &http.Client{}

	err := EditMessageText(
		client,
		"invalid\ntoken",
		"-123456",
		"text",
		42,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to build edit message request") {
		t.Fatalf("unexpected error: %v", err)
	}
}
