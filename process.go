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
	"github.com/olive-io/bee/process"
	"github.com/olive-io/bee/process/builder"
	"github.com/olive-io/bpmn/flow"
	"github.com/olive-io/bpmn/flow_node/activity"
	bprocess "github.com/olive-io/bpmn/process"
	"github.com/olive-io/bpmn/process/instance"
	"github.com/olive-io/bpmn/schema"
	"github.com/olive-io/bpmn/tracing"
)

func (rt *Runtime) Play(ctx context.Context, pr *process.Process, opts ...RunOption) error {
	lg := rt.Logger()

	definitions, _dataObjects, _properties, err := buildBpmn(pr)
	if err != nil {
		return err
	}

	if err = rt.inventory.AddSources(pr.Hosts...); err != nil {
		return err
	}

	processElement := (*definitions.Processes())[0]
	proc := bprocess.New(&processElement, definitions)
	options := []instance.Option{
		instance.WithDataObjects(_dataObjects),
		instance.WithVariables(_properties),
	}

	ins, err := proc.Instantiate(options...)
	if err != nil {
		return err
	}
	traces := ins.Tracer.Subscribe()
	err = ins.StartAll(context.Background())
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

		trace = tracing.Unwrap(trace)
		switch tt := trace.(type) {
		case flow.Trace:
		case *activity.Trace:
			var task process.Task
			act := tt.GetActivity()
			id, _ := act.Element().Id()
			properties := tt.GetProperties()
			data, ok := properties["body"]
			if !ok {
				err = fmt.Errorf("missing script %s body", *id)
				tt.Do(activity.WithErr(err))
				break
			}
			err = json.Unmarshal([]byte(data.(string)), &task)
			if err != nil {
				tt.Do(activity.WithErr(err))
				break
			}

			hosts := task.Hosts
			if len(hosts) == 0 {
				hosts = pr.Hosts
			}

			args := make([]string, 0)
			args = append(args, task.Module)
			for name, arg := range task.Args {
				args = append(args, name+"="+arg)
			}
			shell := strings.Join(args, " ")
			for _, host := range hosts {
				_, err = rt.Execute(ctx, host, shell, opts...)
				if err != nil {
					tt.Do(activity.WithErr(err))
					goto LOOP
				}
			}

			tt.Do()

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

func buildBpmn(pr *process.Process) (*schema.Definitions, map[string]any, map[string]any, error) {
	pb := builder.NewProcessDefinitionsBuilder(pr.Name)
	pb.Id(pr.Id)
	pb.Start()

	dataObjects := map[string]any{}
	properties := map[string]any{}

	for idx := range pr.Tasks {
		st := pr.Tasks[idx]
		sb := builder.NewScriptTaskBuilder(st.Name, "tengo")
		sb.SetId(st.Id)
		data, _ := json.Marshal(st)
		sb.SetProperty("body", data)
		pb.AppendElem(sb.Out())
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
