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

package bee

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	json "github.com/json-iterator/go"
	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/plugins/filter"
	"github.com/olive-io/bee/process"
	"github.com/olive-io/bee/process/builder"
	"github.com/olive-io/bee/stats"
	"github.com/olive-io/bpmn/flow"
	"github.com/olive-io/bpmn/flow_node/activity"
	"github.com/olive-io/bpmn/flow_node/activity/script"
	"github.com/olive-io/bpmn/flow_node/activity/service"
	bprocess "github.com/olive-io/bpmn/process"
	"github.com/olive-io/bpmn/process/instance"
	"github.com/olive-io/bpmn/schema"
	"github.com/olive-io/bpmn/tracing"
)

func (rt *Runtime) Play(ctx context.Context, pr *process.Process, opts ...RunOption) error {
	definitions, dataObjects, properties, err := rt.BuildBpmnProcess(pr)
	if err != nil {
		return err
	}

	return rt.RunBpmnProcess(ctx, definitions, dataObjects, properties, opts...)
}

func (rt *Runtime) RunBpmnProcess(
	ctx context.Context, definitions *schema.Definitions,
	dataObjects, properties map[string]any, opts ...RunOption) error {

	lg := rt.Logger()

	runOptions := newRunOptions()
	for _, opt := range opts {
		opt(runOptions)
	}

	cb := callback.NewCallBack()
	if runOptions.Callback != nil {
		cb = runOptions.Callback
	}

	ft := filter.NewFilter()
	if runOptions.Filter != nil {
		ft = runOptions.Filter
	}

	var sources []string
	if v, ok := properties["hosts"]; ok {
		sources, _ = v.([]string)
	}

	if len(sources) == 0 {
		return fmt.Errorf("missing sources")
	}

	if err := rt.inventory.AddSources(sources...); err != nil {
		return err
	}

	processElement := (*definitions.Processes())[0]
	proc := bprocess.New(&processElement, definitions)
	bpmnOptions := []instance.Option{
		instance.WithDataObjects(dataObjects),
		instance.WithVariables(properties),
	}

	ins, err := proc.Instantiate(bpmnOptions...)
	if err != nil {
		return err
	}
	if tracer := runOptions.Tracer; tracer != nil {
		ins.Tracer.SubscribeChannel(tracer)
		defer ins.Tracer.Unsubscribe(tracer)
	}

	traces := ins.Tracer.Subscribe()
	err = ins.StartAll(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run the instance")
	}
	defer ins.Tracer.Unsubscribe(traces)

LOOP:
	for {
		var trace tracing.ITrace
		select {
		case <-ctx.Done():
			return context.Canceled
		case trace = <-traces:
		}

		if trace == nil {
			continue
		}

		trace = tracing.Unwrap(trace)
		switch tt := trace.(type) {
		case flow.Trace:
		case *activity.Trace:
			act := tt.GetActivity()
			id, _ := act.Element().Id()

			tProps, tHeaders := ft.OnPreTaskProps(*id, tt.GetProperties(), tt.GetHeaders())
			rspProperties := map[string]any{}

			switch act.(type) {
			case *service.ServiceTask:

				sv := decodeServiceTask(tProps, tHeaders)
				hosts := sv.Hosts
				if len(hosts) == 0 {
					hosts = sources
				}

				if caller := rt.opts.caller; caller != nil {
					for _, host := range hosts {
						result := &stats.TaskResult{
							Host: host,
						}

						ropts := append(opts, WithMetadata(tHeaders))
						in, _ := json.Marshal(sv.Args)
						data, err := caller(ctx, host, sv.Action, in, ropts...)
						if err != nil {
							result.ErrMsg = err.Error()
							cb.RunnerOkFailed(result)
							tt.Do(activity.WithErr(err))
							goto LOOP
						}

						stdout := map[string]any{}
						if err = json.Unmarshal(data, &stdout); err != nil {
							result.ErrMsg = err.Error()
							cb.RunnerOkFailed(result)
							tt.Do(activity.WithErr(err))
							goto LOOP
						}

						stdout = ft.OnPostTaskStdout(*id, stdout)

						result.Stdout = stdout
						for key, value := range stdout {
							rspProperties[key] = value
						}

						cb.RunnerOnOk(result)
					}
				}

			case *script.ScriptTask:

				task := decodeScriptTask(tProps, tHeaders)
				hosts := task.Hosts
				if len(hosts) == 0 {
					hosts = sources
				}

				args := make([]string, 0)
				args = append(args, task.Action)
				for name, arg := range task.Args {
					value, _ := json.Marshal(arg)
					args = append(args, name+"="+strings.ReplaceAll(string(value), "\"", ""))
				}
				shell := strings.Join(args, " ")
				for _, host := range hosts {
					result := &stats.TaskResult{
						Host: host,
					}

					ropts := append(opts, WithMetadata(tHeaders))
					data, err := rt.Execute(ctx, host, shell, ropts...)
					if err != nil {
						result.ErrMsg = err.Error()
						cb.RunnerOkFailed(result)
						tt.Do(activity.WithErr(err))
						goto LOOP
					}

					stdout := map[string]any{}
					if err = json.Unmarshal(data, &stdout); err != nil {
						result.ErrMsg = err.Error()
						cb.RunnerOkFailed(result)
						tt.Do(activity.WithErr(err))
						goto LOOP
					}

					stdout = ft.OnPostTaskStdout(*id, stdout)

					result.Stdout = stdout
					for key, value := range stdout {
						rspProperties[key] = value
					}

					cb.RunnerOnOk(result)
				}
			}

			tt.Do(activity.WithProperties(rspProperties))
		case tracing.ErrorTrace:
			return tt.Error
		case flow.CeaseFlowTrace:
			break LOOP
		default:
			lg.Sugar().Debugf("%#v", tt)
		}
	}
	ins.WaitUntilComplete(ctx)

	return nil
}

