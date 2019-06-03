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
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte(`{"subnets": [{"subnet_id": 1, "network": "192.168.1.1", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "AUS"}], "customer_id": 1}, {"subnet_id": 2, "network": "192.168.1.0", "mask_bits": 28, "custom_fields": [{"key": "Location", "value": "SYD"}], "customer_id": 2}]}`)})
	mockIterator.EXPECT().Next().Return(false)
	mockIterator.EXPECT().Close().Return(nil)

	d := &Device42SubnetFetcher{
		Iterator: mockIterator,
	}

	subnets, err := d.FetchSubnets(context.Background())
	assert.ElementsMatch(t, []domain.Subnet{domain.Subnet{ID: "1", Network: "192.168.1.1", MaskBits: int8(32), Location: "AUS", CustomerID: "1"}, domain.Subnet{ID: "2", Network: "192.168.1.0", MaskBits: int8(28), Location: "SYD", CustomerID: "2"}}, subnets)
	assert.Nil(t, err)
}
func TestFetchSubnetsMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte(`{"subnets": [{"subnet_id": 1, "network": "192.168.1.1", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "AUS"}], "customer_id": 1}, {"subnet_id": 2, "network": "192.168.1.0", "mask_bits": 28, "custom_fields": [{"key": "Location", "value": "SYD"}], "customer_id": 2}]}`)})
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte(`{"subnets": [{"subnet_id": 3, "network": "192.168.1.3", "mask_bits": 32, "custom_fields": [{"key": "Location", "value": "LON"}], "customer_id": 3}]}`)})
	mockIterator.EXPECT().Next().Return(false)
	mockIterator.EXPECT().Close().Return(nil)

	d := &Device42SubnetFetcher{
		Iterator: mockIterator,
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
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte("notasubnet")})

	d := &Device42SubnetFetcher{
		Iterator: mockIterator,
	}

	_, err := d.FetchSubnets(context.Background())
	assert.NotNil(t, err)
}

func TestFetchSubnetsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(false)
	mockIterator.EXPECT().Close().Return(errors.New("iterator error"))

	d := &Device42SubnetFetcher{
		Iterator: mockIterator,
	}

	_, err := d.FetchSubnets(context.Background())
	assert.NotNil(t, err)
}
