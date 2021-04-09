package api

import (
	"context"
	"fmt"
	"time"
)

type DashboardSearchService service

type DashboardSearchParameters struct {
	// List of search queries.
	Data []DashboardSearchParameter `json:"data"`
}

type DashboardSearchParameter struct {
	// A unique key to identify the search.
	Key string `json:"key"`
	// Search queries.
	Queries []Query `json:"queries"`
	// Timerange of the search query.
	Timerange string `json:"timerange"`
	// Timezone of the search query.
	Timezone        string              `json:"timezone"`
	VariablesValues VariablesValuesData `json:"variablesValues"`
	// The number of desired data points per series.
	DesiredNumberOfDataPoints int32 `json:"desiredNumberOfDataPoints,omitempty"`
	// The maximum number of data points per series.
	MaximumNumberOfDataPoints int32 `json:"maximumNumberOfDataPoints,omitempty"`
	// A user-generated string to uniquely identify the search request. This field can be safely ignored if you don't intend to identify a search request.
	RequestToken string `json:"requestToken,omitempty"`
	// The list of metadata on which the search will repeat.
	RepeatedMetaData []string `json:"repeatedMetaData,omitempty"`
}

type Query struct {
	// The metrics or logs query.
	QueryString string `json:"queryString"`
	// The type of the query, either `Metrics` or `Logs`.
	QueryType string `json:"queryType"`
	// The key for metric or log queries. Used as an identifier for queries.
	QueryKey string `json:"queryKey"`
	// The mode of the metrics query that the user was editing. Can be `Basic` or `Advanced`. Will ONLY be specified for metrics queries.
	MetricsQueryMode string           `json:"metricsQueryMode,omitempty"`
	MetricsQueryData MetricsQueryData `json:"metricsQueryData,omitempty"`
}

type MetricsQueryData struct {
	// The metric of the query.
	Metric string `json:"metric"`
	// The type of aggregation. Can be `Count`, `Minimum`, `Maximum`, `Sum`, `Average` or `None`.
	AggregationType string `json:"aggregationType,omitempty"`
	// The field to group the results by.
	GroupBy string `json:"groupBy,omitempty"`
	// A list of filters for the metrics query.
	Filters []MetricsFilter `json:"filters,omitempty"`
	// A list of operator data for the metrics query.
	Operators []OperatorData `json:"operators,omitempty"`
}

type MetricsFilter struct {
	// The key of the metrics filter.
	Key string `json:"key,omitempty"`
	// The value of the metrics filter.
	Value string `json:"value"`
	// Whether or not the metrics filter is negated.
	Negation bool `json:"negation,omitempty"`
}

type OperatorData struct {
	// The name of the metrics operator.
	OperatorName string `json:"operatorName"`
	// A list of operator parameters for the operator data.
	Parameters []OperatorParameter `json:"parameters"`
}

type OperatorParameter struct {
	// The key of the operator parameter.
	Key string `json:"key"`
	// The value of the operator parameter.
	Value string `json:"value"`
}

type VariablesValuesData struct {
	// Data for variable values.
	Data map[string][]string `json:"data"`
	// Data for variable values and last run time.
	RichData map[string]VariableValuesData `json:"richData,omitempty"`
}

type VariableValuesData struct {
	// Values for the variable.
	VariableValues []string `json:"variableValues,omitempty"`
	// The last time we fetched values for the variable. If value is `null`, values have not been fetched.
	LastRunTime time.Time `json:"lastRunTime,omitempty"`
}

type DashboardSearchSessionIds struct {
	// Map of search keys to session ids.
	Data map[string]string `json:"data"`
	// Error description for the session keys that failed validation.
	Errors map[string]XXXErrorResponse `json:"errors,omitempty"`
}

