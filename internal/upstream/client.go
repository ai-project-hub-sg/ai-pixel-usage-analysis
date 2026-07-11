package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ErrorKind string

const (
	ErrorAuth      ErrorKind = "auth"
	ErrorRateLimit ErrorKind = "rate_limit"
	ErrorClient    ErrorKind = "client"
	ErrorServer    ErrorKind = "server"
	ErrorTransport ErrorKind = "transport"
)

type Error struct {
	Kind   ErrorKind
	Status int
	Err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("upstream %s error (status %d): %v", e.Kind, e.Status, e.Err)
}
func (e *Error) Unwrap() error { return e.Err }
func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	return errors.As(err, &target) && target.Kind == kind
}

type Client struct {
	baseURL, email, password  string
	http                      *http.Client
	accessToken, refreshToken string
}

func NewClient(baseURL, email, password string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/") + "/api/v1", email: email, password: password, http: client}
}

func (c *Client) Login(ctx context.Context) error {
	var settings struct {
		LoginAgreementRevision string `json:"login_agreement_revision"`
	}
	if err := c.get(ctx, "/settings/public", nil, &settings); err != nil {
		return err
	}
	body := map[string]any{"email": c.email, "password": c.password, "login_agreement_revision": settings.LoginAgreementRevision}
	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.post(ctx, "/auth/login", body, &tokens); err != nil {
		return err
	}
	c.accessToken, c.refreshToken = tokens.AccessToken, tokens.RefreshToken
	return nil
}

func (c *Client) Refresh(ctx context.Context) error {
	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.post(ctx, "/auth/refresh", map[string]string{"refresh_token": c.refreshToken}, &tokens); err != nil {
		return err
	}
	c.accessToken, c.refreshToken = tokens.AccessToken, tokens.RefreshToken
	return nil
}

func (c *Client) ListUsage(ctx context.Context, q UsageQuery) (Page[UsageRecord], error) {
	var page Page[UsageRecord]
	err := c.get(ctx, "/usage", commonQuery(q.Page, q.PageSize, q.StartTime, q.EndTime, q.Timezone, q.SortOrder), &page)
	return page, err
}

func (c *Client) ListLedger(ctx context.Context, q LedgerQuery) (Page[LedgerEntry], error) {
	values := commonQuery(q.Page, q.PageSize, q.StartTime, q.EndTime, q.Timezone, q.SortOrder)
	for key, value := range map[string]string{"direction": q.Direction, "reason": q.Reason, "ref_type": q.RefType} {
		if value != "" {
			values.Set(key, value)
		}
	}
	if q.RefID > 0 {
		values.Set("ref_id", strconv.FormatInt(q.RefID, 10))
	}
	var page Page[LedgerEntry]
	err := c.get(ctx, "/usage/balance-ledger", values, &page)
	return page, err
}

func commonQuery(page, pageSize int, start, end time.Time, timezone, sortOrder string) url.Values {
	v := url.Values{"page": {strconv.Itoa(page)}, "page_size": {strconv.Itoa(pageSize)}, "timezone": {timezone}, "sort_order": {sortOrder}}
	if !start.IsZero() {
		v.Set("start_time", start.Format(time.RFC3339))
		v.Set("start_date", start.Format("2006-01-02"))
	}
	if !end.IsZero() {
		v.Set("end_time", end.Format(time.RFC3339))
		v.Set("end_date", end.Format("2006-01-02"))
	}
	return v
}

func (c *Client) get(ctx context.Context, path string, query url.Values, target any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	return c.do(req, target)
}
func (c *Client) post(ctx context.Context, path string, body any, target any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, target)
}
func (c *Client) do(req *http.Request, target any) error {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return &Error{Kind: ErrorTransport, Err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		kind := ErrorClient
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			kind = ErrorAuth
		} else if resp.StatusCode == 429 {
			kind = ErrorRateLimit
		} else if resp.StatusCode >= 500 {
			kind = ErrorServer
		}
		return &Error{Kind: kind, Status: resp.StatusCode, Err: errors.New(http.StatusText(resp.StatusCode))}
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 16<<20))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode upstream response: %w", err)
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		return fmt.Errorf("decode upstream data: %w", err)
	}
	return nil
}
