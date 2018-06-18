package handlers

import (
	"fmt"
	"strings"
	"sync"
)

func NewMiddlewareMapping() *MiddlewareMapping {
	return &MiddlewareMapping{}
}

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

func (m *MiddlewareMapping) GetMiddlewaresFromUrl(url string) []string {
	key := m.getKeyFromURL(url)
	return m.getMiddleware(key)
}

func (m *MiddlewareMapping) GetMiddlewares(service, method string) []string {
	key := m.getKey(service, method)
	return m.getMiddleware(key)
}

func (m *MiddlewareMapping) getMiddleware(key string) []string {
	fmt.Println("fetching middlewares", key)
	if result, ok := m.mapping.Load(key); ok {
		return result.([]string)
	}
	return []string{}
}

func (m *MiddlewareMapping) AddMiddleware(service, method string, middlewares ...string) {
	key := m.getKey(service, method)
	m.addMiddleware(key, middlewares...)
}

func (m *MiddlewareMapping) addMiddleware(key string, middlewares ...string) {
	fmt.Println("adding middlewares", key, middlewares)
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
