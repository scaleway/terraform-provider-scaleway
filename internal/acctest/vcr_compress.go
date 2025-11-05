package acctest

import (
	"encoding/json"
	"fmt"
	"log"

	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	cassetteV3 "gopkg.in/dnaeon/go-vcr.v3/cassette"
	cassetteV4 "gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

var transientStates = map[string]bool{
	applesilicon.ServerStatusStarting.String():     true,
	applesilicon.ServerStatusRebooting.String():    true,
	applesilicon.ServerStatusReinstalling.String(): true,

	baremetal.ServerStatusDeleting.String():          true,
	baremetal.ServerStatusStarting.String():          true,
	baremetal.ServerStatusDelivering.String():        true,
	baremetal.ServerStatusStarting.String():          true,
	baremetal.ServerStatusMigrating.String():         true,
	baremetal.ServerInstallStatusInstalling.String(): true,

	block.VolumeStatusCreating.String():     true,
	block.VolumeStatusUpdating.String():     true,
	block.VolumeStatusDeleting.String():     true,
	block.VolumeStatusResizing.String():     true,
	block.VolumeStatusSnapshotting.String(): true,

	container.ContainerStatusCreating.String(): true,
	container.ContainerStatusDeleting.String(): true,
	container.ContainerStatusPending.String():  true,

	domain.DNSZoneStatusPending.String(): true,

	file.FileSystemStatusCreating.String(): true,

	function.FunctionStatusCreating.String(): true,
	function.FunctionStatusDeleting.String(): true,
	function.FunctionStatusPending.String():  true,

	instance.ServerStateStarting.String(): true,
	instance.ServerStateStopping.String(): true,

	k8s.ClusterStatusCreating.String(): true,
	k8s.ClusterStatusDeleting.String(): true,
	k8s.ClusterStatusUpdating.String(): true,
	k8s.PoolStatusDeleting.String():    true,
	k8s.PoolStatusScaling.String():     true,
	k8s.PoolStatusUpgrading.String():   true,

	mongodb.InstanceStatusDeleting.String():     true,
	mongodb.InstanceStatusSnapshotting.String(): true,
	mongodb.InstanceStatusConfiguring.String():  true,
	mongodb.InstanceStatusInitializing.String(): true,
	mongodb.InstanceStatusProvisioning.String(): true,

	rdb.DatabaseBackupStatusCreating.String():  true,
	rdb.DatabaseBackupStatusDeleting.String():  true,
	rdb.DatabaseBackupStatusExporting.String(): true,
	rdb.DatabaseBackupStatusRestoring.String(): true,
	rdb.InstanceStatusAutohealing.String():     true,
	rdb.InstanceStatusBackuping.String():       true,
	rdb.InstanceStatusConfiguring.String():     true,
	rdb.InstanceStatusDeleting.String():        true,
	rdb.InstanceStatusInitializing.String():    true,
	rdb.InstanceStatusProvisioning.String():    true,
	rdb.InstanceStatusRestarting.String():      true,
	rdb.InstanceStatusSnapshotting.String():    true,

	redis.ClusterStatusAutohealing.String():  true,
	redis.ClusterStatusConfiguring.String():  true,
	redis.ClusterStatusProvisioning.String(): true,
	redis.ClusterStatusDeleting.String():     true,
	redis.ClusterStatusInitializing.String(): true,

	tem.DomainStatusPending.String(): true,
}

type CompressReport struct {
	Path               string
	Logs               []string
	ErrorLogs          []string
	SkippedInteraction int
}

func (report *CompressReport) AddLog(log string) {
	report.Logs = append(report.Logs, log)
}

func (report *CompressReport) AddErrorLog(log string) {
	report.ErrorLogs = append(report.ErrorLogs, log)
}

func (report *CompressReport) Print() {
	log.Printf("On cassette: %s\n", report.Path)
	log.Printf("%d skipped interactions\n", report.SkippedInteraction)

	for _, msg := range report.ErrorLogs {
		log.Println(msg)
	}
}

func CompressCassetteV3(path string) (CompressReport, error) {
	inputCassette, err := cassetteV3.Load(path)
	if err != nil {
		log.Fatalf("Error while reading file : %v\n", err)
	}

	outputCassette := cassetteV3.New(path)
	transitioning := false

	report := CompressReport{
		SkippedInteraction: 0,
		Path:               path,
		ErrorLogs:          []string{},
		Logs:               []string{},
	}

	for i := range len(inputCassette.Interactions) {
		interaction := inputCassette.Interactions[i]
		responseBody := interaction.Response.Body
		requestMethod := interaction.Request.Method

		if requestMethod != "GET" {
			transitioning = false

			report.AddLog(fmt.Sprintf("Interaction %d in test %s is not a GET request. Recording it\n", i, path))
			outputCassette.AddInteraction(interaction)

			continue
		}

		if responseBody == "" {
			report.AddLog(fmt.Sprintf("Interaction %d in test %s got an empty response body. Recording it\n", i, path))
			outputCassette.AddInteraction(interaction)

			continue
		}

		var m map[string]any

		err := json.Unmarshal([]byte(responseBody), &m)
		if err != nil {
			report.AddErrorLog(fmt.Sprintf("Interaction %d in test %s have an error with unmarshalling response body: %v. Recording it\n", i, path, err))
			outputCassette.AddInteraction(interaction)

			continue
		}

		if m["status"] == nil {
			report.AddLog(fmt.Sprintf("Interaction %d in test %s does not contain a status field. Recording it\n", i, path))
			outputCassette.AddInteraction(interaction)

			continue
		}

		status := m["status"].(string)
		// We test if the state is transient
		if _, ok := transientStates[status]; ok {
			if transitioning {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is in a transient state while we are already in transitient state. No need to record it: %s\n", i, path, status))
				report.SkippedInteraction++
			} else {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is in a transient state: %s, Recording it\n", i, path, status))

				transitioning = true

				outputCassette.AddInteraction(interaction)
			}
		} else {
			if transitioning {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is not in a transient state anymore: %s, Recording it\n", i, path, status))
				outputCassette.AddInteraction(interaction)

				transitioning = false
			} else {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is not in a transient state: %s, Recording it\n", i, path, status))
				outputCassette.AddInteraction(interaction)
			}
		}
	}

	err = outputCassette.Save()
	if err != nil {
		return report, fmt.Errorf("error while saving file: %w", err)
	}

	return report, nil
}

func CompressCassetteV4(path string) (CompressReport, error) {
	inputCassette, err := cassetteV4.Load(path)
	if err != nil {
		log.Fatalf("Error while reading file : %v\n", err)
	}

	outputCassette := cassetteV4.New(path)
	transitioning := false

	report := CompressReport{
		SkippedInteraction: 0,
		Path:               path,
		ErrorLogs:          []string{},
		Logs:               []string{},
	}

	for i := range len(inputCassette.Interactions) {
		interaction := inputCassette.Interactions[i]
		responseBody := interaction.Response.Body
		requestMethod := interaction.Request.Method

		if requestMethod != "GET" {
			transitioning = false

			report.AddLog(fmt.Sprintf("Interaction %d in test %s is not a GET request. Recording it\n", i, path))
			outputCassette.AddInteraction(interaction)

			continue
		}

		if responseBody == "" {
			report.AddLog(fmt.Sprintf("Interaction %d in test %s got an empty response body. Recording it\n", i, path))
			outputCassette.AddInteraction(interaction)

			continue
		}

		var m map[string]any

		err := json.Unmarshal([]byte(responseBody), &m)
		if err != nil {
			report.AddErrorLog(fmt.Sprintf("Interaction %d in test %s have an error with unmarshalling response body: %v. Recording it\n", i, path, err))
			outputCassette.AddInteraction(interaction)

			continue
		}

		if m["status"] == nil {
			report.AddLog(fmt.Sprintf("Interaction %d in test %s does not contain a status field. Recording it\n", i, path))
			outputCassette.AddInteraction(interaction)

			continue
		}

		status := m["status"].(string)
		// We test if the state is transient
		if _, ok := transientStates[status]; ok {
			if transitioning {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is in a transient state while we are already in transitient state. No need to record it: %s\n", i, path, status))
				report.SkippedInteraction++
			} else {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is in a transient state: %s, Recording it\n", i, path, status))

				transitioning = true

				outputCassette.AddInteraction(interaction)
			}
		} else {
			if transitioning {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is not in a transient state anymore: %s, Recording it\n", i, path, status))
				outputCassette.AddInteraction(interaction)

				transitioning = false
			} else {
				report.AddLog(fmt.Sprintf("Interaction %d in test %s is not in a transient state: %s, Recording it\n", i, path, status))
				outputCassette.AddInteraction(interaction)
			}
		}
	}

	err = outputCassette.Save()
	if err != nil {
		return report, fmt.Errorf("error while saving file: %w", err)
	}

	return report, nil
}
