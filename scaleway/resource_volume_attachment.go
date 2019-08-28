package scaleway

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.
 Please use scaleway_instance_server.additional_volumes instead.`,

		Create: resourceScalewayVolumeAttachmentCreate,
		Read:   resourceScalewayVolumeAttachmentRead,
		Delete: resourceScalewayVolumeAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "the server a volume should be attached to",
			},
			"volume": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "the volume to attach",
			},
		},
	}
}

var errVolumeAlreadyAttached = fmt.Errorf("Scaleway volume already attached")

func resourceScalewayVolumeAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	vol, err := scaleway.GetVolume(d.Get("volume").(string))

	if err != nil {
		return err
	}
	if vol.Server != nil {
		log.Printf("[DEBUG] Scaleway volume %q already attached to %q.", vol.Identifier, vol.Server.Identifier)
		return errVolumeAlreadyAttached
	}

	if err := withStoppedServer(scaleway, d.Get("server").(string), func(server *api.Server) error {
		volumes := make(map[string]api.Volume)
		for i, volume := range server.Volumes {
			volumes[i] = volume
		}

		volumes[fmt.Sprintf("%d", len(volumes)+1)] = *vol

		// the API request requires most volume attributes to be unset to succeed
		for k, v := range volumes {
			v.Size = 0
			v.CreationDate = ""
			v.Organization = ""
			v.ModificationDate = ""
			v.VolumeType = ""
			v.Server = nil
			v.ExportURI = ""

			volumes[k] = v
		}

		return resource.Retry(serverWaitTimeout, func() *resource.RetryError {
			var req = api.ServerPatchDefinition{
				Volumes: &volumes,
			}
			err := scaleway.PatchServer(server.Identifier, req)

			if err == nil {
				return nil
			}

			if serr, ok := err.(api.APIError); ok {
				log.Printf("[DEBUG] Error patching server: %q\n", serr.APIMessage)

				if serr.StatusCode == 400 {
					return resource.RetryableError(fmt.Errorf("Waiting for server update to succeed: %q", serr.APIMessage))
				}
			}

			return resource.NonRetryableError(err)
		})
	}); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("scaleway-server:%s/volume/%s", d.Get("server").(string), d.Get("volume").(string)))

	return resourceScalewayVolumeAttachmentRead(d, m)
}

func resourceScalewayVolumeAttachmentRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	server, err := scaleway.GetServer(d.Get("server").(string))
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading server: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if _, err := scaleway.GetVolume(d.Get("volume").(string)); err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading volume: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	for _, volume := range server.Volumes {
		if volume.Identifier == d.Get("volume").(string) {
			return nil
		}
	}

	log.Printf("[DEBUG] Volume %q not attached to server %q\n", d.Get("volume").(string), d.Get("server").(string))
	d.SetId("")
	return nil
}

const serverWaitTimeout = 5 * time.Minute

func resourceScalewayVolumeAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	if err := withStoppedServer(scaleway, d.Get("server").(string), func(server *api.Server) error {
		volumes := make(map[string]api.Volume)
		for _, volume := range server.Volumes {
			if volume.Identifier != d.Get("volume").(string) {
				volumes[fmt.Sprintf("%d", len(volumes))] = volume
			}
		}

		// the API request requires most volume attributes to be unset to succeed
		for k, v := range volumes {
			v.Size = 0
			v.CreationDate = ""
			v.Organization = ""
			v.ModificationDate = ""
			v.VolumeType = ""
			v.Server = nil
			v.ExportURI = ""

			volumes[k] = v
		}

		return resource.Retry(serverWaitTimeout, func() *resource.RetryError {
			var req = api.ServerPatchDefinition{
				Volumes: &volumes,
			}
			err := scaleway.PatchServer(server.Identifier, req)

			if err == nil {
				return nil
			}

			if serr, ok := err.(api.APIError); ok {
				log.Printf("[DEBUG] Error patching server: %q\n", serr.APIMessage)

				if serr.StatusCode == 400 {
					return resource.RetryableError(fmt.Errorf("Waiting for server update to succeed: %q", serr.APIMessage))
				}
			}

			return resource.NonRetryableError(err)
		})
	}); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
