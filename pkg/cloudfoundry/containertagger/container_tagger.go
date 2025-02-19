// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package containertagger

import (
	"context"
	"fmt"
	"strings"

	"code.cloudfoundry.org/garden"

	"github.com/DataDog/datadog-agent/pkg/metadata/host"
	"github.com/DataDog/datadog-agent/pkg/tagger/utils"
	"github.com/DataDog/datadog-agent/pkg/util/cloudproviders/cloudfoundry"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/DataDog/datadog-agent/pkg/workloadmeta"
)

const (
	componentName = "cloudfoundry-container-tagger"
)

// ContainerTagger is a simple component that injects host tags and CAPI metadata
// into cloudfoundry containers. It listens to container collection events from
// the workloadmeta store and injects tags accordingly if it detects a diff
// with the previously injected tags.
type ContainerTagger struct {
	gardenUtil            cloudfoundry.GardenUtilInterface
	store                 workloadmeta.Store
	seen                  map[string]struct{}
	tagsHashByContainerID map[string]string
}

// NewContainerTagger creates a new container tagger.
// Call Start to start the container tagger.
func NewContainerTagger() (*ContainerTagger, error) {
	gu, err := cloudfoundry.GetGardenUtil()
	if err != nil {
		return nil, err
	}
	return &ContainerTagger{
		gardenUtil:            gu,
		store:                 workloadmeta.GetGlobalStore(),
		seen:                  make(map[string]struct{}),
		tagsHashByContainerID: make(map[string]string),
	}, nil
}

// Start starts the container tagger.
// Cancel the context to stop the container tagger.
func (c *ContainerTagger) Start(ctx context.Context) {
	go func() {
		filter := workloadmeta.NewFilter([]workloadmeta.Kind{workloadmeta.KindContainer}, workloadmeta.SourceClusterOrchestrator, workloadmeta.EventTypeAll)
		ch := c.store.Subscribe(componentName, workloadmeta.NormalPriority, filter)
		defer c.store.Unsubscribe(ch)
		for {
			select {
			case bundle := <-ch:
				// close Ch to indicate that the Store can proceed to the next subscriber
				close(bundle.Ch)

				for _, evt := range bundle.Events {
					err := c.processEvent(ctx, evt)
					if err != nil {
						log.Warnf("%v", err)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Infof("CloudFoundry container tagger successfully started")
}

func (c *ContainerTagger) processEvent(ctx context.Context, evt workloadmeta.Event) error {
	containerID := evt.Entity.GetID().ID

	if evt.Type == workloadmeta.EventTypeSet {
		storeContainer := evt.Entity.(*workloadmeta.Container)

		// extract tags
		hostTags := host.GetHostTags(ctx, true)
		tags := storeContainer.CollectorTags
		tags = append(tags, hostTags.System...)
		tags = append(tags, hostTags.GoogleCloudPlatform...)

		// hashing tags to keep track of containers tags
		// as they contain the container_id as `app_instance_guid`
		tagsHash := utils.ComputeTagsHash(tags)

		// will be useful for deletion
		c.tagsHashByContainerID[containerID] = tagsHash

		// check if the tags were already injected
		if _, exist := c.seen[tagsHash]; exist {
			return nil
		}

		// mark as seen
		c.seen[tagsHash] = struct{}{}

		container, err := c.gardenUtil.GetContainer(containerID)
		if err != nil {
			return fmt.Errorf("error retrieving container %s from the garden API: %v", containerID, err)
		}

		log.Infof("Updating tags in container %s", containerID)
		go func() {
			err = updateTagsInContainer(container, tags)
			if err != nil {
				log.Errorf("Error running a process inside container %s: %v", containerID, err)
			}
		}()

	} else if evt.Type == workloadmeta.EventTypeUnset {
		hash := c.tagsHashByContainerID[containerID]
		delete(c.seen, hash)
		delete(c.tagsHashByContainerID, containerID)
	}
	return nil
}

// updateTagsInContainer runs a script inside the container that handles updating the agent with the given tags
func updateTagsInContainer(container garden.Container, tags []string) error {
	process, err := container.Run(garden.ProcessSpec{
		Path: "/bin/sh",
		Args: []string{"/home/vcap/app/.datadog/scripts/update_agent_config.sh"},
		User: "vcap",
		Env:  []string{fmt.Sprintf("DD_NODE_AGENT_TAGS=%s", strings.Join(tags, ","))},
	}, garden.ProcessIO{})
	if err != nil {
		return err
	}
	exitCode, err := process.Wait()
	if err != nil {
		return err
	}
	log.Debugf("Process %s exited with code: %d", process.ID(), exitCode)
	return nil
}
