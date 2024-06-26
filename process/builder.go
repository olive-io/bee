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

type Builder struct {
	p *Process
}

func NewProcessBuilder() *Builder {
	b := &Builder{p: &Process{
		Hosts:    []string{},
		Vars:     map[string]any{},
		Tasks:    []ITask{},
		Handlers: []*Handler{},
	}}
	return b
}

func (b *Builder) Named(id, name, desc string) *Builder {
	if id != "" {
		b.p.Id = id
	}
	if name != "" {
		b.p.Name = name
	}
	if desc != "" {
		b.p.Desc = desc
	}
	return b
}

func (b *Builder) SetHosts(hosts ...string) *Builder {
	b.p.Hosts = hosts
	return b
}

func (b *Builder) SetVar(name string, value any) *Builder {
	if b.p.Vars == nil {
		b.p.Vars = map[string]any{}
	}
	b.p.Vars[name] = value
	return b
}

func (b *Builder) SetHandlers(handlers ...*Handler) *Builder {
	b.p.Handlers = append(b.p.Handlers, handlers...)
	return b
}

func (b *Builder) SetTasks(tasks ...ITask) *Builder {
	b.p.Tasks = append(b.p.Tasks, tasks...)
	return b
}

func (b *Builder) Build() *Process {
	return b.p
}

type ChildProcessBuilder struct {
	p *ChildProcess
}

func NewChildProcessBuilder() *ChildProcessBuilder {
	b := &ChildProcessBuilder{
		p: &ChildProcess{
			Kind:     ChildProcessKey,
			Hosts:    []string{},
			Vars:     map[string]any{},
			Tasks:    []ITask{},
			Handlers: []*Handler{},
		},
	}
	return b
}

func (b *ChildProcessBuilder) Named(id, name, desc string) *ChildProcessBuilder {
	if id != "" {
		b.p.Id = id
	}
	if name != "" {
		b.p.Name = name
	}
	if desc != "" {
		b.p.Desc = desc
	}
	return b
}

func (b *ChildProcessBuilder) SetHosts(hosts ...string) *ChildProcessBuilder {
	b.p.Hosts = hosts
	return b
}

func (b *ChildProcessBuilder) SetVar(name string, value any) *ChildProcessBuilder {
	if b.p.Vars == nil {
		b.p.Vars = map[string]any{}
	}
	b.p.Vars[name] = value
	return b
}

func (b *ChildProcessBuilder) SetHandlers(handlers ...*Handler) *ChildProcessBuilder {
	b.p.Handlers = append(b.p.Handlers, handlers...)
	return b
}

func (b *ChildProcessBuilder) SetTasks(tasks ...ITask) *ChildProcessBuilder {
	b.p.Tasks = append(b.p.Tasks, tasks...)
	return b
}

func (b *ChildProcessBuilder) Build() *ChildProcess {
	return b.p
}

type TaskBuilder struct {
	p *Task
}

func NewTaskBuilder() *TaskBuilder {
	b := &TaskBuilder{
		p: &Task{
			Vars:   map[string]any{},
			Action: "ping",
			Args:   map[string]any{},
			Hosts:  []string{},
		},
	}
	return b
}

func (b *TaskBuilder) Named(id, name, desc string) *TaskBuilder {
	if id != "" {
		b.p.Id = id
	}
	if name != "" {
		b.p.Name = name
	}
	if desc != "" {
		b.p.Desc = desc
	}
	return b
}

func (b *TaskBuilder) SetHosts(hosts ...string) *TaskBuilder {
	b.p.Hosts = hosts
	return b
}

func (b *TaskBuilder) SetVar(name string, value any) *TaskBuilder {
	if b.p.Vars == nil {
		b.p.Vars = map[string]any{}
	}
	b.p.Vars[name] = value
	return b
}

func (b *TaskBuilder) SetAction(action string, args map[string]any) *TaskBuilder {
	b.p.Action = action
	b.p.Args = args
	return b
}

func (b *TaskBuilder) SetCatch(catch *Handler) *TaskBuilder {
	b.p.Catch = catch
	return b
}

func (b *TaskBuilder) SetFinish(finish *Handler) *TaskBuilder {
	b.p.Finish = finish
	return b
}

func (b *TaskBuilder) SetNotify(notify ...string) *TaskBuilder {
	b.p.Notify = append(b.p.Notify, notify...)
	return b
}

func (b *TaskBuilder) Build() *Task {
	return b.p
}

type ServiceBuilder struct {
	p *Service
}

func NewServiceBuilder() *ServiceBuilder {
	b := &ServiceBuilder{
		p: &Service{
			Kind:   ServiceKey,
			Vars:   map[string]any{},
			Action: "ping",
			Args:   map[string]any{},
			Hosts:  []string{},
		},
	}
	return b
}

func (b *ServiceBuilder) Named(id, name, desc string) *ServiceBuilder {
	if id != "" {
		b.p.Id = id
	}
	if name != "" {
		b.p.Name = name
	}
	if desc != "" {
		b.p.Desc = desc
	}
	return b
}

func (b *ServiceBuilder) SetHosts(hosts ...string) *ServiceBuilder {
	b.p.Hosts = hosts
	return b
}

func (b *ServiceBuilder) SetVar(name string, value any) *ServiceBuilder {
	if b.p.Vars == nil {
		b.p.Vars = map[string]any{}
	}
	b.p.Vars[name] = value
	return b
}

func (b *ServiceBuilder) SetAction(action string, args map[string]any) *ServiceBuilder {
	b.p.Action = action
	b.p.Args = args
	return b
}

func (b *ServiceBuilder) SetCatch(catch *Handler) *ServiceBuilder {
	b.p.Catch = catch
	return b
}

func (b *ServiceBuilder) SetFinish(finish *Handler) *ServiceBuilder {
	b.p.Finish = finish
	return b
}

func (b *ServiceBuilder) SetNotify(notify ...string) *ServiceBuilder {
	b.p.Notify = append(b.p.Notify, notify...)
	return b
}

func (b *ServiceBuilder) Build() *Service {
	return b.p
}

type HandlerBuilder struct {
	p *Handler
}

func NewHandlerBuilder() *HandlerBuilder {
	b := &HandlerBuilder{
		p: &Handler{
			Action: "ping",
			Args:   map[string]any{},
		},
	}
	return b
}

func (b *HandlerBuilder) Named(id, name, desc string) *HandlerBuilder {
	if id != "" {
		b.p.Id = id
	}
	if name != "" {
		b.p.Name = name
	}
	if desc != "" {
		b.p.Desc = desc
	}
	return b
}

func (b *HandlerBuilder) SetKind(kind string) *HandlerBuilder {
	b.p.Kind = kind
	return b
}

func (b *HandlerBuilder) SetAction(action string, args map[string]any) *HandlerBuilder {
	b.p.Action = action
	b.p.Args = args
	return b
}

func (b *HandlerBuilder) Build() *Handler {
	return b.p
}
