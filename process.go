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
			tProps, tHeaders := ft.OnPreTaskProps(tt.GetProperties(), tt.GetHeaders())

			task := decodeTask(tProps, tHeaders)

			hosts := task.Hosts
			if len(hosts) == 0 {
				hosts = sources
			}

			rspProperties := map[string]any{}
			args := make([]string, 0)
			args = append(args, task.Action)
			for name, arg := range task.Args {
				args = append(args, name+"="+arg)
			}
			shell := strings.Join(args, " ")
			for _, host := range hosts {
				result := &stats.TaskResult{
					Host: host,
				}

				data, err := rt.Execute(ctx, host, shell, opts...)
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

				stdout = ft.OnPostTaskStdout(stdout)

				result.Stdout = stdout
				for key, value := range stdout {
					rspProperties[key] = value
				}

				cb.RunnerOnOk(result)
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
			props, headers := encodeTask(act)
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
			props, headers := encodeTask(act)
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

func encodeTask(task *process.Task) (props map[string]any, headers map[string]any) {
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

	if task.Sudo {
		props["sudo"] = ""
	}
	if task.SudoUser != "" {
		props["sudo_user"] = task.SudoUser
	}
	for key, value := range task.Vars {
		props["__var_"+key] = value
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

func decodeTask(props, headers map[string]any) *process.Task {
	task := &process.Task{
		Vars: map[string]string{},
		Args: map[string]string{},
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

		if strings.HasPrefix(key, "__var_") {
			task.Vars[strings.TrimPrefix(key, "__var_")] = value.(string)
		}

		if key == "action" {
			task.Action = value.(string)
		}
		if strings.HasPrefix(key, "__arg_") {
			task.Args[strings.TrimPrefix(key, "__arg_")] = value.(string)
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
	}

	return task
}
