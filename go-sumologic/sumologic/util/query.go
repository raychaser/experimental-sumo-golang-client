package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sumologic/go-sumologic/sumologic/api"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Row []interface{}

type ColumnType string

const (
	ColumnType_Empty     ColumnType = "Empty"
	ColumnType_Integer   ColumnType = "Integer"
	ColumnType_Float     ColumnType = "Float"
	ColumnType_String    ColumnType = "String"
	ColumnType_Timestamp ColumnType = "Timestamp"
)

type ColumnDescriptor struct {
	ColumnType ColumnType
}

type Header struct {
	Row               Row
	ColumnDescriptors []ColumnDescriptor
}

type Table struct {
	Header Header
	Body   []Row
}

func (t *Table) SetHeader(header Header) {
	t.Header = header
}

func (t *Table) AppendRow(row []interface{}) {
	t.Body = append(t.Body, row)
}

func RowFromStrings(ss []string) Row {
	r := make([]interface{}, 0, len(ss))
	for _, s := range ss {
		r = append(r, s)
	}
	return r
}

func HeaderFromStrings(fields []string) Header {
	cds := make([]ColumnDescriptor, 0, len(fields))
	for range fields {
		cds = append(cds, ColumnDescriptor{ColumnType_String})
	}
	return Header{RowFromStrings(fields), cds}
}

//TODO error return

//TODO all fatals to Error

func DashboardSearch(client *api.Client, ctx context.Context,
	queryType string, query string, from time.Time, to time.Time,
	desiredDataPoints int32, pollInterval int64) (string, *api.DashboardSearchResultsMap, *Table, time.Duration, []api.ErrorDescription, error) {

	// Check and mangle query type
	qt := strings.ToLower(queryType)
	if qt != "logs" && qt != "metrics" {
		msg := "queryType must be 'logs' or 'metrics'"
		log.Warn().Str("queryType", queryType).Msg(msg)
		return "", nil, nil, time.Duration(0), nil, errors.New(msg)
	}
	if qt == "logs" {
		qt = "Logs"
	} else {
		qt = "Metrics"
	}

	// Query string manipulations
	if qt == "Logs" {
		//TODO: Remove this flag once we know the batchless fixed has been deployed everywhere
		query = "/* _queryFlag=usebatchless */ " + query
	}
	queryShort := query
	if len(queryShort) > 64 {
		queryShort = queryShort[0:64] + "..."
	}
	log.Info().Str("query", queryShort).Str("querytype", qt).
		Str("timerange", from.Format(time.RFC3339)+" "+to.Format(time.RFC3339)).
		Msg("Starting dashboard search")

	// Start the search job
	key, _ := uuid.NewUUID()
	timestampString := fmt.Sprintf("[{\"t\":\"absolute\",\"d\":%d},{\"t\":\"absolute\",\"d\":%d}]",
		from.UnixNano()/int64(time.Millisecond), to.UnixNano()/int64(time.Millisecond))
	requestToken, _ := uuid.NewUUID()
	request := api.DashboardSearchParameters{
		[]api.DashboardSearchParameter{{
			Key: key.String(),
			Queries: []api.Query{{
				QueryString: query,
				QueryType:   qt,
				QueryKey:    "A",
			}},
			Timerange: timestampString,
			Timezone:  "Etc/UTC", //TODO Timezone
			VariablesValues: api.VariablesValuesData{
				Data: map[string][]string{},
			},
			RequestToken:              requestToken.String(),
			DesiredNumberOfDataPoints: desiredDataPoints,
		}},
	}
	timerStart := time.Now()
	sessionIds, _, err := client.DashboardSearch.Create(ctx, request)
	if err != nil {
		msg := "Error creating dashboard search"
		log.Warn().Err(err).Msg(msg)
		return "", nil, nil, time.Duration(0), nil, err
	}
	if sessionIds.Errors != nil && len(sessionIds.Errors) > 0 {
		return "", nil, nil, time.Duration(0), sessionIds.Errors[key.String()].Errors, nil
	}
	////TODO Better?
	//result := resultsMap.Data[sessionId]
	//if result.Errors != nil {
	//	for _, error_ := range result.Errors {
	//		meta := fmt.Sprintf("%v", error.Meta)
	//		log.Warn().Str("sessionId", sessionId).Str("code", error_.Code).
	//			Str("Message", error.Message).Str("Detail", error_.Detail).
	//			Str("meta", meta).Msg("Error in result")
	//	}
	//}

	sessionId := sessionIds.Data[key.String()]
	log.Debug().Str("sessionId", sessionId).Msg("Created dashboard search")

	// Poll until the search job is done
	var resultsMap *api.DashboardSearchResultsMap
	interval := time.Duration(pollInterval) * time.Millisecond
	for {

		// Get results
		lastRequestTime := time.Now()
		resultsMap, _, err = client.DashboardSearch.GetSearchResultsForSessions(ctx, sessionId)
		if err != nil {
			msg := "Error getting dashboard search status"
			log.Warn().Str("sessionId", sessionId).Err(err).Msg(msg)
			return "", nil, nil, time.Duration(0), nil, err
		}

		// Figure out whether we are done
		result := resultsMap.Data[sessionId]
		status := result.Status.State
		percentCompleted := result.Status.PercentCompleted
		log.Debug().Str("sessionId", sessionId).Int32("percent", percentCompleted).
			Str("status", status).Msg("Status")
		if (status != "InProgress" && status != "NotStarted") ||
			(result.Errors != nil && len(result.Errors) > 0) {
			break
		}

		// Not done, figure out how long to sleep
		timeDelta := time.Now().Sub(lastRequestTime)
		sleepTime := interval - timeDelta
		if sleepTime > 0 {
			log.Debug().Str("sessionId", sessionId).
				Int64("duration", sleepTime.Milliseconds()).Msg("Sleeping")
			time.Sleep(sleepTime)
		}
	}
	elapsed := time.Since(timerStart)

	// Deal with errors
	result := resultsMap.Data[sessionId]
	if result.Errors != nil {
		for _, error_ := range result.Errors {
			log.Warn().Str("sessionId", sessionId).Str("code", error_.Code).
				Str("Detail", error_.Detail).
				Msg("Error in result")
		}
	}

	// Timing information
	log.Info().Str("sessionId", sessionId).Int64("elapsed", elapsed.Milliseconds()).
		Msg("Dashboard search timing")

	// Return result
	table := VisualDataSeriesToTable(result.Series)
	resultErrors := result.Errors
	if result.Errors == nil || len(result.Errors) < 1 {
		resultErrors = nil
	}
	return sessionId, resultsMap, table, elapsed, resultErrors, nil
}

