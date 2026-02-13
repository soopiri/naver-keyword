package main

import "math"

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

type TrendSeriesItem struct {
	Value float64 `json:"value"`
	Ratio float64 `json:"ratio"`
}

func normalizeTrend(series []TrendSeriesItem) float64 {
	if len(series) == 0 {
		return 0
	}

	max := 0.0
	for _, item := range series {
		val := item.Value
		if val == 0 {
			val = item.Ratio
		}
		if val > max {
			max = val
		}
	}

	if max <= 0 {
		return 0
	}

	last := series[len(series)-1]
	v := last.Value
	if v == 0 {
		v = last.Ratio
	}

	return clamp(v/max, 0, 1)
}

type ScoreParams struct {
	SearchPc      float64
	SearchMobile  float64
	DocCount      float64
	RecentDocCount float64
	TrendSeries   []TrendSeriesItem
	Params        struct {
		WPC   float64
		WMob  float64
		K     float64
		Alpha float64
		T     float64
		R     float64
	}
}

func CalculateScore(params ScoreParams) ScoreResult {
	wPC := params.Params.WPC
	if wPC == 0 {
		wPC = 0.4
	}
	wMob := params.Params.WMob
	if wMob == 0 {
		wMob = 0.6
	}
	k := params.Params.K
	if k == 0 {
		k = 1.0
	}
	alpha := params.Params.Alpha
	if alpha == 0 {
		alpha = 0.5
	}
	t := params.Params.T
	if t == 0 {
		t = 0.2
	}
	r := params.Params.R
	if r == 0 {
		r = 0.3
	}

	S_pc := params.SearchPc
	S_mob := params.SearchMobile
	D := params.DocCount
	D_recent := params.RecentDocCount

	S := S_pc + S_mob
	S_w := wPC*S_pc + wMob*S_mob

	var score4 float64
	if S+k*D == 0 {
		score4 = 0
	} else {
		score4 = (100 * S) / (S + k*D)
	}

	var score5 float64
	if S_w+k*D == 0 {
		score5 = 0
	} else {
		score5 = (100 * S_w) / (S_w + k*D)
	}

	score := alpha*score4 + (1-alpha)*score5

	trendNorm := normalizeTrend(params.TrendSeries)
	recentRatio := D_recent / (D + 1)

	scoreFinal := clamp(score*(1+t*trendNorm)*(1-r*recentRatio), 0, 100)

	return ScoreResult{
		Score:       roundTo(scoreFinal, 2),
		ScoreRaw:    roundTo(score, 2),
		Score4:      roundTo(score4, 2),
		Score5:      roundTo(score5, 2),
		TrendNorm:   roundTo(trendNorm, 3),
		RecentRatio: roundTo(recentRatio, 3),
	}
}

func roundTo(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

type VolumeGroup struct {
	ID    string
	Label string
}

func getVolumeGroup(totalSearch int) *VolumeGroup {
	if totalSearch >= 100000 {
		return &VolumeGroup{ID: "g4", Label: "10만 이상"}
	}
	if totalSearch >= 50000 {
		return &VolumeGroup{ID: "g3", Label: "5만 ~ 10만"}
	}
	if totalSearch >= 10000 {
		return &VolumeGroup{ID: "g2", Label: "1만 ~ 5만"}
	}
	if totalSearch >= 1000 {
		return &VolumeGroup{ID: "g1", Label: "1천 ~ 1만"}
	}
	return nil
}
