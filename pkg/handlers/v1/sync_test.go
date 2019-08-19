package v1

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

func TestSyncHandlerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ipamData := domain.IPAMData{
		Devices: []domain.Device{
			domain.Device{
				ID:       "1",
				IP:       "127.0.0.1",
				SubnetID: "1",
			},
		},
		Subnets: []domain.Subnet{
			domain.Subnet{
				ID:         "1",
				Network:    "127.0.0.0/31",
				MaskBits:   1,
				Location:   "",
				CustomerID: "1",
			},
		},
		Customers: []domain.Customer{
			domain.Customer{
				ID:            "1",
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Security",
			},
		},
	}

	mockIPAMDataFetcher := NewMockIPAMDataFetcher(ctrl)
	mockAssetStorer := NewMockPhysicalAssetStorer(ctrl)
	handler := SyncIPAMDataHandler{
		IPAMDataFetcher:     mockIPAMDataFetcher,
		PhysicalAssetStorer: mockAssetStorer,
		LogFn:               testLogFn,
	}

	mockIPAMDataFetcher.EXPECT().FetchIPAMData(gomock.Any()).Return(ipamData, nil)
	mockAssetStorer.EXPECT().StorePhysicalAssets(gomock.Any(), ipamData).Return(nil)
	err := handler.Handle(context.Background(), JobMetadata{JobID: "foo-bar-baz-quux"})
	require.Equal(t, nil, err)
}

func TestSyncHandlerIPAMDataFetchFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIPAMDataFetcher := NewMockIPAMDataFetcher(ctrl)
	mockAssetStorer := NewMockPhysicalAssetStorer(ctrl)
	handler := SyncIPAMDataHandler{
		IPAMDataFetcher:     mockIPAMDataFetcher,
		PhysicalAssetStorer: mockAssetStorer,
		LogFn:               testLogFn,
	}

	mockIPAMDataFetcher.EXPECT().FetchIPAMData(gomock.Any()).Return(domain.IPAMData{}, errors.New("boom"))
	err := handler.Handle(context.Background(), JobMetadata{JobID: "foo-bar-baz-quux"})
	require.Equal(t, errors.New("boom"), err)
}

func TestSyncHandlerIPAMDataStorerFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ipamData := domain.IPAMData{
		Devices: []domain.Device{
			domain.Device{
				ID:       "1",
				IP:       "127.0.0.1",
				SubnetID: "1",
			},
		},
		Subnets: []domain.Subnet{
			domain.Subnet{
				ID:         "1",
				Network:    "127.0.0.0/31",
				MaskBits:   1,
				Location:   "",
				CustomerID: "1",
			},
		},
		Customers: []domain.Customer{
			domain.Customer{
				ID:            "1",
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Security",
			},
		},
	}

	mockIPAMDataFetcher := NewMockIPAMDataFetcher(ctrl)
	mockAssetStorer := NewMockPhysicalAssetStorer(ctrl)
	handler := SyncIPAMDataHandler{
		IPAMDataFetcher:     mockIPAMDataFetcher,
		PhysicalAssetStorer: mockAssetStorer,
		LogFn:               testLogFn,
	}

	mockIPAMDataFetcher.EXPECT().FetchIPAMData(gomock.Any()).Return(ipamData, nil)
	mockAssetStorer.EXPECT().StorePhysicalAssets(gomock.Any(), ipamData).Return(errors.New("boom"))
	err := handler.Handle(context.Background(), JobMetadata{JobID: "foo-bar-baz-quux"})
	require.Equal(t, errors.New("boom"), err)
}
