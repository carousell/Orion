package handlers

import (
	"strings"
	"sync"
)

//NewMiddlewareMapping returns a new MiddlewareMapping
func NewMiddlewareMapping() *MiddlewareMapping {
	return &MiddlewareMapping{}
}

//MiddlewareMapping stores mapping between service,method and middlewares
type MiddlewareMapping struct {
	mapping sync.Map
}

func (m *MiddlewareMapping) getKey(service, method string) string {
	return strings.ToLower(cleanSvcName(service) + ":" + method)
}

func (m *MiddlewareMapping) getKeyFromURL(url string) string {
	parts := strings.SplitN(strings.TrimPrefix(url, "/"), "/", 2)
	if len(parts) > 1 {
		return m.getKey(parts[0], parts[1])
	}
	return ""
}

//GetMiddlewaresFromURL fetches all middleware for a specific URL
func (m *MiddlewareMapping) GetMiddlewaresFromURL(url string) []string {
	key := m.getKeyFromURL(url)
	return m.getMiddleware(key)
}

//GetMiddlewares fetches all middlewares for a specific service,method
func (m *MiddlewareMapping) GetMiddlewares(service, method string) []string {
	key := m.getKey(service, method)
	return m.getMiddleware(key)
}

func (m *MiddlewareMapping) getMiddleware(key string) []string {
	if result, ok := m.mapping.Load(key); ok {
		return result.([]string)
	}
	return []string{}
}

//AddMiddleware adds middleware to a service, method
func (m *MiddlewareMapping) AddMiddleware(service, method string, middlewares ...string) {
	key := m.getKey(service, method)
	m.addMiddleware(key, middlewares...)
}

func (m *MiddlewareMapping) addMiddleware(key string, middlewares ...string) {
	list := make([]string, 0)
	if result, ok := m.mapping.Load(key); ok {
		list = append(list, result.([]string)...)
	}
	list = append(list, middlewares...)
	m.mapping.Store(key, list)
}

func cleanSvcName(serviceName string) string {
	serviceName = strings.ToLower(serviceName)
	parts := strings.Split(serviceName, ".")
	if len(parts) > 1 {
		serviceName = parts[1]
	}
	return serviceName
}
