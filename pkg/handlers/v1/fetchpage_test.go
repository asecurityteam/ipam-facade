package v1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestFetchSubnets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit := 10
	offset := 0

	expectedSubnets := []domain.AssetSubnet{
		{
			Network: "0.0.0.0/32",
		},
	}

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchSubnets(gomock.Any(), limit, offset).Return(expectedSubnets, nil)

	h := &FetchPageHandler{
		Fetcher: mockFetcher,
		LogFn:   testLogFn,
	}
	result, err := h.FetchSubnets(context.Background(), PaginationRequest{
		Limit:  limit,
		Offset: offset,
	})

	require.NoError(t, err)

	subnets := result.Result.([]Subnet)
	require.Equal(t, len(expectedSubnets), len(subnets))

	pr, _ := pageFromToken(result.NextPageToken)
	require.Equal(t, limit+offset, pr.Offset)
}

func TestFetchSubnetsDefaultLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit := 10
	offset := 0

	expectedSubnets := []domain.AssetSubnet{
		{
			Network: "0.0.0.0/32",
		},
	}

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchSubnets(gomock.Any(), limit, offset).Return(expectedSubnets, nil)

	h := &FetchPageHandler{
		Fetcher:         mockFetcher,
		LogFn:           testLogFn,
		DefaultPageSize: limit,
	}
	result, err := h.FetchSubnets(context.Background(), PaginationRequest{})

	require.NoError(t, err)

	subnets := result.Result.([]Subnet)
	require.Equal(t, len(expectedSubnets), len(subnets))

	pr, _ := pageFromToken(result.NextPageToken)
	require.Equal(t, limit+offset, pr.Offset)
}

func TestFetchSubnetsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit := 10
	offset := 0

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchSubnets(gomock.Any(), limit, offset).Return(nil, errors.New(""))

	h := &FetchPageHandler{
		Fetcher:         mockFetcher,
		LogFn:           testLogFn,
		DefaultPageSize: limit,
	}
	_, err := h.FetchSubnets(context.Background(), PaginationRequest{})

	require.Error(t, err)
}

func TestFetchIPs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit := 10
	offset := 0

	expectedIPs := []domain.AssetIP{
		{
			IP: "0.0.0.0",
		},
	}

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchIPs(gomock.Any(), limit, offset).Return(expectedIPs, nil)

	h := &FetchPageHandler{
		Fetcher: mockFetcher,
		LogFn:   testLogFn,
	}
	result, err := h.FetchIPs(context.Background(), PaginationRequest{
		Limit:  limit,
		Offset: offset,
	})

	require.NoError(t, err)

	ips := result.Result.([]IP)
	require.Equal(t, len(expectedIPs), len(ips))

	pr, _ := pageFromToken(result.NextPageToken)
	require.Equal(t, limit+offset, pr.Offset)
}

func TestFetchIPsDefaultLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit := 10
	offset := 0

	expectedIPs := []domain.AssetIP{
		{
			IP: "0.0.0.0/32",
		},
	}

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchIPs(gomock.Any(), limit, offset).Return(expectedIPs, nil)

	h := &FetchPageHandler{
		Fetcher:         mockFetcher,
		LogFn:           testLogFn,
		DefaultPageSize: limit,
	}
	result, err := h.FetchIPs(context.Background(), PaginationRequest{})

	require.NoError(t, err)

	ips := result.Result.([]IP)
	require.Equal(t, len(expectedIPs), len(ips))

	fmt.Println(result.NextPageToken)

	pr, _ := pageFromToken(result.NextPageToken)
	require.Equal(t, limit+offset, pr.Offset)
}

func TestFetchIPsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	limit := 10
	offset := 0

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchIPs(gomock.Any(), limit, offset).Return(nil, errors.New(""))

	h := &FetchPageHandler{
		Fetcher:         mockFetcher,
		LogFn:           testLogFn,
		DefaultPageSize: limit,
	}
	_, err := h.FetchIPs(context.Background(), PaginationRequest{})

	require.Error(t, err)
}

func TestFetchNextSubnets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchSubnets(gomock.Any(), 10, 10).Return([]domain.AssetSubnet{{Network: "0.0.0.0/32"}}, nil)

	h := &FetchPageHandler{
		Fetcher: mockFetcher,
		LogFn:   testLogFn,
	}
	pr, err := h.FetchNextSubnets(context.Background(), NextPageRequest{NextPageToken: "PMRGY2LNNF2CEORRGAWCE33GMZZWK5BCHIYTA7I"})
	require.NoError(t, err)

	subnets := pr.Result.([]Subnet)
	require.Equal(t, 1, len(subnets))
}

func TestFetchNextSubnetsError(t *testing.T) {
	h := &FetchPageHandler{
		Fetcher: nil,
		LogFn:   testLogFn,
	}
	_, err := h.FetchNextSubnets(context.Background(), NextPageRequest{NextPageToken: "not valid"})
	require.Error(t, err)
}

func TestFetchNextIPs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := NewMockFetcher(ctrl)
	mockFetcher.EXPECT().FetchIPs(gomock.Any(), 10, 10).Return([]domain.AssetIP{{IP: "0.0.0.0"}}, nil)

	h := &FetchPageHandler{
		Fetcher: mockFetcher,
		LogFn:   testLogFn,
	}
	pr, err := h.FetchNextIPs(context.Background(), NextPageRequest{NextPageToken: "PMRGY2LNNF2CEORRGAWCE33GMZZWK5BCHIYTA7I"})
	require.NoError(t, err)

	ips := pr.Result.([]IP)
	require.Equal(t, 1, len(ips))
}

func TestFetchNextIPsError(t *testing.T) {
	h := &FetchPageHandler{
		Fetcher: nil,
		LogFn:   testLogFn,
	}
	_, err := h.FetchNextIPs(context.Background(), NextPageRequest{NextPageToken: "not valid"})
	require.Error(t, err)
}
