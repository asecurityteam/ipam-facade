package ipamfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/settings"
)

// config contains all configuration values for creating a system logger.
type config struct {
	TypeSearchOrder string `description:"Comma-delimited priority list of 'type' value in the IPAM Contacts objects in which to search for non-empty 'email' to use as the resource owner.  If no values are found, the logic falls back to using Customer.contact_info"`
}

// Name of the configuration as it might appear in config files.
func (*config) Name() string {
	return "Contact"
}

// component enables creating configured logic.
type component struct{}

// settings generates a Config with default values applied.
func (*component) Settings() *config {
	return &config{}
}

// New creates a function that loads the config from the designated source
func (*component) New(_ context.Context, c *config) (*searchOrder, error) {
	return &searchOrder{keys: strings.Split(c.TypeSearchOrder, ",")}, nil
}

type searchOrder struct {
	keys []string
}

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
	Type  string `json:"type"`
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
	   for non-empty Value for the following Key, in the priority order
	   specified by the CONTACT_TYPESEARCHORDER environment variable, with a
	   fall back to The "contact_info" field.
	*/

	searchOrder := new(searchOrder)
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	err = settings.NewComponent(context.Background(), source, &component{}, searchOrder)
	if err != nil {
		panic(err.Error())
	}

	highestPriorityFound := len(searchOrder.keys)

	owner := customer.ContactInfo

	for _, contact := range customer.Contacts {
		keyIndex := keyIndex(searchOrder.keys, contact.Type)
		if keyIndex > -1 && keyIndex < highestPriorityFound && contact.Email != "" {
			// this is one of the contacts we're looking for, and it's higher priority than any others we've found so far
			owner = contact.Email
			highestPriorityFound = keyIndex
		}
	}

	return owner

}

func keyIndex(haystack []string, needle string) int {
	keyIndex := -1
	for _, straw := range haystack {
		keyIndex++
		if straw == needle {
			return keyIndex
		}
	}
	return -1
}
