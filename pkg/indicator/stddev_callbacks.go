// Code generated by "callbackgen -type StdDev"; DO NOT EDIT.

package indicator

import ()

func (inc *StdDev) OnUpdate(cb func(value float64)) {
	inc.updateCallbacks = append(inc.updateCallbacks, cb)
}

func (inc *StdDev) EmitUpdate(value float64) {
	for _, cb := range inc.updateCallbacks {
		cb(value)
	}
}