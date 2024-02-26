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
	"github.com/muyo/sno"
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
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func (rt *Runtime) Play(ctx context.Context, pr *process.Process, opts ...RunOption) error {
	definitions, dataObjects, properties, err := rt.BuildBpmnProcess(pr)
	if err != nil {
		return err
	}

	return rt.RunBpmnProcess(ctx, definitions, dataObjects, properties, opts...)
}

func (rt *Runtime) RunBpmnProcess(ctx context.Context, definitions *schema.Definitions, dataObjects, properties map[string]string, opts ...RunOption) error {

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
		sources = strings.Split(v, ",")
	}

	if len(sources) == 0 {
		return fmt.Errorf("missing sources")
	}

	if err := rt.inventory.AddSources(sources...); err != nil {
		return err
	}

	processElement := (*definitions.Processes())[0]
	proc := bprocess.New(&processElement, definitions)

	_properties := map[string]any{}
	for key, value := range _properties {
		_properties[key] = value
	}
	_dataObjects := map[string]any{}
	for key, value := range dataObjects {
		_dataObjects[key] = value
	}
	bpmnOptions := []instance.Option{
		instance.WithVariables(_properties),
		instance.WithDataObjects(_dataObjects),
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

	runTasks := make([]process.ITask, 0)

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
				runTasks = append(runTasks, sv)
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
							break LOOP
						}

						stdout := map[string]any{}
						if err = json.Unmarshal(data, &stdout); err != nil {
							result.ErrMsg = err.Error()
							cb.RunnerOkFailed(result)
							tt.Do(activity.WithErr(err))
							break LOOP
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
				runTasks = append(runTasks, task)
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
						break LOOP
					}

					stdout := map[string]any{}
					if err = json.Unmarshal(data, &stdout); err != nil {
						result.ErrMsg = err.Error()
						cb.RunnerOkFailed(result)
						tt.Do(activity.WithErr(err))
						break LOOP
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
			err = tt.Error
			break LOOP
		case flow.CeaseFlowTrace:
			break LOOP
		default:
			lg.Sugar().Debugf("%#v", tt)
		}
	}
	ins.WaitUntilComplete(ctx)

	if err != nil {
		for i := len(runTasks) - 1; i >= 0; i-- {
			task := runTasks[i]

			caught, ok := task.(process.ICatchTask)
			if !ok {
				continue
			}

			fields := make([]zap.Field, 0)
			if named, ok := task.(process.INamedTask); ok {
				fields = append(fields,
					zap.String("name", named.GetName()),
					zap.String("id", named.GetId()))
			}

			catch := caught.GetCatch()
			if catch == nil {
				lg.Debug("skip task catch", fields...)
				continue
			}

			hosts := caught.GetHosts()
			if len(hosts) == 0 {
				hosts = sources
			}

			fields = append(fields, zap.Stringer("handler", catch))
			lg.Info("handle task catch", fields...)

			_ = rt.handle(ctx, hosts, catch)
		}
	}

	for i := len(runTasks) - 1; i >= 0; i-- {
		task := runTasks[i]

		caught, ok := task.(process.ICatchTask)
		if !ok {
			continue
		}

		fields := make([]zap.Field, 0)
		if named, ok := task.(process.INamedTask); ok {
			fields = append(fields,
				zap.String("name", named.GetName()),
				zap.String("id", named.GetId()))
		}

		finish := caught.GetFinish()
		if finish == nil {
			lg.Debug("skip service catch", fields...)
			continue
		}

		hosts := caught.GetHosts()
		if len(hosts) == 0 {
			hosts = sources
		}

		fields = append(fields, zap.Stringer("handler", finish))
		lg.Info("handle service finish", fields...)

		_ = rt.handle(ctx, hosts, finish)
	}

	return err
}

func (rt *Runtime) handle(ctx context.Context, hosts []string, handler *process.Handler, opts ...RunOption) error {
	switch handler.Kind {
	case process.ServiceKey:
		caller := rt.opts.caller
		if caller == nil {
			return nil
		}

		for _, host := range hosts {
			ropts := append(opts)
			in, _ := json.Marshal(handler.Args)
			data, err := caller(ctx, host, handler.Action, in, ropts...)
			if err != nil {
				return err
			}

			stdout := map[string]any{}
			if err = json.Unmarshal(data, &stdout); err != nil {
			}
		}

	case process.TaskKey:
		args := make([]string, 0)
		args = append(args, handler.Action)
		for name, arg := range handler.Args {
			value, _ := json.Marshal(arg)
			args = append(args, name+"="+strings.ReplaceAll(string(value), "\"", ""))
		}
		shell := strings.Join(args, " ")
		for _, host := range hosts {
			ropts := append(opts)
			data, err := rt.Execute(ctx, host, shell, ropts...)
			if err != nil {
				return err
			}

			stdout := map[string]any{}
			if err = json.Unmarshal(data, &stdout); err != nil {
			}
		}
	}

	return nil
}

