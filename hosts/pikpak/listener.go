package pikpak

import (
	"github.com/KeepShareOrg/keepshare/hosts"
)

func (p *PikPak) AddEventListener(event hosts.EventType, callback hosts.ListenerCallback) {
	p.api.AddEventListener(event, callback)
}
