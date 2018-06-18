package http

import "strings"

type methodInfoMapping struct {
	mapping map[string]*methodInfo
	order   []string
}

func newMethodInfoMapping() *methodInfoMapping {
	return &methodInfoMapping{
		mapping: make(map[string]*methodInfo),
		order:   make([]string, 0),
	}
}

func (m *methodInfoMapping) getKey(serviceName, methodName string) string {
	return strings.ToLower(cleanSvcName(serviceName) + ":" + methodName)
}

func (m *methodInfoMapping) Add(serviceName, methodName string, mi *methodInfo) {
	if mi != nil {
		key := m.getKey(serviceName, methodName)
		if _, ok := m.mapping[key]; !ok {
			m.order = append(m.order, key)
		}
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

func (m *methodInfoMapping) GetAllMethodInfo() []*methodInfo {
	methods := make([]*methodInfo, 0)
	for _, value := range m.mapping {
		methods = append(methods, value)
	}
	return methods
}

func (m *methodInfoMapping) GetAllMethodInfoByOrder() []*methodInfo {
	methods := make([]*methodInfo, 0)
	for _, key := range m.order {
		if mi, ok := m.mapping[key]; ok {
			methods = append(methods, mi)
		}
	}
	return methods
}
