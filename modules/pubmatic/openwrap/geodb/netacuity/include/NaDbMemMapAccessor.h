//============================================================================
// Program     : NetAcuity C++ Embedded API
// Author      : Digital Envoy
// Version     : 6.4.1.3
// Date        : 25-Jun-2019
// Copyright   : Copyright 2000-2019, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_MEM_MAP_ACCESSOR_H_
#define NA_DB_MEM_MAP_ACCESSOR_H_

#include "NaDbBaseAccessor.h"


namespace netacuity {

class NaDbMemMapAccessor : public NaDbBaseAccessor {

public:
	NaDbMemMapAccessor( int featureCode, std::string directory );
	virtual ~NaDbMemMapAccessor();

protected:
	// override
	std::string getRawResponse( unsigned long offset) const;
	// override
	RangeRecordIpv4 getRangeRecordIpv4( unsigned int index ) const;
	// override
	RangeRecordIpv6 getRangeRecordIpv6( unsigned int index ) const;
	// override
	RangeRecordIpv6NetworkOnly getRangeRecordIpv6NetworkOnly( unsigned int index ) const;
	// override
	ExtendedRangeRecordIpv4 getExtendedRangeRecordIpv4( unsigned int index ) const;
	// override
	ExtendedRangeRecordIpv6 getExtendedRangeRecordIpv6( unsigned int index ) const;
	// override
	ExtendedRangeRecordIpv6NetworkOnly getExtendedRangeRecordIpv6NetworkOnly( unsigned int index ) const;

private:
	// the memory-mapped data for the responses
	char* responseData;

	// the memory-mapped data for the IPv4 ranges
	RangeRecordIpv4* rangeDataIpv4;
	ExtendedRangeRecordIpv4* extendedRangeDataIpv4;

	// the memory-mapped data for the IPv6 ranges
	RangeRecordIpv6* rangeDataIpv6;
	RangeRecordIpv6NetworkOnly* rangeDataIpv6NetworkOnly;
	ExtendedRangeRecordIpv6* extendedRangeDataIpv6;
	ExtendedRangeRecordIpv6NetworkOnly* extendedRangeDataIpv6NetworkOnly;

	static void* loadMemoryMappedFile( std::string fileName, size_t fileSize );
	static void unloadMemoryMappedFile( std::string fileName, void* memMappedData, size_t fileSize );

	void loadResponses();
	void loadRangesIpv4();
	void loadRangesIpv6();
	void loadExtendedRangesIpv4();
	void loadExtendedRangesIpv6();
};

}
#endif /* NA_DB_MEM_MAP_ACCESSOR_H_ */
