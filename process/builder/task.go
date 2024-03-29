// Copyright 2024 The bee Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builder

import "github.com/olive-io/bpmn/schema"

type ScriptTaskBuilder struct {
	id       string
	name     string
	taskType string
	task     *schema.ScriptTask
}

func NewScriptTaskBuilder(name, taskType string) *ScriptTaskBuilder {
	task := schema.DefaultScriptTask()
	task.ExtensionElementsField = &schema.ExtensionElements{
		TaskHeaderField: &schema.TaskHeader{Header: make([]*schema.Item, 0)},
		PropertiesField: &schema.Properties{Property: make([]*schema.Item, 0)},
	}
	b := &ScriptTaskBuilder{
		id:       *randShapeName(&task),
		name:     name,
		taskType: taskType,
		task:     &task,
	}

	return b
}

func (b *ScriptTaskBuilder) SetId(id string) *ScriptTaskBuilder {
	b.id = id
	return b
}

func (b *ScriptTaskBuilder) SetHeader(name string, value any) *ScriptTaskBuilder {
	vv, vt := parseItemValue(value)
	b.task.ExtensionElementsField.TaskHeaderField.Header = append(b.task.ExtensionElementsField.TaskHeaderField.Header, &schema.Item{
		Name:  name,
		Value: vv,
		Type:  vt,
	})
	return b
}

func (b *ScriptTaskBuilder) SetProperty(name string, value any) *ScriptTaskBuilder {
	vv, vt := parseItemValue(value)
	b.task.ExtensionElementsField.PropertiesField.Property = append(b.task.ExtensionElementsField.PropertiesField.Property, &schema.Item{
		Name:  name,
		Value: vv,
		Type:  vt,
	})
	return b
}

func (b *ScriptTaskBuilder) Out() *schema.ScriptTask {
	b.task.SetId(&b.id)
	b.task.SetName(&b.name)
	b.task.ExtensionElementsField.TaskDefinitionField = &schema.TaskDefinition{Type: b.taskType}
	return b.task
}

type ServiceTaskBuilder struct {
	id   string
	name string
	task *schema.ServiceTask
}

func NewServiceTaskBuilder(name string) *ServiceTaskBuilder {
	task := schema.DefaultServiceTask()
	task.ExtensionElementsField = &schema.ExtensionElements{
		TaskHeaderField: &schema.TaskHeader{Header: make([]*schema.Item, 0)},
		PropertiesField: &schema.Properties{Property: make([]*schema.Item, 0)},
	}
	b := &ServiceTaskBuilder{
		id:   *randShapeName(&task),
		name: name,
		task: &task,
	}

	return b
}

func (b *ServiceTaskBuilder) SetId(id string) *ServiceTaskBuilder {
	b.id = id
	return b
}

func (b *ServiceTaskBuilder) SetHeader(name string, value any) *ServiceTaskBuilder {
	vv, vt := parseItemValue(value)
	b.task.ExtensionElementsField.TaskHeaderField.Header = append(b.task.ExtensionElementsField.TaskHeaderField.Header, &schema.Item{
		Name:  name,
		Value: vv,
		Type:  vt,
	})
	return b
}

func (b *ServiceTaskBuilder) SetProperty(name string, value any) *ServiceTaskBuilder {
	vv, vt := parseItemValue(value)
	b.task.ExtensionElementsField.PropertiesField.Property = append(b.task.ExtensionElementsField.PropertiesField.Property, &schema.Item{
		Name:  name,
		Value: vv,
		Type:  vt,
	})
	return b
}

func (b *ServiceTaskBuilder) Out() *schema.ServiceTask {
	b.task.SetId(&b.id)
	b.task.SetName(&b.name)
	b.task.ExtensionElementsField.TaskDefinitionField = &schema.TaskDefinition{}
	return b.task
}