func (s *DashboardSearchService) Create(ctx context.Context,
	parameters DashboardSearchParameters) (*DashboardSearchSessionIds, *Response, error) {

	u := fmt.Sprintf("v1alpha/dashboard/searches")
	body := parameters
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	//TODO Errors don't seem to deserialize anywhere?
	result := new(DashboardSearchSessionIds)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

type DashboardSearchResultsMap struct {
	// Map of session id to search result.
	Data map[string]DashboardSearchResult `json:"data"`
}

type DashboardSearchResult struct {
	Status DashboardSearchStatus `json:"status"`
	Axes   VisualDataAxes        `json:"axes"`
	// The series returned from a search.
	Series []VisualDataSeries `json:"series"`
	// Errors returned by backend.
	Errors    []ErrorDescription    `json:"errors,omitempty"`
	TimeRange BeginBoundedTimeRange `json:"timeRange,omitempty"`
	// A user-generated string to uniquely identify the search request. This field can be safely ignored if you don't intend to identify a search request.
	RequestToken string `json:"requestToken,omitempty"`
}

type DashboardSearchStatus struct {

	// Current state of the search.
	State string `json:"state"`
	// Percentage of search completed.
	PercentCompleted int32 `json:"percentCompleted,omitempty"`
}

type VisualDataAxes struct {
	// The data of the primary x axis.
	X []VisualAxisData `json:"x"`
	// The data of the primary y axis.
	Y []VisualAxisData `json:"y"`
	// The data of the secondary x axis.
	X2 []VisualAxisData `json:"x2,omitempty"`
	// The data of the secondary y axis.
	Y2 []VisualAxisData `json:"y2,omitempty"`
}

type VisualAxisData struct {
	// The value of the axis labels.
	Index int32 `json:"index,omitempty"`
}

type VisualDataSeries struct {
	// The id of the query.
	QueryId string `json:"queryId"`
	// The name of the query.
	Name string `json:"name"`
	// A list of data points in the visual series.
	DataPoints    []VisualPointData   `json:"dataPoints"`
	AggregateInfo VisualAggregateData `json:"aggregateInfo,omitempty"`
	MetaData      VisualMetaData      `json:"metaData,omitempty"`
	// Type of the visual series.
	SeriesType string `json:"seriesType,omitempty"`
	// Keys that will be plotted as a point on the x axis.
	XAxisKeys []string `json:"xAxisKeys,omitempty"`
	// Value that represents if the series values are String or Double
	ValueType string `json:"valueType,omitempty"`
}

type VisualPointData struct {
	// Value that represents a point on the x axis.
	X float64 `json:"x,omitempty"`
	// Value that represents a point on the y axis.
	Y string `json:"y"`
	// Whether the field is interpolated or extrapolated - not derived from underlying data.
	IsFilled bool `json:"isFilled,omitempty"`
	// Values that represents a point on the x axis.
	XAxisValues map[string]string `json:"xAxisValues,omitempty"`
	OutlierData VisualOutlierData `json:"outlierData,omitempty"`
}

type VisualAggregateData struct {
	// The maximum value in the series.
	Max float64 `json:"max"`
	// The minimum value in the series.
	Min float64 `json:"min"`
	// The average value in the series.
	Avg float64 `json:"avg"`
	// The sum of all the values in the series.
	Sum float64 `json:"sum"`
	// The last value in the series.
	Latest float64 `json:"latest"`
	// The number of values in the series.
	Count float64 `json:"count,omitempty"`
}

type VisualMetaData struct {
	// The value of the metadata.
	Data map[string]string `json:"data"`
}

type VisualOutlierData struct {
	// A measure of how anomalous the data point is.
	AnomalyScore float64 `json:"anomalyScore"`
	// The estimated value of the data point.
	Baseline float64 `json:"baseline"`
	// The variation in the estimated value of the data point.
	Unit float64 `json:"unit"`
}

//TODO Find a way to resolve name clash with ErrorResponse in sumologic.go
type XXXErrorResponse struct {
	// An identifier for the error; this is unique to the specific API request.
	Id string `json:"id"`
	// A list of one or more causes of the error.
	Errors []ErrorDescription `json:"errors"`
}

type ErrorDescription struct {
	// An error code describing the type of error.
	Code string `json:"code"`
	// A short English-language description of the error.
	Message string `json:"message"`
	// An optional fuller English-language description of the error.
	Detail string `json:"detail,omitempty"`
	// An optional list of metadata about the error.
	Meta map[string]interface{} `json:"meta,omitempty"`
}

type BeginBoundedTimeRange struct {
	// Type of the time range. Value must be either `CompleteLiteralTimeRange` or `BeginBoundedTimeRange`.
	Type string            `json:"type"`
	From TimeRangeBoundary `json:"from"`
	To   TimeRangeBoundary `json:"to,omitempty"`
}

type TimeRangeBoundary struct {
	// Type of the time range boundary. Value must be from list: - `RelativeTimeRangeBoundary`, - `EpochTimeRangeBoundary`, - `Iso8601TimeRangeBoundary`, - `LiteralTimeRangeBoundary`.
	Type string `json:"type"`
}

func (s *DashboardSearchService) GetSearchResultsForSessions(ctx context.Context,
	sessionIds string) (*DashboardSearchResultsMap, *Response, error) {

	u := fmt.Sprintf("v1alpha/dashboard/searches/%s", sessionIds)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	result := new(DashboardSearchResultsMap)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

func (s *DashboardSearchService) Delete(ctx context.Context, sessionIds string) (*Response, error) {
	u := fmt.Sprintf("v1alpha/dashboard/searches/%s", sessionIds)
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
