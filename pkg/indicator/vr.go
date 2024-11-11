package indicator

import (
	"time"

	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/types"
)

const MaxNumOfVR = 5_000
const MaxNumOfVRTruncateSize = 100

//go:generate callbackgen -type VR
type VR struct {
	types.SeriesBase
	types.IntervalWindow
	Values    floats.Slice
	rawValues *types.Queue
	EndTime   time.Time

	UpdateCallbacks []func(value float64)
}

func (vr *VR) Last(i int) float64 {
	return vr.Values.Last(i)
}

func (vr *VR) Index(i int) float64 {
	return vr.Last(i)
}

func (vr *VR) Length() int {
	return vr.Values.Length()
}

func (vr *VR) Clone() types.UpdatableSeriesExtend {
	out := &VR{
		Values:    vr.Values[:],
		rawValues: vr.rawValues.Clone(),
		EndTime:   vr.EndTime,
	}
	out.SeriesBase.Series = out
	return out
}

var _ types.SeriesExtend = &VR{}

func (vr *VR) Update(volume float64) {
	if vr.rawValues == nil {
		vr.rawValues = types.NewQueue(vr.Window)
		vr.SeriesBase.Series = vr
	}

	vr.rawValues.Update(volume)
	if vr.rawValues.Length() < vr.Window {
		return
	}

	averageVolume := types.Mean(vr.rawValues)
	vratio := volume / averageVolume
	vr.Values.Push(vratio)
	if len(vr.Values) > MaxNumOfVR {
		vr.Values = vr.Values[MaxNumOfVRTruncateSize-1:]
	}
}

func (vr *VR) BindK(target KLineClosedEmitter, symbol string, interval types.Interval) {
	target.OnKLineClosed(types.KLineWith(symbol, interval, vr.PushK))
}

func (vr *VR) PushK(k types.KLine) {
	if vr.EndTime != zeroTime && k.EndTime.Before(vr.EndTime) {
		return
	}

	vr.Update(k.Volume.Float64())
	vr.EndTime = k.EndTime.Time()
	vr.EmitUpdate(vr.Values.Last(0))
}

func (vr *VR) LoadK(allKLines []types.KLine) {
	for _, k := range allKLines {
		vr.PushK(k)
	}
}