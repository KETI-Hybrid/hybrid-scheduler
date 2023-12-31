/*
 * Copyright © 2021 peizhaoyou <peizhaoyou@4paradigm.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scheduler

import (
	"fmt"
	"sync"
)

type NodeInfo struct {
	ID string
}

type nodeManager struct {
	nodes map[string]*NodeInfo
	mutex sync.Mutex
}

func (m *nodeManager) init() {
	m.nodes = make(map[string]*NodeInfo)
}

func (m *nodeManager) addNode(nodeID string, nodeInfo *NodeInfo) {
	if nodeInfo == nil {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.nodes[nodeID] = nodeInfo
}

func (m *nodeManager) GetNode(nodeID string) (*NodeInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if n, ok := m.nodes[nodeID]; ok {
		return n, nil
	}
	return &NodeInfo{}, fmt.Errorf("node %v not found", nodeID)
}

func (m *nodeManager) ListNodes() (map[string]*NodeInfo, error) {
	return m.nodes, nil
}
