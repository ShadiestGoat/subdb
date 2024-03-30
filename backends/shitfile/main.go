package shitfile

import (
	"sync"

	"shadygoat.eu/shitdb"
	"shadygoat.eu/shitdb/backends/all"
)

type ShitFile[IDType shitdb.IDConstraint] struct {
	flush all.AllBackend[IDType]
	
	
	lock sync.RWMutex
}

func (r *ShitFile[IDType]) ReadIDs(inp ...IDType) {
	r.lock.RLock()
	defer r.lock.RUnlock()
}
