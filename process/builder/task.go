/*
   Copyright 2024 The bee Authors

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library;
*/

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
