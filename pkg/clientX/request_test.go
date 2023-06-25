package clientX

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type Repository struct {
	ID              int        `json:"id"`
	NodeID          string     `json:"node_id"`
	Name            string     `json:"name"`
	FullName        string     `json:"full_name"`
	Owner           *Developer `json:"owner"`
	Private         bool       `json:"private"`
	Description     string     `json:"description"`
	Fork            bool       `json:"fork"`
	Language        string     `json:"language"`
	ForksCount      int        `json:"forks_count"`
	StargazersCount int        `json:"stargazers_count"`
	WatchersCount   int        `json:"watchers_count"`
	OpenIssuesCount int        `json:"open_issues_count"`
}

type Developer struct {
	Login      string `json:"login"`
	ID         int    `json:"id"`
	NodeID     string `json:"node_id"`
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	Type       string `json:"type"`
	SiteAdmin  bool   `json:"site_admin"`
}

func TestResty(t *testing.T) {
	client := resty.New().EnableTrace()

	var result []*Repository
	request := client.R().
		SetAuthToken("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").
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetQueryParams(map[string]string{
			"per_page":  "1",
			"page":      "1",
			"sort":      "created",
			"direction": "asc",
		}).
		SetPathParams(map[string]string{
			"org": "ecodeclub",
		}).SetResult(&result)

	resp, err := request.Get("https://api.github.com/orgs/{org}/repos")

	if err != nil {
		t.Fatal(err)
	}

	ti := resp.Request.TraceInfo()

	fmt.Println("Request Trace Info:")
	fmt.Println("DNSLookup:", ti.DNSLookup)
	fmt.Println("ConnTime:", ti.ConnTime)
	fmt.Println("TCPConnTime:", ti.TCPConnTime)
	fmt.Println("TLSHandshake:", ti.TLSHandshake)
	fmt.Println("ServerTime:", ti.ServerTime)
	fmt.Println("ResponseTime:", ti.ResponseTime)
	fmt.Println("TotalTime:", ti.TotalTime)
	fmt.Println("IsConnReused:", ti.IsConnReused)
	fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
	fmt.Println("ConnIdleTime:", ti.ConnIdleTime)
	fmt.Println("RequestAttempt:", ti.RequestAttempt)
	fmt.Println("RemoteAddr:", ti.RemoteAddr.String())

	for i, repo := range result {
		fmt.Printf("repo%d: name:%s stars:%d forks:%d\n", i+1, repo.Name, repo.StargazersCount, repo.ForksCount)
	}

	//fmt.Println("======================================")
	//fmt.Println("Response Info:")
	//fmt.Println("Status Code:", resp.StatusCode())
	//fmt.Println("Status:", resp.Status())
	//fmt.Println("Proto:", resp.Proto())
	//fmt.Println("Time:", resp.Time())
	//fmt.Println("Received At:", resp.ReceivedAt())
	//fmt.Println("Size:", resp.Size())
	//fmt.Println("Headers:")
	//for key, value := range resp.Header() {
	//	fmt.Println(key, "=", value)
	//}
	//fmt.Println("Cookies:")
	//for i, cookie := range resp.Cookies() {
	//	fmt.Printf("cookie%d: name:%s value:%s\n", i, cookie.Name, cookie.Value)
	//}
}

func TestGo(t *testing.T) {
	tests := []struct {
		name    string
		srvName string
		request any
		result  any
		wantErr bool
		wantRes any
	}{
		{
			name:    "base",
			srvName: "test",
			request: &HttpRequest{
				Header: map[string]string{"Accept": "application/vnd.github.v3+json"},
				Method: "GET",
				Path:   "/orgs/ecodeclub/repos",
				QueryParams: map[string]string{
					"per_page":  "1",
					"page":      "1",
					"sort":      "created",
					"direction": "asc",
				},
			},
			result: []*Repository{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewServices("test")
			assert.Nil(t, err)
			err = Go(context.Background(), tt.srvName, tt.request, &tt.result)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantRes, tt.result)
		})
	}
}
