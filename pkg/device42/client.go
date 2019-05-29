package device42

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type Client struct {
	Client *http.Client
	Host   *url.URL
	Limit  int
}

func (c *Client) GetSubnets(ctx context.Context) ([]domain.Subnet, error) {
	getSubnetsResponse, err := c.makePagedSubnetsRequest(ctx, 0)
	if err != nil {
		return nil, err
	}

	totalSubnets := getSubnetsResponse.TotalCount
	subnetsChan := make(chan subnet, totalSubnets)
	errChan := make(chan error, 1)
	for _, subnet := range getSubnetsResponse.Subnets {
		subnetsChan <- subnet
	}

	for offset := c.Limit; offset <= (totalSubnets - c.Limit); offset = offset + c.Limit {
		go func(offset int) {
			getSubnetsResponse, err := c.makePagedSubnetsRequest(ctx, offset)
			if err != nil {
				errChan <- err
				return
			}

			for _, subnet := range getSubnetsResponse.Subnets {
				subnetsChan <- subnet
			}
		}(offset)
	}

	subnets := make([]domain.Subnet, 0, totalSubnets)
	for i := 0; i < totalSubnets; i = i + 1 {
		select {
		case s := <-subnetsChan:
			subnets = append(subnets, domain.Subnet{
				ID:       strconv.Itoa(s.SubnetID),
				Network:  s.Network,
				MaskBits: int8(s.MaskBits),
				Location: "",
				// Customer: s.CustomerID,
			})
		case err := <-errChan:
			return nil, err
		}
	}

	return subnets, nil
}

func (c *Client) makePagedSubnetsRequest(ctx context.Context, offset int) (subnetResponse, error) {
	u, _ := url.Parse(c.Host.String())
	u.Path = path.Join("api", "1.0", "subnets")
	q := u.Query()
	q.Set("limit", strconv.Itoa(c.Limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := c.Client.Do(req.WithContext(ctx))
	if err != nil {
		return subnetResponse{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return subnetResponse{}, err
	}
	var getSubnetsResponse subnetResponse
	err = json.Unmarshal(body, &getSubnetsResponse)
	if err != nil {
		return subnetResponse{}, err
	}
	return getSubnetsResponse, nil
}
