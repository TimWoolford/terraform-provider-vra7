package vrealize

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/terraform-provider-vra7/vrealize/api"
)

//ResourceMachine - use to set resource fields
func ResourceMachine() *schema.Resource {
	return &schema.Resource{
		Create: createResource,
		Read:   readResource,
		Update: updateResource,
		Delete: deleteResource,
		Schema: setResourceSchema(),
	}
}

//set_resource_schema - This function is used to update the catalog template/blueprint
//and replace the values with user defined values added in .tf file.
func setResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"catalog_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"catalog_id": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"wait_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  15,
		},
		"request_status": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
		"failed_message": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
			Optional: true,
		},
		"deployment_configuration": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},
		"resource_configuration": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},
		"catalog_configuration": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

//Function use - to create machine
//Terraform call - terraform apply
func changeTemplateValue(templateInterface map[string]interface{}, field string, value interface{}) map[string]interface{} {
	//Iterate over the map to get field provided as an argument
	for i := range templateInterface {
		//If value type is map then set recursive call which will fiend field in one level down of map interface
		if reflect.ValueOf(templateInterface[i]).Kind() == reflect.Map {
			template, _ := templateInterface[i].(map[string]interface{})
			templateInterface[i] = changeTemplateValue(template, field, value)
		} else if i == field {
			//If value type is not map then compare field name with provided field name
			//If both matches then update field value with provided value
			templateInterface[i] = value
		}
	}
	//Return updated map interface type
	return templateInterface
}

//Function use - to set a create resource call
//Terraform call - terraform apply
func createResource(d *schema.ResourceData, meta interface{}) error {
	//Log file handler to generate logs for debugging purpose
	//Get client handle
	client := meta.(*api.Client)

	//If catalog_name and catalog_id both not provided then throw an error
	if len(d.Get("catalog_name").(string)) <= 0 && len(d.Get("catalog_id").(string)) <= 0 {
		return fmt.Errorf("Either catalog_name or catalog_id should be present in given configuration")
	}

	//If catalog name is provided then get catalog ID using name for further process
	//else if catalog id is provided then fetch catalog name
	if len(d.Get("catalog_name").(string)) > 0 {
		catalogID, returnErr := client.ReadCatalogIDByName(d.Get("catalog_name").(string))
		log.Printf("createResource->catalog_id %v\n", catalogID)
		if returnErr != nil {
			return fmt.Errorf("%v", returnErr)
		}
		if catalogID == nil {
			return fmt.Errorf("No catalog found with name %v", d.Get("catalog_name").(string))
		} else if catalogID == "" {
			return fmt.Errorf("No catalog found with name %v", d.Get("catalog_name").(string))
		}
		d.Set("catalog_id", catalogID.(string))
	} else if len(d.Get("catalog_id").(string)) > 0 {
		CatalogName, nameError := client.ReadCatalogNameByID(d.Get("catalog_id").(string))
		if nameError != nil {
			return fmt.Errorf("%v", nameError)
		}
		if nameError != nil {
			d.Set("catalog_name", CatalogName.(string))
		}
	}
	//Get catalog blueprint
	templateCatalogItem, err := client.GetCatalogItem(d.Get("catalog_id").(string))
	log.Printf("createResource->templateCatalogItem %v\n", templateCatalogItem)

	catalogConfiguration, _ := d.Get("catalog_configuration").(map[string]interface{})
	for field1 := range catalogConfiguration {
		templateCatalogItem.Data[field1] = catalogConfiguration[field1]

	}
	log.Printf("createResource->templateCatalogItem.Data %v\n", templateCatalogItem.Data)

	//Get all resource keys from blueprint in array
	var keyList []string
	for field := range templateCatalogItem.Data {
		if reflect.ValueOf(templateCatalogItem.Data[field]).Kind() == reflect.Map {
			keyList = append(keyList, field)
		}
	}
	log.Printf("createResource->key_list %v\n", keyList)

	//Arrange keys in descending order of text length
	for field1 := range keyList {
		for field2 := range keyList {
			if len(keyList[field1]) > len(keyList[field2]) {
				temp := keyList[field1]
				keyList[field1], keyList[field2] = keyList[field2], temp
			}
		}
	}

	//Update template field values with user configuration
	resourceConfiguration, _ := d.Get("resource_configuration").(map[string]interface{})
	for configKey := range resourceConfiguration {
		for dataKey := range keyList {
			//compare resource list (resource_name) with user configuration fields (resource_name+field_name)
			if strings.Contains(configKey, keyList[dataKey]) {
				//If user_configuration contains resource_list element
				// then split user configuration key into resource_name and field_name
				splitedArray := strings.Split(configKey, keyList[dataKey]+".")
				//Function call which changes the template field values with  user values
				templateCatalogItem.Data[keyList[dataKey]] = changeTemplateValue(
					templateCatalogItem.Data[keyList[dataKey]].(map[string]interface{}),
					splitedArray[1],
					resourceConfiguration[configKey])
			}
		}
		//delete used user configuration
		delete(resourceConfiguration, configKey)
	}
	//update template with deployment level config
	// limit to description and reasons as other things could get us into trouble
	deploymentConfiguration, _ := d.Get("deployment_configuration").(map[string]interface{})
	for depField := range deploymentConfiguration {
		fieldstr := fmt.Sprintf("%s", depField)
		switch fieldstr {
		case "description":
			templateCatalogItem.Description = deploymentConfiguration[depField].(string)
		case "reasons":
			templateCatalogItem.Reasons = deploymentConfiguration[depField].(string)
		default:
			log.Printf("unknown option [%s] with value [%s] ignoring\n", depField, deploymentConfiguration[depField])
		}
	}
	//Log print of template after values updated
	log.Printf("Updated template - %v\n", templateCatalogItem.Data)

	//Through an exception if there is any error while getting catalog template
	if err != nil {
		return fmt.Errorf("Invalid CatalogItem ID %v", err)
	}

	//Set a  create machine function call
	requestMachine, err := client.RequestMachine(templateCatalogItem)

	//Check if error got while create machine call
	//If Error is occured, through an exception with an error message
	if err != nil {
		return fmt.Errorf("Resource Machine Request Failed: %v", err)
	}

	//Set request ID
	d.SetId(requestMachine.ID)
	//Set request status
	d.Set("request_status", "SUBMITTED")

	waitTimeout := d.Get("wait_timeout").(int) * 60

	for i := 0; i < waitTimeout/30; i++ {
		time.Sleep(3e+10)
		readResource(d, meta)

		if d.Get("request_status") == "SUCCESSFUL" {
			return nil
		}
		if d.Get("request_status") == "FAILED" {
			//If request is failed during the time then
			//unset resource details from state.
			d.SetId("")
			return fmt.Errorf("instance got failed while creating." +
				" kindly check detail for more information")
		}
	}
	if d.Get("request_status") == "IN_PROGRESS" {
		//If request is in_progress state during the time then
		//keep resource details in state files and throw an error
		//so that the child resource won't go for create call.
		//If execution gets timed-out and status is in progress
		//then dependent machine won't be get created in this iteration.
		//A user needs to ensure that the status should be a success state
		//using terraform refresh command and hit terraform apply again.
		return fmt.Errorf("resource is still being created")
	}

	return nil
}

