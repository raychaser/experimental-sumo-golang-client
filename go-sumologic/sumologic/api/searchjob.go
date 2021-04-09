package api

import (
	"context"
	"fmt"
)

// SearchJobService provides access to the search related functions
// in the Sumo Logic API.
//
// Sumo Logic API docs: https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API
type SearchJobService service

type CreateRequest struct {
	Query         string `json:"query"`
	From          string `json:"from"`
	To            string `json:"to"`
	TimeZone      string `json:"timeZone"`
	ByReceiptTime bool   `json:"byReceiptTime"`
}

type CreateResult struct {
	SearchJobId string `json:"id"`
}

func (s *SearchJobService) Create(ctx context.Context, query string, from string, to string, timezone string, byReceiptTime bool, autoParsingMode string) (*CreateResult, *Response, error) {
	u := fmt.Sprintf("v1/search/jobs")
	body := CreateRequest{
		Query:         query,
		From:          from,
		To:            to,
		TimeZone:      timezone,
		ByReceiptTime: byReceiptTime,
	}
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	result := new(CreateResult)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

type StatusResult struct {
	State string `json:"state"`
	//Buckets         []Buckets `json:"histogramBuckets"`
	MessageCount    uint     `json:"messageCount"`
	RecordCount     uint     `json:"recordCount"`
	PendingWarnings []string `json:"pendingWarnings"`
	PendingErrors   []string `json:"pendingErrors"`
	//UsageDetails    string   `json:"usageDetails"`
}

type Buckets struct {
	Alias   string           `json:"alias"`
	Buckets []BucketsByAlias `json:"buckets"`
}

type BucketsByAlias struct {
	StartTimestamp int64 `json:"startTimestamp"`
	Length         uint  `json:"length"`
	Count          uint  `json:"count"`
}

type TimeRange struct {
	StartMillis uint64 `json:"startMillis"`
	EndMillis   uint64 `json:"endMillis"`
}

type MessageCountByAlias struct {
	Alias string `json:"alias"`
	Count uint   `json:"count"`
}

type SearchPerformance struct {
	Difficulty string        `json:"difficulty"`
	Reasons    []interface{} `json:"reasons"`
}

func (s *SearchJobService) Status(ctx context.Context,
	jobId string) (*StatusResult, *Response, error) {
	u := fmt.Sprintf("v1/search/jobs/%s", jobId)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	result := new(StatusResult)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

type RecordsResult struct {
	Fields  []Field  `json:"fields"`
	Records []Record `json:"records"`
}

type Field struct {
	Name      string `json:"name"`
	FieldType string `json:"fieldType"`
	KeyField  bool   `json:"keyField"`
}

type Record struct {
	Map map[string]string `json:"map"`
}

func (s *SearchJobService) Records(ctx context.Context,
	jobId string, offset uint, length uint) (*RecordsResult, *Response, error) {

	u := fmt.Sprintf("v1/search/jobs/%s/records?offset=%d&limit=%d",
		jobId, offset, length)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	result := new(RecordsResult)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

func (s *SearchJobService) Delete(ctx context.Context, jobId string) (*Response, error) {
	u := fmt.Sprintf("v1/search/jobs/%s", jobId)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	result := new(interface{})
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
