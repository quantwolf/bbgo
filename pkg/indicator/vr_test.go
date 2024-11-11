package indicator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

/*
python:
import pandas as pd
data = pd.Series([0,1,2,3,4,5,6,7,8,9,0,1,2,3,4,5,6,7,8,9,0,1,2,3,4,5,6,7,8,9,0])
size = 5
rolling_mean = data.rolling(window=size).mean()
vr = data / rolling_mean
print(vr)
*/
func Test_VR(t *testing.T) {
	Delta := 0.001
	var randomVolumes = []byte(`[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 2, 2, 2, 2, 2, 2, 2]`)
	var input []fixedpoint.Value
	if err := json.Unmarshal(randomVolumes, &input); err != nil {
		panic(err)
	}
	tests := []struct {
		name         string
		kLines       []types.KLine
		want         float64
		next         float64
		update       float64
		updateResult float64
		all          int
	}{
		{
			name:         "test",
			kLines:       buildVRKLines(input),
			want:         1.0, // Expected VR value for the last element
			next:         1.0, // Expected VR value for the second last element
			update:       0,
			updateResult: 0.0, // Expected VR value after updating with 0
			all:          28,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := VR{
				IntervalWindow: types.IntervalWindow{Window: 5},
			}

			for _, k := range tt.kLines {
				vr.PushK(k)
			}

			assert.InDelta(t, tt.want, vr.Last(0), Delta)
			assert.InDelta(t, tt.next, vr.Index(1), Delta)
			vr.Update(tt.update)
			assert.InDelta(t, tt.updateResult, vr.Last(0), Delta)
			assert.Equal(t, tt.all, vr.Length())
		})
	}
}

func buildVRKLines(values []fixedpoint.Value) []types.KLine {
	var kLines []types.KLine
	for _, v := range values {
		kLines = append(kLines, types.KLine{
			Close:  v,
			Volume: v,
		})
	}
	return kLines
}