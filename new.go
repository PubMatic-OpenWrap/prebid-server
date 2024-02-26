package main_ow

func groupAnagrams(strs []string) [][]string {

	anagramMap := make(map[[26]int][]string)

	for _, str := range strs {
		var key [26]int
		for _, c := range str {
			key[c-'a']++
		}
		anagramMap[key] = append(anagramMap[key], str)
	}

	var result [][]string

	for _, v := range anagramMap {
		result = append(result, v)
	}

	return result

}
