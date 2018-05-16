package http

import "strings"

type methodInfoMapping struct {
	mapping map[string]*methodInfo
}

func newPaths() *methodInfoMapping {
	return &methodInfoMapping{
		mapping: make(map[string]*methodInfo),
	}
}

func (m *methodInfoMapping) getKey(serviceName, methodName string) string {
	return strings.ToLower(cleanSvcName(serviceName) + ":" + methodName)
}

func (m *methodInfoMapping) Add(serviceName, methodName string, mi *methodInfo) {
	if mi != nil {
		key := m.getKey(serviceName, methodName)
		m.mapping[key] = mi
	}
}

func (m *methodInfoMapping) Delete(serviceName, methodName string) {
	key := m.getKey(serviceName, methodName)
	delete(m.mapping, key)
}

func (m *methodInfoMapping) Get(serviceName, methodName string) (*methodInfo, bool) {
	key := m.getKey(serviceName, methodName)
	mi, ok := m.mapping[key]
	return mi, ok
}

func (m *methodInfoMapping) GetAllPathInfo() []*methodInfo {
	paths := make([]*methodInfo, 0)
	for _, value := range m.mapping {
		paths = append(paths, value)
	}
	return paths
}
