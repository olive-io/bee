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

package process

import (
	"strings"

	"github.com/olive-io/bpmn/schema"
	"github.com/samber/lo"

	"github.com/olive-io/bee/process/builder"
)

type Process struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`
	Desc string `json:"desc,omitempty" yaml:"desc,omitempty"`

	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	Vars map[string]any `json:"vars,omitempty" yaml:"vars,omitempty"`

	RemoteUser string `json:"remote_user,omitempty" yaml:"remote_user,omitempty"`

	Sudo     bool   `json:"sudo,omitempty" yaml:"sudo,omitempty"`
	SudoUser string `json:"sudo_user,omitempty" yaml:"sudo_user,omitempty"`

	Tasks []ITask `json:"tasks,omitempty" yaml:"tasks,omitempty"`

	Handlers []*Handler `json:"handlers,omitempty" yaml:"handlers,omitempty"`
}

func (p *Process) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var kv YamlKV
	if err = unmarshal(&kv); err != nil {
		return err
	}

	for key, value := range kv {
		if key == "name" {
			_, err = kv.Apply("name", &p.Name)
			if err != nil {
				return
			}
			continue
		}

		if key == "id" {
			_, err = kv.Apply("id", &p.Id)
			if err != nil {
				return err
			}
			continue
		}

		if key == "desc" {
			_, err = kv.Apply("desc", &p.Desc)
			if err != nil {
				return err
			}
			continue
		}

		if key == "hosts" {
			_, err = kv.ApplyArray("hosts", &p.Hosts)
			if err != nil {
				return
			}
			continue
		}

		if key == "vars" {
			_, err = kv.ApplyMap("vars", &p.Vars)
			if err != nil {
				return
			}
			continue
		}

		if key == "remote_user" {
			_, err = kv.Apply("remote_user", &p.RemoteUser)
			if err != nil {
				return
			}
		}

		if key == "sudo" {
			_, err = kv.Apply("sudo", &p.Sudo)
			if err != nil {
				return
			}
			continue
		}

		if key == "sudo_user" {
			_, err = kv.Apply("sudo_user", &p.SudoUser)
			if err != nil {
				return
			}
		}

		if key == "tasks" {
			vv, ok := value.([]any)
			if !ok {
				continue
			}
			p.Tasks = make([]ITask, len(vv))
			for i, item := range vv {
				var kind string
				ykv := item.(YamlKV)
				if exists, _ := ykv.Apply("kind", &kind); exists {
					switch kind {
					case ChildProcessKey:
						cp := new(ChildProcess)
						if err = cp.fromKV(ykv); err != nil {
							return
						}
						p.Tasks[i] = cp
					case ServiceKey:
						sv := new(Service)
						if err = sv.fromKV(ykv); err != nil {
							return
						}
						p.Tasks[i] = sv
					}
				} else {
					task := new(Task)
					if err = task.fromKV(ykv); err != nil {
						return
					}
					p.Tasks[i] = task
				}
			}
			continue
		}

		if key == "handlers" {
			vv, ok := value.([]any)
			if !ok {
				continue
			}
			p.Handlers = make([]*Handler, len(vv))
			for i, item := range vv {
				handler := new(Handler)
				if err = handler.fromKV(item.(YamlKV)); err != nil {
					return
				}
				p.Handlers[i] = handler
			}
		}
	}

	return nil
}

func (p *Process) ToBuilder() *Builder {
	return &Builder{p: p}
}

func (p *Process) Build() (*schema.Definitions, map[string]string, map[string]string, error) {
	pb := builder.NewProcessDefinitionsBuilder(p.Name)
	if p.Id == "" {
		p.Id = newSnoId()
	}
	pb.Id(p.Id)

	dataObjects := map[string]string{}
	properties := map[string]string{}

	hosts := p.Hosts
	if p.Sudo {
		properties["sudo"] = ""
	}
	if p.SudoUser != "" {
		properties["sudo_user"] = p.SudoUser
	}

	pb.Start()

	mappingPrefix := "__step_mapping__"
	for idx := range p.Tasks {
		st := p.Tasks[idx]
		switch act := st.(type) {
		case *ChildProcess:
			out, ds, props, err := buildChildProcess(act)
			if err != nil {
				return nil, nil, nil, err
			}
			for key, value := range ds {
				dataObjects[key] = value
			}
			for key, value := range props {
				properties[key] = value
			}
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(out)
		case *Task:
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := EncodeScriptTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		case *Service:
			sb := builder.NewServiceTaskBuilder(act.Name)
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := EncodeServiceTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		}
	}
	pb.End()

	for key, property := range pb.PopProperty() {
		properties[key] = property
	}
	properties["hosts"] = strings.Join(lo.Uniq[string](hosts), ",")

	definitions, err := pb.ToDefinitions()
	if err != nil {
		return nil, nil, nil, err
	}

	return definitions, dataObjects, properties, nil
}

func (p *Process) SubBuild() (*schema.Definitions, map[string]string, map[string]string, error) {
	pb := builder.NewSubProcessDefinitionsBuilder(p.Name)
	if p.Id == "" {
		p.Id = newSnoId()
	}
	pb.Id(p.Id)

	dataObjects := map[string]string{}
	properties := map[string]string{}

	hosts := p.Hosts
	if p.Sudo {
		properties["sudo"] = ""
	}
	if p.SudoUser != "" {
		properties["sudo_user"] = p.SudoUser
	}

	pb.Start()

	mappingPrefix := "__step_mapping__"
	for idx := range p.Tasks {
		st := p.Tasks[idx]
		switch act := st.(type) {
		case *ChildProcess:
			out, ds, props, err := buildChildProcess(act)
			if err != nil {
				return nil, nil, nil, err
			}
			for key, value := range ds {
				dataObjects[key] = value
			}
			for key, value := range props {
				properties[key] = value
			}
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(out)
		case *Task:
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := EncodeScriptTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		case *Service:
			sb := builder.NewServiceTaskBuilder(act.Name)
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := EncodeServiceTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		}
	}
	pb.End()

	for key, property := range pb.PopProperty() {
		properties[key] = property
	}
	properties["hosts"] = strings.Join(lo.Uniq[string](hosts), ",")

	definitions := pb.ToDefinitions()

	return definitions, dataObjects, properties, nil
}

func buildChildProcess(pr *ChildProcess) (*builder.SubProcessBuilder, map[string]string, map[string]string, error) {
	pb := builder.NewSubProcessDefinitionsBuilder(pr.Name)
	if pr.Id == "" {
		pr.Id = newSnoId()
	}
	pb.Id(pr.Id)

	dataObjects := map[string]string{}
	properties := map[string]string{}
	hosts := pr.Hosts

	pb.Start()

	mappingPrefix := "__step_mapping__"
	for idx := range pr.Tasks {
		st := pr.Tasks[idx]
		if act, ok := st.(*ChildProcess); ok {
			out, _, props, err := buildChildProcess(act)
			if err != nil {
				return nil, nil, nil, err
			}
			for key, value := range props {
				properties[key] = value
			}
			pb.AppendElem(out)
		}
		if act, ok := st.(*Task); ok {
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := EncodeScriptTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		}
		if act, ok := st.(*Service); ok {
			sb := builder.NewServiceTaskBuilder(act.Name)
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := EncodeServiceTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		}
	}
	pr.Hosts = lo.Uniq[string](hosts)

	pb.End()
	return pb.Out(), dataObjects, properties, nil
}
