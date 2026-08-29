//============================================================================
// Program     : C++ NetAcuity Embedded API
// Author      : Digital Envoy
// Version     : 7.0.0.1
// Date        : 2023-DEC-08
// Copyright   : Copyright 2000-2023, Digital Envoy, Inc.  All rights reserved.
//============================================================================

#ifndef NA_DB_RESPONSE_ITERATOR_H_
#define NA_DB_RESPONSE_ITERATOR_H_

#include "NaDbUtil.h"

namespace netacuity {

typedef struct FieldInfo_struct {
	FieldInfo_struct() : fieldName(), fieldValue() {}
	std::string fieldName;
	std::string fieldValue;
} FieldInfo;

class NaDbResponseIterator {

public:
	NaDbResponseIterator( const ResponseMap &responseMap );
	virtual ~NaDbResponseIterator();

	bool hasNext();
	FieldInfo next();

private:
	ResponseMap::const_iterator iter;
	ResponseMap::const_iterator end;
};

NaDbResponseIterator getIterator( const ResponseMap &responseMap );

}
#endif /* NA_DB_RESPONSE_ITERATOR_H_ */
