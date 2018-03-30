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
	mu.Lock()
	defer mu.Unlock()

	scaleway := m.(*Client).scaleway

	vol, err := scaleway.GetVolume(d.Get("volume").(string))
	if err != nil {
		return err
	}
	if vol.Server != nil {
		log.Printf("[DEBUG] Scaleway volume %q already attached to %q.", vol.Identifier, vol.Server.Identifier)
		return errVolumeAlreadyAttached
	}

	serverID := d.Get("server").(string)

	server, err := scaleway.GetServer(serverID)
	if err != nil {
		fmt.Printf("Failed getting server: %q", err)
		return err
	}

	var startServerAgain = false
	// volumes can only be modified when the server is powered off
	if server.State != "stopped" {
		startServerAgain = true

		if _, err := scaleway.PostServerAction(server.Identifier, "poweroff"); err != nil {
			return err
		}
	}

	if err := waitForServerShutdown(scaleway, server.Identifier); err != nil {
		return err
	}

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

	if err := resource.Retry(serverWaitTimeout, func() *resource.RetryError {
		var req = api.ServerPatchDefinition{
			Volumes: &volumes,
		}
		err := scaleway.PatchServer(serverID, req)

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
	}); err != nil {
		return err
	}

	if startServerAgain {
		if _, err := scaleway.PostServerAction(serverID, "poweron"); err != nil {
			return err
		}
		if err := waitForServerStartup(scaleway, serverID); err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("scaleway-server:%s/volume/%s", serverID, d.Get("volume").(string)))

	return resourceScalewayVolumeAttachmentRead(d, m)
}

func resourceScalewayVolumeAttachmentRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

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
	mu.Lock()
	defer mu.Unlock()

	scaleway := m.(*Client).scaleway

	var startServerAgain = false

	serverID := d.Get("server").(string)

	server, err := scaleway.GetServer(serverID)
	if err != nil {
		return err
	}

	// volumes can only be modified when the server is powered off
	if server.State != "stopped" {
		startServerAgain = true
		if _, err := scaleway.PostServerAction(server.Identifier, "poweroff"); err != nil {
			return err
		}
	}
	if err := waitForServerShutdown(scaleway, server.Identifier); err != nil {
		return err
	}

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

	if err := resource.Retry(serverWaitTimeout, func() *resource.RetryError {
		var req = api.ServerPatchDefinition{
			Volumes: &volumes,
		}
		err := scaleway.PatchServer(serverID, req)

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
	}); err != nil {
		return err
	}

	if startServerAgain {
		if _, err := scaleway.PostServerAction(serverID, "poweron"); err != nil {
			return err
		}
		if err := waitForServerStartup(scaleway, serverID); err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}
