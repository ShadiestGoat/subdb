package file

import (
	"sync"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/all"
)

type File[IDType subdb.IDConstraint] struct {
	flush all.AllBackend[IDType]
	
	
	lock sync.RWMutex
}

func (r *File[IDType]) ReadIDs(inp ...IDType) {
	r.lock.RLock()
	defer r.lock.RUnlock()
}
