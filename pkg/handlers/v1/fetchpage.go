package v1

import (
	"context"
	"encoding/base32"
	"encoding/json"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/ipam-facade/pkg/logs"
)

// PaginationRequest contains information for paging through subnets
type PaginationRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PaginationResponse contains information for paging through subnets
type PaginationResponse struct {
	NextPageToken string      `json:"nextPageToken"`
	Result        interface{} `json:"result"`
}

// NextPageRequest contains a token indicating which page to fetch next
type NextPageRequest struct {
	NextPageToken string `json:"nextPageToken"`
}

// Subnet represents information about a subnet
type Subnet struct {
	Network       string `json:"network"`
	ResourceOwner string `json:"resourceOwner"`
	BusinessUnit  string `json:"businessUnit"`
	Location      string `json:"location"`
}

// IP represents informatino about an IP
type IP struct {
	IP            string `json:"ip"`
	Network       string `json:"network"`
	ResourceOwner string `json:"resourceOwner"`
	BusinessUnit  string `json:"businessUnit"`
	Location      string `json:"location"`
}

// FetchPageHandler handles requests for fetching pages of data
type FetchPageHandler struct {
	Fetcher         domain.Fetcher
	LogFn           domain.LogFn
	DefaultPageSize int
}

// FetchSubnets gets and returns a page of subnets
func (f *FetchPageHandler) FetchSubnets(ctx context.Context, input PaginationRequest) (PaginationResponse, error) {
	if input.Limit == 0 {
		input.Limit = f.DefaultPageSize
	}
	subnets, err := f.Fetcher.FetchSubnets(ctx, input.Limit, input.Offset)
	if err != nil {
		f.LogFn(ctx).Error(logs.AssetFetcherFailure{Reason: err.Error()})
		return PaginationResponse{}, err
	}
	result := make([]Subnet, 0, len(subnets))
	for _, subnet := range subnets {
		result = append(result, Subnet(subnet))
	}
	npt := ""                        // empty nextPageToken in the response is indicator to the caller that this returned page is the last
	if len(subnets) == input.Limit { // there is probably a next page
		npt = getNextPageToken(input)
	}
	return PaginationResponse{
		NextPageToken: npt,
		Result:        result,
	}, nil
}

// FetchNextSubnets fetches the next page of subnets
func (f *FetchPageHandler) FetchNextSubnets(ctx context.Context, input NextPageRequest) (PaginationResponse, error) {
	pr, err := pageFromToken(input.NextPageToken)
	if err != nil {
		f.LogFn(ctx).Info(logs.InvalidInput{Reason: err.Error()})
		return PaginationResponse{}, domain.InvalidInput{Input: input.NextPageToken}
	}
	return f.FetchSubnets(ctx, pr)
}

// FetchIPs gets and returns a page of IPs
func (f *FetchPageHandler) FetchIPs(ctx context.Context, input PaginationRequest) (PaginationResponse, error) {
	if input.Limit == 0 {
		input.Limit = f.DefaultPageSize
	}
	ips, err := f.Fetcher.FetchIPs(ctx, input.Limit, input.Offset)
	if err != nil {
		f.LogFn(ctx).Error(logs.AssetFetcherFailure{Reason: err.Error()})
		return PaginationResponse{}, err
	}
	result := make([]IP, 0, len(ips))
	for _, ip := range ips {
		result = append(result, IP(ip))
	}
	npt := ""                    // empty nextPageToken in the response is indicator to the caller that this returned page is the last
	if len(ips) == input.Limit { // there is probably a next page
		npt = getNextPageToken(input)
	}
	return PaginationResponse{
		NextPageToken: npt,
		Result:        result,
	}, nil
}

// FetchNextIPs fetches the next page of IPs
func (f *FetchPageHandler) FetchNextIPs(ctx context.Context, input NextPageRequest) (PaginationResponse, error) {
	pr, err := pageFromToken(input.NextPageToken)
	if err != nil {
		f.LogFn(ctx).Info(logs.InvalidInput{Reason: err.Error()})
		return PaginationResponse{}, domain.InvalidInput{Input: input.NextPageToken}
	}
	return f.FetchIPs(ctx, pr)
}

func getNextPageToken(pr PaginationRequest) string {
	pr.Offset = pr.Offset + pr.Limit
	js, _ := json.Marshal(pr)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(js)
}

func pageFromToken(token string) (PaginationRequest, error) {
	js, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(token)
	if err != nil {
		return PaginationRequest{}, err
	}
	var pr PaginationRequest
	err = json.Unmarshal(js, &pr)
	if err != nil {
		return PaginationRequest{}, err
	}
	return pr, nil
}
