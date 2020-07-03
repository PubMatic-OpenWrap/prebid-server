package testdata

type eout struct {
	MaximizeForDuration [][2]int64
	MinMaxAlgorithm     [][2]int64
}

// Scenario returns expected impression breaks for given algorithm and for given
// test scenario
var Scenario = map[string]eout{

	"TC2": {
		MaximizeForDuration: [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}},
		MinMaxAlgorithm:     [][2]int64{{11, 13}, {11, 11}, {11, 11}, {11, 11}, {11, 11}, {11, 11}, {11, 11}, {11, 11}, {15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}}},

	"TC3": {
		MaximizeForDuration: [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}},
		MinMaxAlgorithm:     [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}},
	},
	"TC4": {
		MaximizeForDuration: [][2]int64{{15, 15}},
		MinMaxAlgorithm:     [][2]int64{{1, 15}, {1, 1}},
	},
	"TC5": {
		MaximizeForDuration: [][2]int64{{10, 10}, {5, 5}},
		MinMaxAlgorithm:     [][2]int64{{1, 1}, {1, 5}, {1, 15}, {1, 10}},
	},
	"TC6": {
		MaximizeForDuration: [][2]int64{{15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {1, 1}},
	},
	"TC7": {
		MaximizeForDuration: [][2]int64{{15, 15}},
		MinMaxAlgorithm:     [][2]int64{{15, 15}},
	},
	"TC8": {
		MaximizeForDuration: [][2]int64{{15, 15}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{10, 10}, {10, 10}, {15, 15}},
	},
	"TC9": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC10": {
		MaximizeForDuration: [][2]int64{{15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {15, 15}},
	},
	"TC11": {
		MaximizeForDuration: [][2]int64{{9, 11}, {9, 9}, {9, 9}, {9, 9}, {9, 9}, {9, 9}, {9, 9}},
		MinMaxAlgorithm:     [][2]int64{{9, 11}, {9, 9}, {9, 9}, {9, 9}, {9, 9}, {9, 9}, {9, 9}},
	},
	"TC12": {
		MaximizeForDuration: [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {20, 20}, {20, 20}, {15, 15}, {15, 15}, {15, 15}, {15, 15}},
	},
	"TC13": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC14": {
		MaximizeForDuration: [][2]int64{{5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}},
		MinMaxAlgorithm:     [][2]int64{{5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}},
	},
	"TC15": {
		MaximizeForDuration: [][2]int64{{5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}},
		MinMaxAlgorithm:     [][2]int64{{6, 9}, {6, 9}, {6, 9}, {6, 9}, {6, 9}, {5, 6}, {5, 5}, {5, 5}, {5, 5}},
	},
	"TC16": {
		MaximizeForDuration: [][2]int64{{1, 12}, {1, 12}, {1, 12}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {6, 6}, {12, 12}, {12, 12}, {12, 12}},
	},
	"TC17": {
		MaximizeForDuration: [][2]int64{{1, 12}, {1, 12}, {1, 12}, {1, 12}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{1, 11}, {6, 7}, {4, 12}, {4, 12}, {4, 12}, {4, 12}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
	},
	"TC18": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC19": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC20": {
		MaximizeForDuration: [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
	},
	"TC21": {
		MaximizeForDuration: [][2]int64{{3, 9}, {3, 9}, {3, 9}, {3, 9}, {3, 9}, {3, 9}, {3, 9}, {3, 9}, {3, 9}},
		MinMaxAlgorithm:     [][2]int64{{5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}, {5, 9}},
	},
	"TC23": {
		MaximizeForDuration: [][2]int64{{4, 14}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
		MinMaxAlgorithm:     [][2]int64{{4, 13}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {4, 14}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},
	},
	"TC24": {
		MaximizeForDuration: [][2]int64{{60, 69}, {65, 65}},
		MinMaxAlgorithm:     [][2]int64{{60, 69}, {65, 65}},
	},
	"TC25": {
		MaximizeForDuration: [][2]int64{{1, 68}, {20, 20}},
		MinMaxAlgorithm:     [][2]int64{{1, 68}, {20, 20}},
	},
	"TC26": {
		MaximizeForDuration: [][2]int64{{45, 45}, {45, 45}},
		MinMaxAlgorithm:     [][2]int64{{45, 45}, {45, 45}},
	},
	"TC27": {
		MaximizeForDuration: [][2]int64{{30, 30}, {30, 30}, {30, 30}},
		MinMaxAlgorithm:     [][2]int64{{2, 3}, {2, 2}, {30, 30}, {30, 30}, {30, 30}, {3, 45}, {3, 45}},
	},
	"TC28": {
		MaximizeForDuration: [][2]int64{{30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}},
		MinMaxAlgorithm:     [][2]int64{{3, 90}, {3, 90}, {3, 3}, {2, 2}, {30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}},
	},
	"TC29": {
		MaximizeForDuration: [][2]int64{{25, 25}, {20, 20}, {20, 20}},
		MinMaxAlgorithm:     [][2]int64{{5, 25}, {3, 20}, {3, 20}, {3, 3}, {2, 2}, {5, 35}, {5, 30}},
	},
	"TC30": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC31": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC32": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC33": {
		MaximizeForDuration: [][2]int64{{30, 42}, {35, 35}, {35, 35}, {35, 35}},
		MinMaxAlgorithm:     [][2]int64{{30, 42}, {35, 35}, {35, 35}, {35, 35}},
	},
	"TC34": {
		MaximizeForDuration: [][2]int64{{30, 30}, {30, 30}, {30, 30}},
		MinMaxAlgorithm:     [][2]int64{{30, 30}, {30, 30}, {30, 30}},
	},
	"TC35": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC36": {
		MaximizeForDuration: [][2]int64{{45, 45}, {45, 45}},
		MinMaxAlgorithm:     [][2]int64{{45, 45}, {45, 45}},
	},
	"TC37": {
		MaximizeForDuration: [][2]int64{{25, 25}, {20, 20}},
		MinMaxAlgorithm:     [][2]int64{{25, 25}, {20, 20}},
	},
	"TC38": {
		MaximizeForDuration: [][2]int64{{25, 25}, {25, 25}, {20, 20}, {20, 20}},
		MinMaxAlgorithm:     [][2]int64{{25, 25}, {25, 25}, {20, 20}, {20, 20}, {45, 45}, {45, 45}},
	},
	"TC39": {
		MaximizeForDuration: [][2]int64{{25, 25}, {25, 25}, {20, 20}, {20, 20}},
		MinMaxAlgorithm:     [][2]int64{{30, 45}, {30, 45}, {30, 30}, {30, 30}, {25, 25}, {25, 25}, {20, 20}, {20, 20}},
	},
	"TC40": {
		MaximizeForDuration: [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {5, 5}},
		MinMaxAlgorithm:     [][2]int64{{30, 30}, {30, 30}, {30, 45}, {30, 45}, {25, 25}, {25, 25}, {20, 20}, {20, 20}},
	},
	"TC41": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC42": {
		MaximizeForDuration: [][2]int64{{1, 1}},
		MinMaxAlgorithm:     [][2]int64{{1, 1}},
	},
	"TC43": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC44": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC45": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC46": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC47": {
		MaximizeForDuration: [][2]int64{{6, 6}},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC48": {
		MaximizeForDuration: [][2]int64{{6, 6}, {6, 6}},
		MinMaxAlgorithm:     [][2]int64{{6, 6}, {6, 6}},
	},
	"TC49": {
		MaximizeForDuration: [][2]int64{},
		MinMaxAlgorithm:     [][2]int64{},
	},
	"TC50": {
		MaximizeForDuration: [][2]int64{{1, 1}},
		MinMaxAlgorithm:     [][2]int64{{1, 1}},
	},
	"TC51": {
		MaximizeForDuration: [][2]int64{{13, 13}, {13, 13}, {13, 13}},
		MinMaxAlgorithm:     [][2]int64{{13, 13}, {13, 13}, {13, 13}},
	},
	"TC52": {
		MaximizeForDuration: [][2]int64{{12, 18}, {12, 18}, {12, 18}, {12, 18}},
		MinMaxAlgorithm:     [][2]int64{{12, 17}, {15, 15}, {17, 18}, {17, 18}, {17, 18}, {17, 18}},
	},
	"TC53": {
		MaximizeForDuration: [][2]int64{{20, 20}, {20, 20}, {20, 20}, {20, 20}, {20, 20}, {20, 20}, {1, 6}},
		MinMaxAlgorithm:     [][2]int64{{1, 6}, {20, 20}, {20, 20}, {20, 20}, {20, 20}, {20, 20}, {20, 20}},
	},
}
