package tagbidder

type macroCallBack struct {
	cached   bool
	callback func(IBidderMacro, string) string
}

//Mapper will map macro with its respective call back function
type Mapper map[string]*macroCallBack

func (obj Mapper) clone() Mapper {
	cloned := make(Mapper, len(obj))
	for k, v := range obj {
		newCallback := *v
		cloned[k] = &newCallback
	}
	return cloned
}

//NewMapperFromConfig returns new Mapper from JSON details
func NewMapperFromConfig(config *BidderConfig) Mapper {
	newMapper := GetNewDefaultMapper()
	for macro, key := range config.Keys {
		macroCB, ok := newMapper[macro]
		if !ok {
			//create new entry if not present
			macroCB = &macroCallBack{cached: false, callback: IBidderMacro.ConstantValue}
			newMapper[macro] = macroCB
		}

		//default definition
		switch key.ValueType {
		case JSONKeyValueType: /*json key value*/
			macroCB.callback = IBidderMacro.JSONKey
		case ConstantValueType: /*constant*/
			macroCB.callback = IBidderMacro.ConstantValue
		}

		//Cache Key
		if nil != key.Cached && *key.Cached {
			macroCB.cached = true
		}
	}
	return newMapper
}

/*
//SetCache value to specific key
func (obj *Mapper) SetCache(key string, value bool) {
	if value, ok := (*obj)[key]; ok {
		value.cached = true
	}
}

//AddCustomMacro for adding custom macro whose definition will be present in IBidderMacro.Custom method
func (obj *Mapper) AddCustomMacro(key string, isCached bool) {
	(*obj)[key] = &macroCallBack{cached: isCached, callback: IBidderMacro.Custom}
}*/