//Function use - to update centOS 6.3 machine present in state file
//Terraform call - terraform refresh
func updateResource(d *schema.ResourceData, meta interface{}) error {
	log.Println(d)
	return nil
}

//Function use - To read configuration of centOS 6.3 machine present in state file
//Terraform call - terraform refresh
func readResource(d *schema.ResourceData, meta interface{}) error {
	//Get requester machine ID from schema.dataresource
	requestMachineID := d.Id()
	//Get client handle
	client := meta.(*api.Client)
	//Get requested status
	resourceTemplate, errTemplate := client.GetRequestStatus(requestMachineID)

	//Raise an exception if error occured while fetching request status
	if errTemplate != nil {
		return fmt.Errorf("Resource view failed to load:  %v", errTemplate)
	}

	//Update resource request status in state file
	d.Set("request_status", resourceTemplate.Phase)
	//If request is failed then set failed message in state file
	if resourceTemplate.Phase == "FAILED" {
		d.Set("failed_message", resourceTemplate.RequestCompletion.CompletionDetails)
	}
	return nil
}

//Function use - To delete resources which are created by terraform and present in state file
//Terraform call - terraform destroy
func deleteResource(d *schema.ResourceData, meta interface{}) error {
	//Get requester machine ID from schema.dataresource
	requestMachineID := d.Id()
	//Get client handle
	client := meta.(*api.Client)

	//Through an error if request ID has no value or empty value
	if len(d.Id()) == 0 {
		return fmt.Errorf("Resource not found")
	}
	//If resource create status is in_progress then skip delete call and through an exception
	if d.Get("request_status").(string) != "SUCCESSFUL" {
		if d.Get("request_status").(string) == "FAILED" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Machine cannot be deleted while in-progress state. Please try later")

	}
	//Fetch machine template
	templateResources, errTemplate := client.GetResourceViews(requestMachineID)

	if errTemplate != nil {
		return fmt.Errorf("Resource view failed to load:  %v", errTemplate)
	}

	//Set a delete machine template function call.
	//Which will fetch and return the delete machine template from the given template
	DestroyMachineTemplate, resourceTemplate, errDestroyAction := client.GetDestroyActionTemplate(templateResources)
	if errDestroyAction != nil {
		return fmt.Errorf("Destory Machine action template failed to load: %v", errDestroyAction)
	}
	//Set a destroy machine REST call
	_, errDestroyMachine := client.DestroyMachine(DestroyMachineTemplate, resourceTemplate)
	//Raise an exception if error got while deleting resource
	if errDestroyMachine != nil {
		return fmt.Errorf("Destory Machine machine operation failed: %v", errDestroyMachine)
	}
	//If resource got deleted then unset the resource ID from state file
	d.SetId("")
	return nil
}