func (rt *Runtime) BuildBpmnProcess(pr *process.Process) (*schema.Definitions, map[string]string, map[string]string, error) {
	pb := builder.NewProcessDefinitionsBuilder(pr.Name)
	if pr.Id == "" {
		pr.Id = newSnoId()
	}
	pb.Id(pr.Id)

	dataObjects := map[string]string{}
	properties := map[string]string{}

	hosts := pr.Hosts
	if pr.Sudo {
		properties["sudo"] = ""
	}
	if pr.SudoUser != "" {
		properties["sudo_user"] = pr.SudoUser
	}

	pb.Start()

	mappingPrefix := "__step_mapping__"
	for idx := range pr.Tasks {
		st := pr.Tasks[idx]
		switch act := st.(type) {
		case *process.ChildProcess:
			out, props, err := buildChildProcess(act)
			if err != nil {
				return nil, nil, nil, err
			}
			for key, value := range props {
				properties[key] = value
			}
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(out)
		case *process.Task:
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := encodeScriptTask(act)
			for key, value := range props {
				sb.SetProperty(key, value)
			}
			for key, value := range headers {
				sb.SetHeader(key, value)
			}
			pb.SetProperty(mappingPrefix+act.Id, strings.Join(act.Hosts, ","))
			hosts = append(hosts, act.Hosts...)
			pb.AppendElem(sb.Out())
		case *process.Service:
			sb := builder.NewServiceTaskBuilder(act.Name)
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := encodeServiceTask(act)
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

func buildChildProcess(pr *process.ChildProcess) (*builder.SubProcessBuilder, map[string]string, error) {
	pb := builder.NewSubProcessDefinitionsBuilder(pr.Name)
	if pr.Id == "" {
		pr.Id = newSnoId()
	}
	pb.Id(pr.Id)

	properties := map[string]string{}
	hosts := pr.Hosts

	pb.Start()

	mappingPrefix := "__step_mapping__"
	for idx := range pr.Tasks {
		st := pr.Tasks[idx]
		if act, ok := st.(*process.ChildProcess); ok {
			out, props, err := buildChildProcess(act)
			if err != nil {
				return nil, nil, err
			}
			for key, value := range props {
				properties[key] = value
			}
			pb.AppendElem(out)
		}
		if act, ok := st.(*process.Task); ok {
			sb := builder.NewScriptTaskBuilder(act.Name, "tengo")
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := encodeScriptTask(act)
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
		if act, ok := st.(*process.Service); ok {
			sb := builder.NewServiceTaskBuilder(act.Name)
			if act.Id == "" {
				act.Id = newSnoId()
			}
			sb.SetId(act.Id)
			props, headers := encodeServiceTask(act)
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
	return pb.Out(), properties, nil
}

func encodeScriptTask(task *process.Task) (props map[string]any, headers map[string]any) {
	props = map[string]any{}
	headers = map[string]any{}

	token, _ := yaml.Marshal(task)
	props["token"] = string(token)
	headers["hosts"] = strings.Join(task.Hosts, ",")

	return
}

func decodeScriptTask(props, headers map[string]any) *process.Task {
	var task *process.Task
	if v, ok := props["token"]; ok {
		vv, _ := v.(string)
		_ = yaml.Unmarshal([]byte(vv), &task)
	}

	return task
}

func encodeServiceTask(service *process.Service) (props map[string]any, headers map[string]any) {
	props = map[string]any{}
	headers = map[string]any{}

	token, _ := yaml.Marshal(service)
	props["token"] = string(token)
	headers["hosts"] = strings.Join(service.Hosts, ",")

	return
}

func decodeServiceTask(props, headers map[string]any) *process.Service {
	var s *process.Service
	if v, ok := props["token"]; ok {
		vv, _ := v.(string)
		_ = yaml.Unmarshal([]byte(vv), &s)
	}

	return s
}

func newSnoId() string {
	return string(sno.New(0).Bytes())
}
