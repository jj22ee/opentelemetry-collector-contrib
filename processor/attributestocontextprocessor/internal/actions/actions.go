// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package actions // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributestocontextprocessor/internal/actions"

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// Action is the enum to capture the four types of actions to perform on client metadata.
type Action string

const (
	INSERT Action = "insert"
	UPDATE Action = "update"
	UPSERT Action = "upsert"
	DELETE Action = "delete"
)

// KeyValue specifies the client metadata key to act upon.
type KeyValue struct {
	Key                   string `mapstructure:"key"`
	Action                Action `mapstructure:"action"`
	Value                 any    `mapstructure:"value"`
	FromAttribute         string `mapstructure:"from_attribute"`
	FromResourceAttribute string `mapstructure:"from_resource_attribute"`
}

type metadataAction struct {
	Key                   string
	Action                Action
	StaticValue           string
	FromResourceAttribute string
	FromAttribute         string
}

// Actions defines the interface for processing client metadata actions.
type Actions interface {
	ProcessStatic(metadata map[string][]string)
	ProcessResource(metadata map[string][]string, attrs pcommon.Map)
	ProcessAttributes(metadata map[string][]string, attrs pcommon.Map)
	HasResourceActions() bool
	HasAttributeActions() bool
}

type actions struct {
	staticActions    []metadataAction
	resourceActions  []metadataAction
	attributeActions []metadataAction
}

func NewActions(keyValues []KeyValue) Actions {
	var staticActions []metadataAction
	var resourceActions []metadataAction
	var attributeActions []metadataAction

	for _, kv := range keyValues {
		action := metadataAction{
			Key:    kv.Key,
			Action: kv.Action,
		}

		switch {
		case kv.Value != nil:
			action.StaticValue = fmt.Sprintf("%v", kv.Value)
			staticActions = append(staticActions, action)
		case kv.FromResourceAttribute != "":
			action.FromResourceAttribute = kv.FromResourceAttribute
			resourceActions = append(resourceActions, action)
		case kv.FromAttribute != "":
			action.FromAttribute = kv.FromAttribute
			attributeActions = append(attributeActions, action)
		case kv.Action == DELETE:
			staticActions = append(staticActions, action)
		}
	}

	return &actions{
		staticActions:    staticActions,
		resourceActions:  resourceActions,
		attributeActions: attributeActions,
	}
}

func (a *actions) HasResourceActions() bool {
	return len(a.resourceActions) > 0
}

func (a *actions) HasAttributeActions() bool {
	return len(a.attributeActions) > 0
}

func (a *actions) ProcessStatic(metadata map[string][]string) {
	for _, action := range a.staticActions {
		if action.StaticValue != "" {
			processAction(metadata, action, action.StaticValue)
		} else if action.Action == DELETE {
			processAction(metadata, action, "")
		}
	}
}

func (a *actions) ProcessResource(metadata map[string][]string, attrs pcommon.Map) {
	for _, action := range a.resourceActions {
		if value := getAttributeValue(attrs, action.FromResourceAttribute); value != "" {
			processAction(metadata, action, value)
		}
	}
}

func (a *actions) ProcessAttributes(metadata map[string][]string, attrs pcommon.Map) {
	for _, action := range a.attributeActions {
		if value := getAttributeValue(attrs, action.FromAttribute); value != "" {
			processAction(metadata, action, value)
		}
	}
}

func getAttributeValue(attrs pcommon.Map, key string) string {
	if key != "" {
		if val, found := attrs.Get(key); found {
			return val.AsString()
		}
	}
	return ""
}

func processAction(metadata map[string][]string, action metadataAction, value string) {
	switch action.Action {
	case INSERT:
		if _, exists := metadata[action.Key]; !exists {
			metadata[action.Key] = []string{value}
		}
	case UPDATE:
		if _, exists := metadata[action.Key]; exists {
			metadata[action.Key] = []string{value}
		}
	case UPSERT:
		metadata[action.Key] = []string{value}
	case DELETE:
		delete(metadata, action.Key)
	}
}
