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
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPlayBook_UnmarshalYAML(t *testing.T) {
	text := `
name: this is a test playbook
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
    action: ping 
- name: write the apache config file
  kind: service
  template:
    src: /srv/httpd.j2
    dest: /etc/httpd.conf
  catch:
    name: doing it while failed
    service: name=httpd state=restarted
  finish:
    name: always do it
    service: name=httpd state=restarted 
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
