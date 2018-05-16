package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestAPIClient_RequestMachine_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	template := new(CatalogItemTemplate)
	bytes, _ := ioutil.ReadFile("./test_data/catalog_item_request_template.json")
	json.Unmarshal(bytes, template)

	catalogId := "e5dd4fba-45ed-4943-b1fc-7f96239286be"
	path := fmt.Sprintf("http://localhost/catalog-service/api/consumer/entitledCatalogItems/%s/requests", catalogId)
	httpmock.RegisterResponder("POST", path,
		httpmock.NewStringResponder(201, testData("request")))

	requestMachine, errorRequestMachine := client.RequestMachine(template)

	assert.Nil(t, errorRequestMachine)

	if len(requestMachine.ID) == 0 {
		t.Errorf("Failed to request machine.")
	}
}

func TestAPIClient_RequestMachine_Failure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	template := new(CatalogItemTemplate)
	bytes, _ := ioutil.ReadFile("./test_data/catalog_item_request_template.json")
	json.Unmarshal(bytes, template)

	catalogId := "e5dd4fba-45ed-4943-b1fc-7f96239286be"
	path := fmt.Sprintf("http://localhost/catalog-service/api/consumer/entitledCatalogItems/%s/requests", catalogId)
	httpmock.RegisterResponder("POST", path,
		httpmock.NewErrorResponder(errors.New(testData("api_error"))))

	requestMachine, errorRequestMachine := client.RequestMachine(template)

	if errorRequestMachine == nil {
		t.Errorf("Failed to generate exception.")
	}

	if requestMachine != nil {
		t.Errorf("Deploy machine request succeeded.")
	}
}