func SearchJob(client *api.Client, ctx context.Context,
	query string, from string, to string, timezone string, pollInterval int64) *Table {

	//TODO adapt to time.Time, and multi-result

	// Start the search job
	queryString := query
	if len(queryString) > 64 {
		queryString = queryString[0:64] + "..."
	}
	log.Info().Str("query", queryString).Str("timerange", from+" "+to).
		Msg("Starting search job")
	timerStart := time.Now()
	createResult, _, err := client.SearchJob.Create(ctx, query, from, to, timezone, false, "performance")
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating search job")
	}
	searchJobId := createResult.SearchJobId
	log.Info().Str("id", searchJobId).Msg("Created search job")

	// Delete the job eventually
	defer func() {
		_, err = client.SearchJob.Delete(ctx, searchJobId)
		if err != nil {
			log.Fatal().Str("id", searchJobId).Err(err).
				Msg("Error deleting search job")
		}
		log.Info().Str("id", searchJobId).Msg("Deleted search job")
	}()

	// Poll until the search job is done
	var (
		messageCount uint
		recordCount  uint
	)
	interval := time.Duration(pollInterval) * time.Millisecond
	for {
		lastRequestTime := time.Now()
		statusResult, _, err := client.SearchJob.Status(ctx, searchJobId)
		if err != nil {
			log.Fatal().Str("id", searchJobId).Err(err).
				Msg("Error getting search job status")
		}
		//TODO Deal with errors and warnings
		log.Debug().Str("id", searchJobId).Str("status", statusResult.State).Msg("Status")
		if statusResult.State == "DONE GATHERING RESULTS" {
			messageCount = statusResult.MessageCount
			recordCount = statusResult.RecordCount
			break
		}
		timeDelta := time.Now().Sub(lastRequestTime)
		sleepTime := interval - timeDelta
		if sleepTime > 0 {
			log.Debug().Str("id", searchJobId).
				Int64("duration", sleepTime.Milliseconds()).Msg("Sleeping")
			time.Sleep(sleepTime)
		}
	}
	log.Info().Str("id", searchJobId).Uint("messages", messageCount).
		Uint("records", recordCount).Msg("Finished search job")

	//TODO: Get the messages

	// Get the records
	recordsResult := &api.RecordsResult{}
	if recordCount > 0 {
		//TODO: Page the results if large count
		recordsResult, _, err =
			client.SearchJob.Records(ctx, searchJobId, 0, recordCount)
		if err != nil {
			log.Fatal().Str("id", searchJobId).Err(err).
				Msg("Error getting search job result records")
		}
	}
	elapsed := time.Since(timerStart)

	// Timing information
	log.Info().Str("id", searchJobId).Int64("elapsed", elapsed.Milliseconds()).
		Uint("messages", messageCount).Uint("records", recordCount).
		Msg("Search job timing")

	return RecordsResultsToTable(recordsResult)
}

