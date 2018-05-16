package api

import "fmt"

//DestroyMachine - To set resource destroy call
func (c *Client) DestroyMachine(destroyTemplate *ActionTemplate, resourceViewTemplate *ResourceViewsTemplate) (*ActionResponseTemplate, error) {
	//Get a destroy template URL from given resource template
	var destroyactionURL string
	destroyactionURL = getactionURL(resourceViewTemplate, "POST: {com.vmware.csp.component.cafe.composition@resource.action.deployment.destroy.name}")
	//Raise an error if any exception raised while fetching delete resource URL
	if len(destroyactionURL) == 0 {
		return nil, fmt.Errorf("Resource is not created or not found")
	}

	actionResponse := new(ActionResponseTemplate)
	apiError := new(Error)

	//Set a REST call with delete resource request and delete resource template as a data
	resp, err := c.HTTPClient.New().Post(destroyactionURL).
		BodyJSON(destroyTemplate).Receive(actionResponse, apiError)

	if resp.StatusCode != 201 {
		return nil, err
	}

	if !apiError.isEmpty() {
		return nil, apiError
	}

	return actionResponse, nil
}

//PowerOffMachine - To set resource power-off call
func (c *Client) PowerOffMachine(powerOffTemplate *ActionTemplate, resourceViewTemplate *ResourceViewsTemplate) (*ActionResponseTemplate, error) {
	//Get power-off resource URL from given template
	var powerOffMachineactionURL string
	powerOffMachineactionURL = getactionURL(resourceViewTemplate, "POST: {com.vmware.csp.component.iaas.proxy.provider@resource.action.name.machine.PowerOff}")
	//Raise an exception if error got while fetching URL
	if len(powerOffMachineactionURL) == 0 {
		return nil, fmt.Errorf("resource is not created or not found")
	}

	actionResponse := new(ActionResponseTemplate)
	apiError := new(Error)

	//Set a rest call to power-off the resource with resource power-off template as a data
	response, err := c.HTTPClient.New().Post(powerOffMachineactionURL).
		BodyJSON(powerOffTemplate).Receive(actionResponse, apiError)

	response.Close = true
	if response.StatusCode == 201 {
		return actionResponse, nil
	}

	if !apiError.isEmpty() {
		return nil, apiError
	}

	if err != nil {
		return nil, err
	}

	return nil, err
}
