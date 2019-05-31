package device42

import (
	"context"
	"fmt"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFetchSubnets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPaginator := NewMockPaginator(ctrl)
	mockPaginator.EXPECT().BatchPagedRequests(gomock.Any()).Return(
		[][]byte{[]byte(`{"subnets": [{"subnet_id": 1, "network": "192.168.1.1", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "AUS"}], "customer_id": 1}, {"subnet_id": 2, "network": "192.168.1.0", "mask_bits": 28, "custom_fields": [{"key": "Location", "value": "SYD"}], "customer_id": 2}]}`)},
		nil,
	)
	d := &Device42SubnetFetcher{
		Paginator: mockPaginator,
	}

	subnets, err := d.FetchSubnets(context.Background())
	assert.ElementsMatch(t, []domain.Subnet{domain.Subnet{ID: "1", Network: "192.168.1.1", MaskBits: int8(32), Location: "AUS", CustomerID: "1"}, domain.Subnet{ID: "2", Network: "192.168.1.0", MaskBits: int8(28), Location: "SYD", CustomerID: "2"}}, subnets)
	assert.Nil(t, err)
}

func TestFetchSubnetsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPaginator := NewMockPaginator(ctrl)
	mockPaginator.EXPECT().BatchPagedRequests(gomock.Any()).Return(nil, fmt.Errorf("batch error"))
	d := &Device42SubnetFetcher{
		Paginator: mockPaginator,
	}

	_, err := d.FetchSubnets(context.Background())
	assert.NotNil(t, err)
}

func TestFetchSubnetsUnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPaginator := NewMockPaginator(ctrl)
	mockPaginator.EXPECT().BatchPagedRequests(gomock.Any()).Return([][]byte{[]byte("notasubnet")}, nil)
	d := &Device42SubnetFetcher{
		Paginator: mockPaginator,
	}

	_, err := d.FetchSubnets(context.Background())
	assert.NotNil(t, err)
}
