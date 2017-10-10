package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/dataplane/keyvault"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmKeyVaultSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmKeyVaultSecretCreate,
		Read:   resourceArmKeyVaultSecretRead,
		Update: resourceArmKeyVaultSecretUpdate,
		Delete: resourceArmKeyVaultSecretDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateKeyVaultChildName,
			},

			"vault_uri": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmKeyVaultSecretCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).keyVaultManagementClient

	log.Print("[INFO] preparing arguments for AzureRM KeyVault Secret creation.")

	name := d.Get("name").(string)
	keyVaultBaseUrl := d.Get("vault_uri").(string)
	value := d.Get("value").(string)
	contentType := d.Get("content_type").(string)
	tags := d.Get("tags").(map[string]interface{})

	parameters := keyvault.SecretSetParameters{
		Value:       utils.String(value),
		ContentType: utils.String(contentType),
		Tags:        expandTags(tags),
	}

	_, err := client.SetSecret(keyVaultBaseUrl, name, parameters)
	if err != nil {
		return err
	}

	// "" indicates the latest version
	read, err := client.GetSecret(keyVaultBaseUrl, name, "")
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read KeyVault Secret '%s' (in key vault '%s')", name, keyVaultBaseUrl)
	}

	d.SetId(*read.ID)

	return resourceArmKeyVaultSecretRead(d, meta)
}

func resourceArmKeyVaultSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).keyVaultManagementClient
	log.Print("[INFO] preparing arguments for AzureRM KeyVault Secret update.")

	id, err := parseKeyVaultChildID(d.Id())
	if err != nil {
		return err
	}

	value := d.Get("value").(string)
	contentType := d.Get("content_type").(string)
	tags := d.Get("tags").(map[string]interface{})

	if d.HasChange("value") {
		// for changing the value of the secret we need to create a new version
		parameters := keyvault.SecretSetParameters{
			Value:       utils.String(value),
			ContentType: utils.String(contentType),
			Tags:        expandTags(tags),
		}

		_, err := client.SetSecret(id.KeyVaultBaseUrl, id.Name, parameters)
		if err != nil {
			return err
		}

		// "" indicates the latest version
		read, err := client.GetSecret(id.KeyVaultBaseUrl, id.Name, "")
		id, err = parseKeyVaultChildID(*read.ID)
		if err != nil {
			return err
		}

		// the ID is suffixed with the secret version
		d.SetId(*read.ID)
	} else {
		parameters := keyvault.SecretUpdateParameters{
			ContentType: utils.String(contentType),
			Tags:        expandTags(tags),
		}

		_, err = client.UpdateSecret(id.KeyVaultBaseUrl, id.Name, id.Version, parameters)
		if err != nil {
			return err
		}
	}

	return resourceArmKeyVaultSecretRead(d, meta)
}

func resourceArmKeyVaultSecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).keyVaultManagementClient

	id, err := parseKeyVaultChildID(d.Id())
	if err != nil {
		return err
	}

	// we always want to get the latest version
	resp, err := client.GetSecret(id.KeyVaultBaseUrl, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure KeyVault Secret %s: %+v", id.Name, err)
	}

	// the version may have changed, so parse the updated id
	respID, err := parseKeyVaultChildID(*resp.ID)
	if err != nil {
		return err
	}

	d.Set("name", respID.Name)
	d.Set("vault_uri", respID.KeyVaultBaseUrl)
	d.Set("value", resp.Value)
	d.Set("version", respID.Version)
	d.Set("content_type", resp.ContentType)

	flattenAndSetTags(d, resp.Tags)
	return nil
}

func resourceArmKeyVaultSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).keyVaultManagementClient

	id, err := parseKeyVaultChildID(d.Id())
	if err != nil {
		return err
	}

	_, err = client.DeleteSecret(id.KeyVaultBaseUrl, id.Name)

	return err
}