func MetricsAPIQuery(client *api.Client, ctx context.Context,
	query string, from time.Time, to time.Time, stepSeconds int) (*api.QueryResult, *Table, time.Duration) {

	timerStartTime := time.Now()
	queryRows := &[]api.QueryRow{{Query: query, RowId: "A"}}
	queryResult, _, err := client.MetricsQuery.Query(ctx, *queryRows,
		UnixMillis(from), UnixMillis(to), 0, 0, 0, stepSeconds)
	if err != nil {
		log.Fatal().Err(err).Msg("Error querying metrics")
	}
	elapsed := time.Since(timerStartTime)

	// Timing information
	log.Info().Int64("elapsed", elapsed.Milliseconds()).Msg("Metrics query timing")

	return queryResult, QueryResultToTable(queryResult), elapsed
}

func VisualDataSeriesToTable(dataSeries []api.VisualDataSeries) *Table {
	result := &Table{}

	// No series, no table
	if len(dataSeries) < 1 {
		return result
	}

	// Dispath based on series type
	if dataSeries[0].SeriesType == "nontimeseries" {
		return TableFromNonTimeseriesVisualDataSeries(dataSeries, result)
	} else {
		return TableFromTimeseriesVisualDataSeries(dataSeries, result)
	}
}

func TableFromNonTimeseriesVisualDataSeries(dataSeries []api.VisualDataSeries, result *Table) *Table {
	series := dataSeries[0]

	// Figure out the fields
	fieldsMap := make(map[string]bool, 0)
	for _, dp := range series.DataPoints {
		for k := range dp.XAxisValues {
			fieldsMap[k] = true
		}
	}
	fields := make([]string, 0, len(fieldsMap))
	for k := range fieldsMap {
		fields = append(fields, k)
	}
	sort.Strings(fields)

	// Append the value field name
	fields = append(fields, series.Name)

	// Append a row for each data point
	cds := make([]ColumnDescriptor, 0, len(fields))
	for i, dp := range series.DataPoints {
		row := make([]interface{}, 0, len(fields))
		for _, field := range fields {
			value := dp.XAxisValues[field]
			if field == series.Name {
				value = dp.Y
			}
			ct := determineColumnType(value, field)
			v := convertStringToColumnType(value, ct)
			row = append(row, v)

			// Set column type if first iteration
			if i == 0 {
				cds = append(cds, ColumnDescriptor{ColumnType: ct})
			}
		}
		result.AppendRow(row)
	}

	// Rename timestamp fields if any
	for i, field := range fields {
		f := strings.ToLower(field)
		if f == "_timeslice" || f == "_messagetime" {
			fields[i] = "Time"
		}
	}

	// Create the header row
	hr := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		fieldName := normalizeFieldName(field, 32)
		hr = append(hr, fieldName)
	}

	// Set the header
	header := &Header{hr, cds}
	result.SetHeader(*header)

	// And we are done
	return result
}

