package ortbbidder

import "github.com/prebid/openrtb/v20/openrtb2"

// Define function types for getters and setters
type GetterFunc func(currNode map[string]any, location []string) (value interface{}, ok bool)
type SetterFunc func(node map[string]any, value interface{})

// Define a struct to hold the getter and setter functions for a parameter
type ParamFuncs struct {
	GetFromORTBFields GetterFunc
	GetFromLocation   GetterFunc
	AutoDetect        GetterFunc
	SetValue          SetterFunc
}

// Define a type for the mapper
type ParamMapper map[string]ParamFuncs

type ParamMapperFactory interface {
	NewBidParamMapper() ParamMapper
	NewRequestParamMapper() ParamMapper
}

type ParamMapperFactoryImpl struct{}

func (f ParamMapperFactoryImpl) NewBidParamMapper() ParamMapper {
	return ParamMapper{
		"mtype": ParamFuncs{
			GetFromORTBFields: func(bid map[string]any, location []string) (value interface{}, ok bool) {
				mType, ok := bid["mtype"].(float64)
				if ok {
					return getMediaTypeForBidFromMType(openrtb2.MarkupType(mType)), true
				}
				return nil, false

			},
			GetFromLocation: func(bid map[string]any, location []string) (value interface{}, ok bool) {
				return getValueFromLocation(bid, location[2:])
			},
			AutoDetect: func(bid map[string]any, location []string) (value interface{}, ok bool) {
				return nil, false
			},
			SetValue: func(typeBid map[string]any, value any) {
				typeBid["BidType"] = value
			},
		},
	}
}

func (f ParamMapperFactoryImpl) NewRequestParamMapper() ParamMapper {
	return ParamMapper{
		"fledge": ParamFuncs{
			GetFromORTBFields: func(bid map[string]any, location []string) (value interface{}, ok bool) {
				return nil, false
			},
			GetFromLocation: func(bid map[string]any, location []string) (value interface{}, ok bool) {
				return nil, false
			},
			AutoDetect: func(bid map[string]any, location []string) (value interface{}, ok bool) {
				return nil, false
			},
			SetValue: func(typeBid map[string]any, value any) {

			},
		},
	}
}

// Method to add a parameter and its associated functions to the ParamMapper
func (pm ParamMapper) AddParam(param string, funcs ParamFuncs) {
	pm[param] = funcs
}

// Method to process a parameter using its associated functions
func (pf ParamFuncs) ProcessParam(currNode, newNode map[string]any, location []string) {
	var (
		value any
		ok    bool
	)
	getters := []GetterFunc{pf.GetFromORTBFields, pf.GetFromLocation, pf.AutoDetect}
	for _, getter := range getters {
		value, ok = getter(currNode, location)
		if ok {
			break
		}
	}
	if !ok {
		return
	}

	pf.SetValue(newNode, value)
}
