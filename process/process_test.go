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

package process

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPlayBook_UnmarshalYAML(t *testing.T) {
	text := `---
- name: this is a test playbook
  hosts: webservers
  vars:
    http_port: 80
    max_clients: 200
  remote_user: root
  tasks:
  - name: ensure apache is at the latest version
    yum: pkg=httpd state=latest
  - name: child process
    kind: process
    tasks:
    - name: first child task
      ping: 
  - name: write the apache config file
    kind: service
    template:
      src: /srv/httpd.j2
      dest: /etc/httpd.conf
    notify:
    - restart apache
  - name: ensure apache is running
    action: service
    args:
      name: httpd 
      state: started
      languages:
        - Go
        - Javascript
      size:
        height: 200px
        width: 100px
  handlers:
    - name: restart apache
      service: name=httpd state=restarted`

	pr := &Process{}
	err := yaml.Unmarshal([]byte(text), pr)
	if err != nil {
		t.Fatal(err)
	}

	var data []byte
	data, err = yaml.Marshal(pr)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}
