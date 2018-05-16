package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestAPIClient_GetCatalogItem(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	catalogId := "e5dd4fba-45ed-4943-b1fc-7f96239286be"
	path := fmt.Sprintf("http://localhost/catalog-service/api/consumer/entitledCatalogItems/%s/requests/template", catalogId)
	httpmock.RegisterResponder("GET", path,
		httpmock.NewStringResponder(200, testData("catalog_item_request_template")))

	template, err := client.GetCatalogItem(catalogId)

	assert.Nil(t, err)
	assert.Equal(t, catalogId, template.CatalogItemID)
}

func TestAPIClient_GetCatalogItem_Failure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/"+
		"api/consumer/entitledCatalogItems/e5dd4fba-45ed-4943-b1fc-7f96239286be/requests/template",
		httpmock.NewErrorResponder(errors.New(testData("api_error"))))

	template, err := client.GetCatalogItem("e5dd4fba-45ed-4943-b1fc-7f96239286be")

	assert.NotNil(t, err)
	assert.Nil(t, template)
}
