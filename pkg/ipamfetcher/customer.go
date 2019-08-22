package ipamfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type customersResponse struct {
	Customers []customer `json:"Customers"`
}

type customer struct {
	CustomFields customFields `json:"Custom Fields"`
	ContactInfo  string       `json:"contact_info"`
	ID           int          `json:"id"`
}

// NewDevice42CustomerFetcher generates a new Device42CustomerFetcher
func NewDevice42CustomerFetcher(dc *Device42Client) *Device42CustomerFetcher {
	resourceEndpoint, _ := url.Parse(dc.Endpoint.String())
	resourceEndpoint.Path = path.Join(resourceEndpoint.Path, "api", "1.0", "customers")
	return &Device42CustomerFetcher{
		Client:   dc.Client,
		Endpoint: resourceEndpoint,
	}
}

// Device42CustomerFetcher fetches customer data from Device42
type Device42CustomerFetcher struct {
	Client   *http.Client
	Endpoint *url.URL
}

// FetchCustomers fetches customers from IPAM
func (d *Device42CustomerFetcher) FetchCustomers(ctx context.Context) ([]domain.Customer, error) {
	u, _ := url.Parse(d.Endpoint.String())
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := d.Client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error from device42 api: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var getCustomersResponse customersResponse
	if err := json.Unmarshal(body, &getCustomersResponse); err != nil {
		return nil, err
	}
	customers := make([]domain.Customer, 0, len(getCustomersResponse.Customers))
	for _, customer := range getCustomersResponse.Customers {
		customers = append(customers, domain.Customer{
			ID:            strconv.Itoa(customer.ID),
			ResourceOwner: customer.ContactInfo,
			BusinessUnit:  customer.CustomFields.GetValue("Description"),
		})
	}
	return customers, nil
}
