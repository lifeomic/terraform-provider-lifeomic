package appstore

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/lifeomic/phc-sdk-go/client"
)

func readApplet(d *schema.ResourceData, meta interface{}) error {
	appstoreClient := meta.(*client.LambdaClient).AppStore()
	id := d.Id()
	app, err := appstoreClient.GetAppStoreListing(id)
	if err != nil {
		return err
	}
	d.Set("name", app.Name)
	d.Set("description", app.Description)
	d.Set("author_display", app.AuthorDisplay)
	d.Set("image", app.Image)
	d.Set("url", app.Url)
	return nil
}

func createApplet(d *schema.ResourceData, meta interface{}) error {
	appstoreClient := meta.(*client.LambdaClient).AppStore()
	id, err := appstoreClient.CreateAppStoreListing(client.AppStoreCreate{
		Name:          d.Get("name").(string),
		AuthorDisplay: d.Get("author_display").(string),
		Image:         d.Get("image").(string),
		Url:           d.Get("url").(string),
		Description:   d.Get("description").(string),
	})
	if err != nil {
		return err
	}
	d.SetId(*id)
	return nil
}

func updateApplet(d *schema.ResourceData, meta interface{}) error {
	appstoreClient := meta.(*client.LambdaClient).AppStore()
	err := appstoreClient.EditAppStoreListing(d.Id(), client.AppStoreCreate{
		Name:          d.Get("name").(string),
		AuthorDisplay: d.Get("author_display").(string),
		Image:         d.Get("image").(string),
		Url:           d.Get("url").(string),
		Description:   d.Get("description").(string),
	})
	return err
}

func deleteApplet(d *schema.ResourceData, meta interface{}) error {
	appstoreClient := meta.(*client.LambdaClient).AppStore()
	return appstoreClient.DeleteAppStoreListing(d.Id())
}

func AppletResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"author_display": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		Create: createApplet,
		Read:   readApplet,
		Update: updateApplet,
		Delete: deleteApplet,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}