func (rt *Runtime) BuildBpmnProcess(pr *process.Process) (*schema.Definitions, map[string]any, map[string]any, error) {
	pb := builder.NewProcessDefinitionsBuilder(pr.Name)
	pb.Id(pr.Id)
	pb.Start()

	dataObjects := map[string]any{}
	properties := map[string]any{}

	properties["hosts"] = pr.Hosts
	if pr.Sudo {
		properties["sudo"] = ""
	}
	if pr.SudoUser != "" {
		properties["sudo_user"] = pr.SudoUser
	}

	for idx := range pr.Tasks {
		st := pr.Tasks[idx]
		switch act := st.(type) {
		case *process.ChildProcess:
			out, err := buildChildProcess(act)
			if err != nil {
				return nil, nil, nil, err
			}
			pb.AppendElem(out)
		case *process.Task:
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			sb.SetId(act.Id)
			props, headers := encodeScriptTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.AppendElem(sb.Out())
		case *process.Service:
			sb := builder.NewServiceTaskBuilder(act.Name)
			sb.SetId(act.Id)
			props, headers := encodeServiceTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.AppendElem(sb.Out())
		}
	}
	pb.End()

	for key, property := range pb.PopProperty() {
		properties[key] = property
	}

	definitions, err := pb.ToDefinitions()
	if err != nil {
		return nil, nil, nil, err
	}

	return definitions, dataObjects, properties, nil
}

func buildChildProcess(pr *process.ChildProcess) (*builder.SubProcessBuilder, error) {
	pb := builder.NewSubProcessDefinitionsBuilder(pr.Name)
	pb.Id(pr.Id)
	pb.Start()

	for idx := range pr.Tasks {
		st := pr.Tasks[idx]
		if act, ok := st.(*process.Task); ok {
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			sb.SetId(act.Id)
			props, headers := encodeScriptTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.AppendElem(sb.Out())
		}
		if act, ok := st.(*process.Service); ok {
			sb := builder.NewServiceTaskBuilder(act.Name)
			sb.SetId(act.Id)
			props, headers := encodeServiceTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.AppendElem(sb.Out())
		}
	}

	pb.End()
	return pb.Out(), nil
}

