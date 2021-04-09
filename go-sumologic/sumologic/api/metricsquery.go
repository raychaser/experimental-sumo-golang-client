package api

import (
	"context"
	"fmt"
)

// MetricsQueryService provides access to the search related functions
// in the Sumo Logic API.
//
// Sumo Logic API docs: INDECENT EXPOSURE IN PUBLIC
type MetricsQueryService service

type QueryRequest struct {
	Query                     []QueryRow `json:"query"`
	StartTime                 int64      `json:"startTime"`
	EndTime                   int64      `json:"endTime"`
	RequestedDataPoints       int        `json:"requestedDataPoints,omitempty"`
	MaxDataPoints             int        `json:"maxDataPoints,omitempty"`
	MaxTotalDataPoints        int        `json:"maxTotalDataPoints,omitempty"`
	DesiredQuantizationInSecs int        `json:"desiredQuantizationInSecs,omitempty"`
}

type QueryRow struct {
	Query string `json:"query"`
	RowId string `json:"rowId"`
}

//TODO Make desired quantization into a Duration
func (s *MetricsQueryService) Query(ctx context.Context, queryRows []QueryRow, startTime int64,
	endTime int64, requestedDataPoints int, maxDataPoints int, maxTotalDataPoints int,
	desiredQuantizationInSecs int) (*QueryResult, *Response, error) {

	u := fmt.Sprintf("v1/metrics/annotated/results")
	body := QueryRequest{
		Query:                     queryRows,
		StartTime:                 startTime,
		EndTime:                   endTime,
		RequestedDataPoints:       requestedDataPoints,
		MaxDataPoints:             maxDataPoints,
		MaxTotalDataPoints:        maxTotalDataPoints,
		DesiredQuantizationInSecs: desiredQuantizationInSecs,
	}
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	result := new(QueryResult)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

type QueryResult struct {
	Response []struct {
		RowID   string          `json:"rowId"`
		Results []MetricsResult `json:"results"`
	} `json:"response"`
	QueryInfo struct {
		StartTime                 int64 `json:"startTime"`
		EndTime                   int64 `json:"endTime"`
		DesiredQuantizationInSecs struct {
			Empty   bool `json:"empty"`
			Defined bool `json:"defined"`
		} `json:"desiredQuantizationInSecs"`
		ActualQuantizationInSecs int `json:"actualQuantizationInSecs"`
	} `json:"queryInfo"`
}

type MetricsResult struct {
	Metric     Metric     `json:"metric"`
	Datapoints Datapoints `json:"datapoints"`
}

type Metric struct {
	Dimensions []Dimension `json:"dimensions"`
	AlgoID     int         `json:"algoId"`
}

type Dimension struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Datapoints struct {
	Timestamps    []int64     `json:"timestamp"`
	Values        []float64   `json:"value"`
	OutlierParams [][]float64 `json:"outlierParams"`
	Max           []float64   `json:"max"`
	Min           []float64   `json:"min"`
	Avg           []float64   `json:"avg"`
	Count         []int       `json:"count"`
	IsFilled      []bool      `json:"isFilled"`
}
