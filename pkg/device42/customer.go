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

type customersResponse struct {
	Customers []customer `json:"Customers"`
}

type customer struct {
	Contacts     []contact    `json:"Contacts"`
	CustomFields customFields `json:"Custom Fields"`
	ContactInfo  string       `json:"contact_info"`
	DevicesURL   string       `json:"devices_url"`
	Groups       string       `json:"groups"`
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Notes        string       `json:"notes"`
	SubnetsURL   string       `json:"subnets_url"`
}

type contact struct {
	Address string `json:"address"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Type    string `json:"type"`
}

// Device42CustomerFetcher fetches customer data from Device42
type Device42CustomerFetcher struct {
	Client *http.Client
	Host   *url.URL
}

// FetchCustomers fetches customers from IPAM
func (d *Device42CustomerFetcher) FetchCustomers(ctx context.Context) ([]domain.Customer, error) {
	u, _ := url.Parse(d.Host.String())
	u.Path = path.Join("api", "1.0", "customers")
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := d.Client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
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