func encodeScriptTask(task *process.Task) (props map[string]any, headers map[string]any) {
	props = map[string]any{}
	headers = map[string]any{}

	if task.Id != "" {
		headers["id"] = task.Id
	}
	if task.Name != "" {
		headers["name"] = task.Name
	}
	if task.RemoteUser != "" {
		headers["remote_user"] = task.RemoteUser
	}
	for key, value := range task.Vars {
		headers["__var_"+key] = value
	}

	if task.Sudo {
		props["sudo"] = ""
	}
	if task.SudoUser != "" {
		props["sudo_user"] = task.SudoUser
	}
	if len(task.Hosts) != 0 {
		props["hosts"] = strings.Join(task.Hosts, ",")
	}
	if task.Action != "" {
		props["action"] = task.Action
	}
	for name, arg := range task.Args {
		props["__arg_"+name] = arg
	}

	if len(task.Notify) != 0 {
		props["notify"] = strings.Join(task.Notify, ",")
	}

	return
}

func decodeScriptTask(props, headers map[string]any) *process.Task {
	task := &process.Task{
		Vars: map[string]any{},
		Args: map[string]any{},
	}
	for key, value := range props {
		if key == "sudo" {
			task.Sudo = true
		}
		if key == "sudo_user" {
			task.SudoUser = value.(string)
		}
		if key == "hosts" {
			task.Hosts = strings.Split(value.(string), ",")
		}

		if key == "action" {
			task.Action = value.(string)
		}
		if strings.HasPrefix(key, "__arg_") {
			task.Args[strings.TrimPrefix(key, "__arg_")] = value
		}

		if key == "notify" {
			task.Notify = strings.Split(value.(string), ",")
		}
	}

	for key, value := range headers {
		if key == "id" {
			task.Id = value.(string)
		}
		if key == "name" {
			task.Name = value.(string)
		}
		if key == "remote_user" {
			task.RemoteUser = value.(string)
		}
		if strings.HasPrefix(key, "__var_") {
			task.Vars[strings.TrimPrefix(key, "__var_")] = value.(string)
		}
	}

	return task
}

func encodeServiceTask(s *process.Service) (props map[string]any, headers map[string]any) {
	props = map[string]any{}
	headers = map[string]any{}

	if s.Id != "" {
		headers["id"] = s.Id
	}
	if s.Name != "" {
		headers["name"] = s.Name
	}
	for key, value := range s.Vars {
		headers["__var_"+key] = value
	}
	if len(s.Hosts) != 0 {
		props["hosts"] = strings.Join(s.Hosts, ",")
	}
	if s.Action != "" {
		props["action"] = s.Action
	}
	for name, arg := range s.Args {
		props["__arg_"+name] = arg
	}

	if len(s.Notify) != 0 {
		props["notify"] = strings.Join(s.Notify, ",")
	}

	return
}

func decodeServiceTask(props, headers map[string]any) *process.Service {
	s := &process.Service{
		Vars: map[string]any{},
		Args: map[string]any{},
	}
	for key, value := range props {
		if key == "hosts" {
			s.Hosts = strings.Split(value.(string), ",")
		}

		if key == "action" {
			s.Action = value.(string)
		}
		if strings.HasPrefix(key, "__arg_") {
			s.Args[strings.TrimPrefix(key, "__arg_")] = value
		}

		if key == "notify" {
			s.Notify = strings.Split(value.(string), ",")
		}
	}

	for key, value := range headers {
		if key == "id" {
			s.Id = value.(string)
		}
		if key == "name" {
			s.Name = value.(string)
		}
		if strings.HasPrefix(key, "__var_") {
			s.Vars[strings.TrimPrefix(key, "__var_")] = value
		}
	}

	return s
}
