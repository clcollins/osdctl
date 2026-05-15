package rhobs

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/sync/singleflight"
)

var fetcherCache sync.Map
var fetcherInit singleflight.Group

func getCachedFetcher(clusterId string, usage RhobsFetchUsage) (*RhobsFetcher, error) {
	key := fmt.Sprintf("%s:%s", clusterId, usage)
	if cached, ok := fetcherCache.Load(key); ok {
		return cached.(*RhobsFetcher), nil
	}

	v, err, _ := fetcherInit.Do(key, func() (interface{}, error) {
		if cached, ok := fetcherCache.Load(key); ok {
			return cached, nil
		}
		fetcher, err := CreateRhobsFetcher(clusterId, usage, commonOptions.hiveOcmUrl)
		if err != nil {
			return nil, err
		}
		actual, _ := fetcherCache.LoadOrStore(key, fetcher)
		return actual, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*RhobsFetcher), nil
}

func mcpResultJSON(data interface{}) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %v", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
	}, nil
}

func mcpError(format string, args ...interface{}) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
		IsError: true,
	}, nil
}

func boolPtr(b bool) *bool {
	return &b
}
