package api

import "fmt"

//GetRequestStatus - To read request status of resource
// which is used to show information to user post create call.
func (c *Client) GetRequestStatus(ResourceID string) (*RequestStatusView, error) {
	//Form a URL to read request status
	path := fmt.Sprintf("catalog-service/api/consumer/requests/%s", ResourceID)
	RequestStatusViewTemplate := new(RequestStatusView)
	apiError := new(Error)
	//Set a REST call and fetch a resource request status
	_, err := c.HTTPClient.New().Get(path).Receive(RequestStatusViewTemplate, apiError)
	if err != nil {
		return nil, err
	}
	if !apiError.isEmpty() {
		return nil, apiError
	}
	return RequestStatusViewTemplate, nil
}

//GetResourceViews - To read resource configuration
func (c *Client) GetResourceViews(ResourceID string) (*ResourceViewsTemplate, error) {
	//Form an URL to fetch resource list view
	path := fmt.Sprintf("catalog-service/api/consumer/requests/%s"+
		"/resourceViews", ResourceID)
	resourceViewsTemplate := new(ResourceViewsTemplate)
	apiError := new(Error)
	//Set a REST call to fetch resource view data
	_, err := c.HTTPClient.New().Get(path).Receive(resourceViewsTemplate, apiError)
	if err != nil {
		return nil, err
	}
	if !apiError.isEmpty() {
		return nil, apiError
	}
	return resourceViewsTemplate, nil
}
