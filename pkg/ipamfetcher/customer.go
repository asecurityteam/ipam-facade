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
	Contacts     []Contact    `json:"Contacts"`
	CustomFields customFields `json:"custom_fields"`
	ContactInfo  string       `json:"contact_info"`
	ID           int          `json:"id"`
	Name         string       `json:"name"`
}

// Contact represents each contact object under the Customer object "Contacts" array.
type Contact struct {
	Type  string `json:"type"` // As of checking IPAM on Nov 25, 2019, can be any one of "Team Lead", "Technical", "Administrative", "Billing", "SRE"
	Email string `json:"email"`
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
		businessUnit := customer.CustomFields.GetValue("Description")
		if businessUnit == "" {
			// fallback to customer name
			businessUnit = customer.Name
		}
		customers = append(customers, domain.Customer{
			ID:            strconv.Itoa(customer.ID),
			ResourceOwner: getResourceOwner(customer),
			BusinessUnit:  businessUnit,
		})
	}
	return customers, nil
}

func getResourceOwner(customer customer) string {

	/*
	   Find the best resource owner by searching through the contacts
	   for non-empty Value for the following Key, in this priority order:
	   1. Team Lead Contact Email
	   2. Administrative Contact Email
	   3. SRE Contact Email
	   4. Technical Contact Email
	   5. Fall back to The "contact_info" field directly on the customer record
	*/

	owner := customer.ContactInfo
	fromPriority := 5
	for _, contact := range customer.Contacts {
		if contact.Type == "Team Lead" && contact.Email != "" && fromPriority > 1 {
			owner = contact.Email
			fromPriority = 1
		} else if contact.Type == "Administrative" && contact.Email != "" && fromPriority > 2 {
			owner = contact.Email
			fromPriority = 2
		} else if contact.Type == "SRE" && contact.Email != "" && fromPriority > 3 {
			owner = contact.Email
			fromPriority = 3
		} else if contact.Type == "Technical" && contact.Email != "" && fromPriority > 4 {
			owner = contact.Email
			fromPriority = 4
		}
	}

	return owner

}
