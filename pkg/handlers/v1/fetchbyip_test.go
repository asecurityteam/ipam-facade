package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestPhysicalAssetToResponse(t *testing.T) {
	asset := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
		Network:       "127.0.0.0/31",
		Location:      "",
		DeviceID:      1,
		SubnetID:      1,
		CustomerID:    1,
	}
	expectedResult := PhysicalAssetDetails{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
		Tags: tags{
			Network:    "127.0.0.0/31",
			Location:   "",
			DeviceID:   "1",
			SubnetID:   "1",
			CustomerID: "1",
		},
	}

	result := physicalAssetToResponse(asset)
	require.Equal(t, expectedResult, result)
}

func TestPhysicalAssetToResponseZeroValues(t *testing.T) {
	asset := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
		Network:       "127.0.0.0/31",
		Location:      "",
		DeviceID:      0,
		SubnetID:      1,
		CustomerID:    0,
	}
	expectedResult := PhysicalAssetDetails{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
		Tags: tags{
			Network:    "127.0.0.0/31",
			Location:   "",
			DeviceID:   "",
			SubnetID:   "1",
			CustomerID: "",
		},
	}

	result := physicalAssetToResponse(asset)
	require.Equal(t, expectedResult, result)
}

func TestFetchHandlerInvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPhysicalAssetFetcher := NewMockPhysicalAssetFetcher(ctrl)
	handler := FetchByIPAddressHandler{
		PhysicalAssetFetcher: mockPhysicalAssetFetcher,
		LogFn:                testLogFn,
	}

	response, err := handler.Handle(context.Background(), IPAddressQuery{IPAddress: "boom!"})
	require.Equal(t, PhysicalAssetDetails{}, response)
	require.Equal(t, domain.InvalidInput{Input: "boom!"}, err)
}

func TestFetchHandlerAssetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testIP := "127.0.0.1"

	mockPhysicalAssetFetcher := NewMockPhysicalAssetFetcher(ctrl)
	handler := FetchByIPAddressHandler{
		PhysicalAssetFetcher: mockPhysicalAssetFetcher,
		LogFn:                testLogFn,
	}

	mockPhysicalAssetFetcher.EXPECT().FetchPhysicalAsset(gomock.Any(), testIP).Return(
		domain.PhysicalAsset{}, domain.AssetNotFound{IP: testIP})
	response, err := handler.Handle(context.Background(), IPAddressQuery{IPAddress: testIP})
	require.Equal(t, PhysicalAssetDetails{}, response)
	require.Equal(t, domain.AssetNotFound{IP: testIP}, err)
}

func TestFetchHandlerAssetFetcherFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testIP := "127.0.0.1"
	fetchError := domain.AssetFetchError{Inner: errors.New("bang"), IP: testIP}

	mockPhysicalAssetFetcher := NewMockPhysicalAssetFetcher(ctrl)
	handler := FetchByIPAddressHandler{
		PhysicalAssetFetcher: mockPhysicalAssetFetcher,
		LogFn:                testLogFn,
	}

	mockPhysicalAssetFetcher.EXPECT().FetchPhysicalAsset(gomock.Any(), testIP).Return(
		domain.PhysicalAsset{}, fetchError)
	response, err := handler.Handle(context.Background(), IPAddressQuery{IPAddress: testIP})
	require.Equal(t, PhysicalAssetDetails{}, response)
	require.Equal(t, fetchError, err)
}

func TestFetchHandlerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	asset := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
		Network:       "127.0.0.0/31",
		Location:      "",
		DeviceID:      1,
		SubnetID:      1,
		CustomerID:    1,
	}

	mockPhysicalAssetFetcher := NewMockPhysicalAssetFetcher(ctrl)
	handler := FetchByIPAddressHandler{
		PhysicalAssetFetcher: mockPhysicalAssetFetcher,
		LogFn:                testLogFn,
	}

	mockPhysicalAssetFetcher.EXPECT().FetchPhysicalAsset(gomock.Any(), asset.IP).Return(asset, nil)
	response, err := handler.Handle(context.Background(), IPAddressQuery{IPAddress: asset.IP})
	require.Equal(t, nil, err)
	require.Equal(t, asset.IP, response.IP)
	require.Equal(t, asset.ResourceOwner, response.ResourceOwner)
	require.Equal(t, asset.Location, response.Tags.Location)
}
