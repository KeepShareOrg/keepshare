package pikpak

import (
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
)

func (p *PikPak) AddEventListener(event comm.PPEventType, callback comm.ListenerCallback) {
	p.api.AddEventListener(event, callback)
}
