package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

var client Client

func init() {
	client = NewClient(
		"admin@myvra.local",
		"pass!@#",
		"vsphere.local",
		"http://localhost/",
		true,
	)
}

func TestNewClient(t *testing.T) {
	username := "admin@myvra.local"
	password := "pass!@#"
	tenant := "vshpere.local"
	baseURL := "http://localhost/"

	client := NewClient(
		username,
		password,
		tenant,
		baseURL,
		true,
	)

	assert.Equal(t, username, client.Username)
	assert.Equal(t, password, client.Password)
	assert.Equal(t, tenant, client.Tenant)
	assert.Equal(t, baseURL, client.BaseURL)
}

func TestClient_Authenticate_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost/identity/api/tokens",
		httpmock.NewStringResponder(200, testData("token")))

	err := client.Authenticate()

	assert.Nil(t, err)
	assert.True(t, len(client.BearerToken) > 0)
}

func TestClient_Authenticate_Failure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost/identity/api/tokens",
		httpmock.NewErrorResponder(errors.New(testData("token_error"))))

	err := client.Authenticate()

	assert.NotNil(t, err)
}

func TestClient_GetResourceViews_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/"+
		"api/consumer/requests/937099db-5174-4862-99a3-9c2666bfca28/resourceViews",
		httpmock.NewStringResponder(200, testData("request_resource_views")))

	template, err := client.GetResourceViews("937099db-5174-4862-99a3-9c2666bfca28")
	if err != nil {
		t.Errorf("Fail to get resource views %v.", err)
	}
	if len(template.Content) == 0 {
		t.Errorf("No resources provisioned.")
	}
}

func TestClient_GetResourceViews_Failure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/"+
		"api/consumer/requests/937099db-5174-4862-99a3-9c2666bfca28/resourceViews",
		httpmock.NewErrorResponder(errors.New(testData("api_error"))))

	template, err := client.GetResourceViews("937099db-5174-4862-99a3-9c2666bfca28")
	if err == nil {
		t.Errorf("Succeed to get resource views %v.", err)
	}
	if template != nil {
		t.Errorf("Resources provisioned.")
	}

}

func TestClient_GetDestroyActionTemplate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/"+
		"api/consumer/requests/937099db-5174-4862-99a3-9c2666bfca28/resourceViews",
		httpmock.NewStringResponder(200, testData("request_resource_views")))

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/api/consumer/resources/b313acd6-0738-439c-b601-e3ebf9ebb49b/actions/3da0ca14-e7e2-4d7b-89cb-c6db57440d72/requests/template",
		httpmock.NewStringResponder(200, `{"type":"com.vmware.vcac.catalog.domain.request.CatalogResourceRequest","resourceId":"b313acd6-0738-439c-b601-e3ebf9ebb49b","actionId":"3da0ca14-e7e2-4d7b-89cb-c6db57440d72","description":null,"data":{"ForceDestroy":false}}`))

	templateResources, errTemplate := client.GetResourceViews("937099db-5174-4862-99a3-9c2666bfca28")
	if errTemplate != nil {
		t.Errorf("Failed to get the template resources %v", errTemplate)
	}

	_, _, err := client.GetDestroyActionTemplate(templateResources)

	if err != nil {
		t.Errorf("Fail to get destroy action template %v", err)
	}

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/api/consumer/resources/b313acd6-0738-439c-b601-e3ebf9ebb49b/actions/3da0ca14-e7e2-4d7b-89cb-c6db57440d72/requests/template",
		httpmock.NewErrorResponder(errors.New(`{"errors":[{"code":50505,"source":null,"message":"System exception.","systemMessage":null,"moreInfoUrl":null}]}`)))

	_, _, err = client.GetDestroyActionTemplate(templateResources)

	if err == nil {
		t.Errorf("Fail to get destroy action template exception.")
	}
}

func TestClient_destroyMachine(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/"+
		"api/consumer/requests/937099db-5174-4862-99a3-9c2666bfca28/resourceViews",
		httpmock.NewStringResponder(200, testData("request_resource_views")))

	httpmock.RegisterResponder("GET", "http://localhost/catalog-service/api/consumer/resources/b313acd6-0738-439c-b601-e3ebf9ebb49b/actions/3da0ca14-e7e2-4d7b-89cb-c6db57440d72/requests/template",

		httpmock.NewStringResponder(200, `{"type":"com.vmware.vcac.catalog.domain.request.CatalogResourceRequest","resourceId":"b313acd6-0738-439c-b601-e3ebf9ebb49b","actionId":"3da0ca14-e7e2-4d7b-89cb-c6db57440d72","description":null,"data":{"ForceDestroy":false}}`))

	httpmock.RegisterResponder("POST", "http://localhost/catalog-service/api/consumer/resources/b313acd6-0738-439c-b601-e3ebf9ebb49b/actions/3da0ca14-e7e2-4d7b-89cb-c6db57440d72/requests",
		httpmock.NewStringResponder(200, ``))

	templateResources, errTemplate := client.GetResourceViews("937099db-5174-4862-99a3-9c2666bfca28")
	if errTemplate != nil {
		t.Errorf("Failed to get the template resources %v", errTemplate)
	}
	destroyActionTemplate, resourceTemplate, err := client.GetDestroyActionTemplate(templateResources)

	if err != nil {
		t.Errorf("Failed to get destroy action template %v", err)
	}
	client.DestroyMachine(destroyActionTemplate, resourceTemplate)
}

func testData(name string) string {
	bytes, readErr := ioutil.ReadFile(fmt.Sprintf("./test_data/%s.json", name))
	if readErr != nil {
		panic(readErr)
	}
	return string(bytes)
}
