package main

type volumeInfo struct {
	name                 string
	minSizeGB            int64
	maxSizeTB            int64
	minDurability        float64
	maxIOPS              int64
	maxIOPSPerGB         int64
	maxThroughput        int64
	pricePerGB           float64
	pIOPSPrice1          float64
	pIOPSPrice1Threshold int64
	pIOPSPrice2          float64
	pIOPSPrice2Threshold int64
	pIOPSPrice3          float64
	pIOPSFree            int64
	throughputPerMBs     float64
	throughputFree       int64
	minIOPS              int64
	baselineIOPSPerGB    int64
	iopsBurst            int64
}

var gp3Info = map[string]volumeInfo{
	"gp2": {
		name:              "gp2",
		minSizeGB:         1,
		maxSizeTB:         16,
		minDurability:     99.8,
		maxIOPS:           16000,
		maxThroughput:     250,
		pricePerGB:        0.1,
		minIOPS:           100,
		baselineIOPSPerGB: 3,
		iopsBurst:         3000,
	},
	"gp3": {
		name:             "gp3",
		minSizeGB:        1,
		maxSizeTB:        16,
		minDurability:    99.8,
		maxIOPS:          16000,
		maxThroughput:    1000,
		pricePerGB:       0.08,
		pIOPSFree:        3000,
		pIOPSPrice1:      0.005,
		throughputPerMBs: 0.04,
		throughputFree:   125,
	},
	"io1": {
		name:          "io1",
		minSizeGB:     4,
		maxSizeTB:     16,
		minDurability: 99.8,
		maxIOPS:       64000,
		maxIOPSPerGB:  50,
		maxThroughput: 1000,
		pricePerGB:    0.125,
		pIOPSPrice1:   0.065,
	},
	"io2": {
		name:                 "io2",
		minSizeGB:            4,
		maxSizeTB:            16,
		minDurability:        99.999,
		maxIOPS:              64000,
		maxIOPSPerGB:         500,
		maxThroughput:        1000,
		pricePerGB:           0.125,
		pIOPSPrice1:          0.065,
		pIOPSPrice1Threshold: 32000,
	},
	"io2be": {
		name:                 "io2be",
		minSizeGB:            4,
		maxSizeTB:            64,
		minDurability:        99.999,
		maxIOPS:              256000,
		maxIOPSPerGB:         1000,
		maxThroughput:        4000,
		pricePerGB:           0.125,
		pIOPSPrice1:          0.065,
		pIOPSPrice1Threshold: 32000,
		pIOPSPrice2:          0.046,
		pIOPSPrice2Threshold: 64000,
		pIOPSPrice3:          0.032,
	},
}
