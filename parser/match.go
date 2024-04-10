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

package parser

import "path"

// MatchHosts looks for hosts that match the pattern
func (dl *DataLoader) MatchHosts(pattern string) (map[string]*Host, error) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return MatchHosts(dl.Hosts, pattern)
}

// MatchHosts looks for hosts that match the pattern
func (group *Group) MatchHosts(pattern string) (map[string]*Host, error) {
	return MatchHosts(group.Hosts, pattern)
}

// MatchHosts looks for hosts that match the pattern
func MatchHosts(hosts map[string]*Host, pattern string) (map[string]*Host, error) {
	matchedHosts := make(map[string]*Host)
	for _, host := range hosts {
		m, err := path.Match(pattern, host.Name)
		if err != nil {
			return nil, err
		}
		if m {
			matchedHosts[host.Name] = host
		}
	}
	return matchedHosts, nil
}

// MatchGroups looks for groups that match the pattern
func (dl *DataLoader) MatchGroups(pattern string) (map[string]*Group, error) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return MatchGroups(dl.Groups, pattern)
}

// MatchGroups looks for groups that match the pattern
func (host *Host) MatchGroups(pattern string) (map[string]*Group, error) {
	return MatchGroups(host.Groups, pattern)
}

// MatchGroups looks for groups that match the pattern
func MatchGroups(groups map[string]*Group, pattern string) (map[string]*Group, error) {
	matchedGroups := make(map[string]*Group)
	for _, group := range groups {
		m, err := path.Match(pattern, group.Name)
		if err != nil {
			return nil, err
		}
		if m {
			matchedGroups[group.Name] = group
		}
	}
	return matchedGroups, nil
}

// MatchVars looks for vars that match the pattern
func (group *Group) MatchVars(pattern string) (map[string]string, error) {
	return MatchVars(group.Vars, pattern)
}

// MatchVars looks for vars that match the pattern
func (host *Host) MatchVars(pattern string) (map[string]string, error) {
	return MatchVars(host.Vars, pattern)
}

// MatchVars looks for vars that match the pattern
func MatchVars(vars map[string]string, pattern string) (map[string]string, error) {
	matchedVars := make(map[string]string)
	for k, v := range vars {
		m, err := path.Match(pattern, v)
		if err != nil {
			return nil, err
		}
		if m {
			matchedVars[k] = v
		}
	}
	return matchedVars, nil
}
