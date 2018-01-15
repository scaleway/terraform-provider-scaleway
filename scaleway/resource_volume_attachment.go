package scaleway

import (
	"fmt"
	"log"
	"time"

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

	var (
		vol *api.Volume
		err error
	)
	if err = retry(func() error {
		vol, err = scaleway.GetVolume(d.Get("volume").(string))
		return err
	}); err != nil {
		return err
	}
	if vol.Server != nil {
		log.Printf("[DEBUG] Scaleway volume %q already attached to %q.", vol.Identifier, vol.Server.Identifier)
		return errVolumeAlreadyAttached
	}

	serverID := d.Get("server").(string)

	var server *api.Server
	if err = retry(func() error {
		server, err = scaleway.GetServer(serverID)
		return err
	}); err != nil {
		fmt.Printf("Failed getting server: %q", err)
		return err
	}

	var startServerAgain = false
	// volumes can only be modified when the server is powered off
	if server.State != "stopped" {
		startServerAgain = true

		if err = retry(func() error {
			return scaleway.PostServerAction(server.Identifier, "poweroff")
		}); err != nil {
			return err
		}
	}
	if err := waitForServerState(scaleway, server.Identifier, "stopped"); err != nil {
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

	if err := retryWithCodes([]int{429, 400}, func() error {
		var req = api.ServerPatchDefinition{
			Volumes: &volumes,
		}
		return scaleway.PatchServer(serverID, req)
	}); err != nil {
		return err
	}

	if startServerAgain {
		if err := retry(func() error {
			return scaleway.PostServerAction(serverID, "poweron")
		}); err != nil {
			return err
		}
		if err := waitForServerState(scaleway, serverID, "running"); err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("scaleway-server:%s/volume/%s", serverID, d.Get("volume").(string)))

	return resourceScalewayVolumeAttachmentRead(d, m)
}

func resourceScalewayVolumeAttachmentRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	var (
		server *api.Server
		err    error
	)
	if err = retry(func() error {
		server, err = scaleway.GetServer(d.Get("server").(string))
		return err
	}); err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading server: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err := retry(func() error {
		_, err := scaleway.GetVolume(d.Get("volume").(string))
		return err
	}); err != nil {
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

	var (
		server *api.Server
		err    error
	)
	if err := retry(func() error {
		server, err = scaleway.GetServer(serverID)
		return err
	}); err != nil {
		return err
	}

	// volumes can only be modified when the server is powered off
	if server.State != "stopped" {
		startServerAgain = true
		if err := retry(func() error {
			return scaleway.PostServerAction(server.Identifier, "poweroff")
		}); err != nil {
			return err
		}
	}
	if err := waitForServerState(scaleway, server.Identifier, "stopped"); err != nil {
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

	if err := retryWithCodes([]int{429, 400}, func() error {
		var req = api.ServerPatchDefinition{
			Volumes: &volumes,
		}
		return scaleway.PatchServer(serverID, req)
	}); err != nil {
		return err
	}

	if startServerAgain {
		if err := retry(func() error {
			return scaleway.PostServerAction(serverID, "poweron")
		}); err != nil {
			return err
		}
		if err := waitForServerState(scaleway, serverID, "running"); err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}
