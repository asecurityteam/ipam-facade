package device42

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubnetLocation(t *testing.T) {
	tc := []struct {
		name             string
		customFields     customFields
		expectedLocation string
	}{
		{
			"success",
			[]customField{customField{Key: "Location", Value: "SYD"}},
			"SYD",
		},
		{
			"null value",
			[]customField{customField{Key: "Location", Value: nil}},
			"",
		},
		{
			"no location",
			[]customField{customField{Key: "NotLocation", Value: nil}},
			"",
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			location := test.customFields.GetValue("Location")
			assert.Equal(tt, test.expectedLocation, location)
		})
	}
}
