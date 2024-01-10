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

package executor

const (
	connectVars   = "bee_connect"
	hostVars      = "bee_host"
	portVars      = "bee_part"
	userVars      = "bee_user"
	sshPasswdVars = "bee_ssh_passwd"
	sshPrivateKey = "bee_ssh_private_key"
	sshPassphrase = "bee_ssh_passphrase"

	winRMPasswdVars = "bee_winrm_passwd"
)

//type connectVars struct {
//	User   string
//	Passwd string
//}
//
//type WinRVars struct {
//	RemoteUser string
//}
//
//type GRPCVars struct {
//}
//
//func getSSHVars(host *parser.Host) *SSHVars {
//
//}
