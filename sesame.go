package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const DefaultEndpoint = "https://app.candyhouse.co/api/sesame2"

type Client struct {
	// API endpoint. DefaultEndpoint will be used when this value is empty.
	Endpoint string
	// API key you can get via https://dash.candyhouse.co
	APIKey string
}

type StatusResponse struct {
	BatteryPercentage int       `json:"batteryPercentage"`
	BatteryVoltage    float64   `json:"batteryVoltage"`
	Position          int       `json:"position"`
	Status            State     `json:"CHSesame2Status"`
	Timestamp         time.Time `json:"timestamp"`
}

type State string

const (
	Locked   State = "locked"
	Unlocked State = "unlocked"
	Moved    State = "moved"
)

func (state State) String() string { return string(state) }

// Status API
// https://doc.candyhouse.co/ja/SesameAPI#sesame%E3%81%AE%E7%8A%B6%E6%85%8B%E3%82%92%E5%8F%96%E5%BE%97
// The server responses internal server error when uuid is not found or invalid. UUID string must be upper case.
func (c *Client) Status(ctx context.Context, uuid string) (*StatusResponse, error) {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/"+uuid, nil)
	if err != nil {
		return nil, fmt.Errorf("creating new HTTP request: %w", err)
	}

	req.Header.Add("x-api-key", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing HTTP request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	return &status, nil
}

type HistoryResponse struct {
	Pages []HistoryPage
}

type HistoryPage struct {
	RecordID   int         `json:"recordID"`
	Type       HistoryType `json:"type"`
	HistoryTag string      `json:"historyTag"`
	DevicePK   string      `json:"devicePk"`
	Timestamp  time.Time   `json:"timestamp"`
}

//go:generate stringer -type=HistoryType
type HistoryType int

// defined at https://doc.candyhouse.co/ja/SesameAPI#sesame%E3%81%AE%E5%B1%A5%E6%AD%B4%E3%82%92%E5%8F%96%E5%BE%97
const (
	None                   HistoryType = 0
	BLELock                HistoryType = 1
	BLEUnlock              HistoryType = 2
	TimeChanged            HistoryType = 3
	AutoLockUpdated        HistoryType = 4
	MechSettingUpdated     HistoryType = 5
	AutoLock               HistoryType = 6
	ManualLocked           HistoryType = 7
	ManualUnlocked         HistoryType = 8
	ManualElse             HistoryType = 9
	DriveLocked            HistoryType = 10
	DriveUnlocked          HistoryType = 11
	DriveFailed            HistoryType = 12
	BLEAdvParameterUpdated HistoryType = 13
)

// History API
// https://doc.candyhouse.co/ja/SesameAPI#sesame%E3%81%AE%E5%B1%A5%E6%AD%B4%E3%82%92%E5%8F%96%E5%BE%97
// The server responses when UUID is not found or invalid. UUID string must be upper case.
func (c *Client) History(ctx context.Context, uuid string, page, maxResults int) (*HistoryResponse, error) {
	q := url.Values{}
	q.Add("page", strconv.Itoa(page))
	q.Add("lg", strconv.Itoa(maxResults))

	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/"+uuid+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating new HTTP request: %w", err)
	}

	req.Header.Add("x-api-key", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing HTTP request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var hist HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&hist); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	return &hist, nil
}