// Turns results from dashborads
func TableFromTimeseriesVisualDataSeries(dataSeries []api.VisualDataSeries, result *Table) *Table {

	// Figure out the header fields
	fields := make([]string, 0)
	for _, series := range dataSeries {
		fields = append(fields, series.Name)
	}
	sort.Strings(fields)

	// Get the column types
	cds := make([]ColumnDescriptor, 0)
	cds = append(cds, ColumnDescriptor{ColumnType_Timestamp})
	for _, series := range dataSeries {
		columnType := ColumnType_Empty
		for _, dp := range series.DataPoints {
			ct := determineColumnType(dp.Y, series.Name)
			if columnType == ColumnType_Empty {
				columnType = ct
			}
			if columnType == ColumnType_Integer && ct == ColumnType_Float {
				columnType = ct
			}
		}
		cds = append(cds, ColumnDescriptor{columnType})
	}

	// Get all timestamps
	timestampSet := make(map[float64]bool)
	for _, series := range dataSeries {
		for _, dataPoint := range series.DataPoints {
			timestampSet[dataPoint.X] = true
		}
	}
	timestamps := make([]float64, 0, len(timestampSet))
	for timestamp := range timestampSet {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	// Append a row for each timestamp
	//TODO This is beyond inefficient... A 3 year old could likely do this better
	for _, timestamp := range timestamps {
		row := make([]interface{}, 0)
		row = append(row, time.Unix(0, int64(timestamp)*int64(time.Millisecond)))
		for _, field := range fields {
			for i, series := range dataSeries {
				if series.Name == field {
					found := false
					for _, dataPoint := range series.DataPoints {
						if dataPoint.X == timestamp {
							v := convertStringToColumnType(dataPoint.Y, cds[i+1].ColumnType)
							row = append(row, v)
							found = true
						}
					}
					if !found {
						row = append(row, "")
					}
				}
			}
		}
		result.AppendRow(row)
	}

	// Create the header row
	hr := make([]interface{}, 0)
	hr = append(hr, "Time")
	for _, field := range fields {
		fieldName := normalizeFieldName(field, 32)
		hr = append(hr, fieldName)
	}

	// Set the header
	header := &Header{hr, cds}
	result.SetHeader(*header)

	// And we are done
	return result
}

// DashboardSearch results to Table
func RecordsResultsToTable(recordsResult *api.RecordsResult) *Table {
	result := &Table{}

	// Figure out the header fields and column types
	fields := make([]string, 0)
	cds := make([]ColumnDescriptor, 0)
	for _, field := range recordsResult.Fields {
		fields = append(fields, field.Name)
		ct := ColumnType_String
		switch field.FieldType {
		case "string":
			ct = ColumnType_String
		case "int":
			ct = ColumnType_Integer
		case "float", "double":
			ct = ColumnType_Float
		case "long":
			if field.Name == "_timeslice" {
				//TODO Other field names for timestamps?
				ct = ColumnType_Timestamp
			}
			//TODO Do we get long type in other circumstances?
		}
		cds = append(cds, ColumnDescriptor{ct})
	}
	sort.Strings(fields)
	h := make([]interface{}, 0)
	for _, field := range recordsResult.Fields {
		h = append(h, field.Name)
	}
	header := &Header{
		Row:               h,
		ColumnDescriptors: cds,
	}
	result.SetHeader(*header)

	for _, record := range recordsResult.Records {
		row := make([]interface{}, 0)
		for _, field := range recordsResult.Fields {
			value := record.Map[field.Name]
			switch field.FieldType {
			case "long":
				longValue, _ := strconv.ParseInt(value, 10, 64)
				lowerFieldName := strings.ToLower(field.Name)
				if lowerFieldName == "_timeslice" || lowerFieldName == "_messagetime" {
					row = append(row, time.Unix(0, longValue*int64(time.Millisecond)))
				} else {
					row = append(row, longValue)
				}
			case "int":
				intValue, _ := strconv.ParseInt(value, 10, 32)
				row = append(row, intValue)
			case "double":
				floatValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Fatal().Err(err).Str("float", value).
						Msg("Error converting string to float")
				}
				row = formatFloat64(floatValue, row)
			default:
				row = append(row, value)
			}
		}
		result.AppendRow(row)
	}

	return result
}

