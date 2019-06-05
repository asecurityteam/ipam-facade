package ipamfetcher

import (
	"context"
	"errors"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFetchSubnets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 1, Offset: 0, Body: []byte(`{"offset": 0, "limit": 1, "total_count": 1, "subnets": [{"subnet_id": 1, "network": "192.168.1.1", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "AUS"}], "customer_id": 1}]}`)}, nil)

	d := &Device42SubnetFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	subnets, err := d.FetchSubnets(context.Background())
	assert.ElementsMatch(t, []domain.Subnet{domain.Subnet{ID: "1", Network: "192.168.1.1", MaskBits: int8(32), Location: "AUS", CustomerID: "1"}}, subnets)
	assert.Nil(t, err)
}

func TestFetchSubnetsMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 3, Offset: 0, Body: []byte(`{"offset": 0, "limit": 1, "total_count": 3, "subnets": [{"subnet_id": 1, "network": "192.168.1.1", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "AUS"}], "customer_id": 1}]}`)}, nil)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 1, 1).Return(PagedResponse{TotalCount: 3, Offset: 1, Body: []byte(`{"offset": 1, "limit": 1, "total_count": 3, "subnets": [{"subnet_id": 2, "network": "192.168.1.0", "mask_bits": 28, "custom_fields": [{"key": "Location", "value": "SYD"}], "customer_id": 2}]}`)}, nil)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 2, 1).Return(PagedResponse{TotalCount: 3, Offset: 2, Body: []byte(`{"offset": 2, "limit": 1, "total_count": 3, "subnets": [{"subnet_id": 3, "network": "192.168.1.3", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "LON"}], "customer_id": 3}]}`)}, nil)

	d := &Device42SubnetFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	subnets, err := d.FetchSubnets(context.Background())
	assert.ElementsMatch(t, []domain.Subnet{
		domain.Subnet{ID: "1", Network: "192.168.1.1", MaskBits: int8(32), Location: "AUS", CustomerID: "1"},
		domain.Subnet{ID: "2", Network: "192.168.1.0", MaskBits: int8(28), Location: "SYD", CustomerID: "2"},
		domain.Subnet{ID: "3", Network: "192.168.1.3", MaskBits: int8(32), Location: "LON", CustomerID: "3"},
	}, subnets)
	assert.Nil(t, err)
}

func TestFetchSubnetsUnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 1, Offset: 0, Body: []byte("notasubnet")}, nil)

	d := &Device42SubnetFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	_, err := d.FetchSubnets(context.Background())
	assert.NotNil(t, err)
}

func TestFetchSubnetsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 1, Offset: 0, Body: []byte("{}")}, errors.New("request err"))

	d := &Device42SubnetFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	_, err := d.FetchSubnets(context.Background())
	assert.NotNil(t, err)
}
