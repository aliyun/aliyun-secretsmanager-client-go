package service

import "sync"

var UserAgentManager = &userAgentManager{
	priority: -1,
}

type userAgentManager struct {
	userAgent      string
	projectVersion string
	priority       int

	mtx sync.Mutex
}

func (m *userAgentManager) RegisterUserAgent(userAgent string, priority int, projectVersion string) {
	if priority > m.priority {
		m.mtx.Lock()
		if priority > m.priority {
			m.userAgent = userAgent
			m.projectVersion = projectVersion
			m.priority = priority
		}
		m.mtx.Unlock()
	}
}

func (m *userAgentManager) GetUserAgent() string {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.userAgent
}

func (m *userAgentManager) GetProjectVersion() string {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.projectVersion
}