func QueryResultToTable(queryResult *api.QueryResult) *Table {
	result := &Table{}

	// Figure out the h fields
	fields := make([]string, 0)
	for _, response := range queryResult.Response {
		rowId := response.RowID
		for _, metricsResult := range response.Results {
			dimensionString := dimensionsToString(metricsResult.Metric.Dimensions)
			field := rowId + " " + dimensionString
			fields = append(fields, field)
		}
	}
	sort.Strings(fields)
	h := make([]interface{}, 0)
	h = append(h, "Time")
	for _, field := range fields {
		h = append(h, field[0:16])
	}
	header := &Header{
		Row:               h,
		ColumnDescriptors: nil,
	}
	result.SetHeader(*header)

	// Get all timestamps
	timestampSet := make(map[int64]bool)
	for _, response := range queryResult.Response {
		for _, metricsResult := range response.Results {
			for _, timestamp := range metricsResult.Datapoints.Timestamps {
				timestampSet[timestamp] = true
			}
		}
	}
	timestamps := make([]int64, 0, len(timestampSet))
	for timestamp := range timestampSet {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	// Append a row for each timestamp
	//TODO This is beyond inefficient... A 3 year old could do this better
	for _, timestamp := range timestamps {
		row := make([]interface{}, 0)
		row = append(row, time.Unix(0, timestamp*int64(time.Millisecond)))
		for _, field := range fields {
			value := 0.0
			for _, response := range queryResult.Response {
				rowId := response.RowID
				for _, metricsResult := range response.Results {
					dimensionString := dimensionsToString(metricsResult.Metric.Dimensions)
					currentField := rowId + " " + dimensionString
					if currentField == field {
						for index, currentTimestamp := range metricsResult.Datapoints.Timestamps {
							if currentTimestamp == timestamp {
								value = metricsResult.Datapoints.Values[index]
							}
						}
					}
				}
			}
			row = append(row, int64(math.Round(value)))
		}
		result.AppendRow(row)
	}

	return result
}

func determineColumnType(value string, fieldName string) ColumnType {

	// Start by assuming it is a string
	ct := ColumnType_String

	// Is it a timestamp?
	fn := strings.ToLower(fieldName)
	if fn == "_timeslice" || fn == "_messagetime" {
		return ColumnType_Timestamp
	}

	// Then see if this is a number
	//TODO parseInt("1.0") produces an error so need to also return a new value
	//     if we want to turn into a parseable int...
	//floatValue, err := strconv.ParseFloat(value, 64)
	_, err := strconv.ParseFloat(value, 64)
	if err == nil { // Looks like it is a number, but is it float or int?
		//if floatValue == float64(int64(floatValue)) {
		//	ct = ColumnType_Integer // Must be an integer
		//} else {
		ct = ColumnType_Float // Float, then!
		//}
	}

	// And that's all she wrote
	return ct
}

func convertStringToColumnType(value string, columnType ColumnType) interface{} {
	switch columnType {
	case ColumnType_Timestamp:
		return time.Unix(0, int64(toInt(value))*int64(time.Millisecond))
	case ColumnType_Integer:
		return toInt(value)
	case ColumnType_Float:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatal().Stack().Err(err).Msgf(
				"While converting to float, error on strconv.ParseFloat for %v of type %T", value, value)
		}
		return v
	}
	return value
}

func toInt(value string) int64 {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Warn().Err(err).Msgf(
			"While converting to int, error on strconv.ParseInt for %v of type %T", value, value)
		fv, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Warn().Err(err).Msgf(
				"While converting to float while trying to convert to in, "+
					"error on strconv.ParseFloat for %v of type %T", value, value)

		}
		v = int64(math.Round(fv))
	}
	return v
}

func normalizeFieldName(field string, length int) string {
	result := field
	//TODO Rationalize this
	//half := length / 2
	//if len(result) > length {
	//	result = result[0:half] + "..." + result[len(result)-half:]
	//}
	return result
}

func formatFloat64(floatValue float64, row []interface{}) []interface{} {
	if floatValue == float64(int64(floatValue)) {
		value := fmt.Sprintf("%.0f", floatValue)
		row = append(row, value)
	} else {
		value := fmt.Sprintf("%.5f", floatValue)
		row = append(row, value)
	}
	return row
}

func dimensionsToString(dimensions []api.Dimension) string {
	var result bytes.Buffer
	for index, dimension := range dimensions {
		result.WriteString(dimension.Key + "=" + dimension.Value)
		if index < len(dimensions)-1 {
			result.WriteString(" ")
		}
	}
	return result.String()
}
