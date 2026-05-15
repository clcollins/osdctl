package rhobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	rhobsclient "github.com/observatorium/api/client"
	rhobsparameters "github.com/observatorium/api/client/parameters"
)

type mcpLogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string            `json:"message"`
	Stream    map[string]string `json:"stream,omitempty"`
}

func (q *RhobsFetcher) QueryInstantMetrics(ctx context.Context, promExpr string, filterCluster bool) ([]instantMetricResult, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}

	promQuery := rhobsparameters.PromqlQuery(promExpr)
	queryParams := &rhobsclient.GetInstantQueryParams{Query: &promQuery}

	response, err := client.GetInstantQueryWithResponse(ctx, "hcp", queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to RHOBS: %v", err)
	}
	if response.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RHOBS query failed with status code: %d - body: %s", response.HTTPResponse.StatusCode, string(response.Body))
	}

	var formattedResponse getInstantMetricsResponse
	if err := json.Unmarshal(response.Body, &formattedResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from RHOBS: %v", err)
	}
	if formattedResponse.Status != "success" {
		return nil, fmt.Errorf("RHOBS query failed with status: %s", formattedResponse.Status)
	}

	if !filterCluster {
		return formattedResponse.Data.Results, nil
	}

	var filtered []instantMetricResult
	for _, result := range formattedResponse.Data.Results {
		if q.isManagementCluster {
			mcId := result.Metric["_mc_id"]
			mcName := result.Metric["mc_name"]
			if mcId == q.clusterId || mcName == q.clusterName {
				filtered = append(filtered, result)
			}
		} else {
			if result.Metric["_id"] == q.clusterExternalId {
				filtered = append(filtered, result)
			}
		}
	}
	return filtered, nil
}

func (q *RhobsFetcher) QueryRangeMetrics(ctx context.Context, promExpr, start, end, step string) (json.RawMessage, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}

	promQuery := rhobsparameters.PromqlQuery(promExpr)
	queryParams := &rhobsclient.GetRangeQueryParams{
		Query: &promQuery,
		Start: (*rhobsparameters.StartTS)(&start),
		End:   (*rhobsparameters.EndTS)(&end),
		Step:  &step,
	}

	response, err := client.GetRangeQueryWithResponse(ctx, "hcp", queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to RHOBS: %v", err)
	}
	if response.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RHOBS query failed with status code: %d - body: %s", response.HTTPResponse.StatusCode, string(response.Body))
	}

	return response.Body, nil
}

func (q *RhobsFetcher) QueryLogs(ctx context.Context, lokiExpr string, startTime, endTime time.Time, logsCount int) ([]mcpLogEntry, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}

	startTimeStamp := startTime.UnixNano()
	endTimeStamp := endTime.UnixNano()
	logsDir := "backward"

	var entries []mcpLogEntry

	for logsCount > 0 {
		lokiQuery := rhobsparameters.LogqlQuery(lokiExpr)
		startTimeStr := strconv.FormatInt(startTimeStamp, 10)
		endTimeStr := strconv.FormatInt(endTimeStamp, 10)

		limit := float32(500)
		if logsCount < 500 {
			limit = float32(logsCount)
		}

		queryParams := &rhobsclient.GetLogRangeQueryParams{
			Query:     &lokiQuery,
			Start:     (*rhobsparameters.StartTS)(&startTimeStr),
			End:       (*rhobsparameters.EndTS)(&endTimeStr),
			Direction: &logsDir,
			Limit:     (*rhobsparameters.Limit)(&limit),
		}

		response, err := client.GetLogRangeQueryWithResponse(ctx, "hcp", queryParams)
		if err != nil {
			return nil, fmt.Errorf("failed to send request to RHOBS: %v", err)
		}
		if response.HTTPResponse.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("RHOBS query failed with status code: %d - body: %s", response.HTTPResponse.StatusCode, string(response.Body))
		}

		var formattedResponse getLogsResponse
		if err := json.Unmarshal(response.Body, &formattedResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response from RHOBS: %v", err)
		}
		if formattedResponse.Status != "success" {
			return nil, fmt.Errorf("RHOBS query failed with status: %s", formattedResponse.Status)
		}
		if len(formattedResponse.Data.Results) == 0 {
			break
		}

		var flattenedResults []*logResult
		for _, result := range formattedResponse.Data.Results {
			for valIdx := range result.Values {
				flattenedResults = append(flattenedResults, &logResult{
					Stream: result.Stream,
					Values: []*[]string{result.Values[valIdx]},
				})
			}
		}

		sort.Slice(flattenedResults, func(i, j int) bool {
			return flattenedResults[i].getTimeStamp() > flattenedResults[j].getTimeStamp()
		})

		edgeTimeStamp := flattenedResults[len(flattenedResults)-1].getTimeStamp()
		if flattenedResults[0].getTimeStamp() == edgeTimeStamp {
			endTimeStamp = edgeTimeStamp - 1
			edgeTimeStamp = 0
		} else {
			endTimeStamp = edgeTimeStamp
		}

		for _, result := range flattenedResults {
			ts := result.getTimeStamp()
			if ts != edgeTimeStamp {
				entry := mcpLogEntry{
					Timestamp: result.getTime(),
					Message:   result.getMessage(),
				}
				if result.Stream != nil {
					entry.Stream = *result.Stream
				}
				entries = append(entries, entry)
				logsCount--
				if logsCount <= 0 {
					break
				}
			}
		}
	}

	return entries, nil
}

func (q *RhobsFetcher) QueryAlerts(ctx context.Context) (json.RawMessage, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetAlertsWithResponse(ctx, "hcp", &rhobsclient.GetAlertsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to send request to RHOBS: %v", err)
	}
	if response.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RHOBS query failed with status code: %d - body: %s", response.HTTPResponse.StatusCode, string(response.Body))
	}

	return response.Body, nil
}

func (q *RhobsFetcher) QueryRules(ctx context.Context, ruleType string) (json.RawMessage, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}

	params := &rhobsclient.GetRulesParams{}
	if ruleType != "" {
		params.Type = &ruleType
	}

	response, err := client.GetRulesWithResponse(ctx, "hcp", params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to RHOBS: %v", err)
	}
	if response.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RHOBS query failed with status code: %d - body: %s", response.HTTPResponse.StatusCode, string(response.Body))
	}

	return response.Body, nil
}
