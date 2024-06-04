//============================================================================
// Program     : NetAcuity C++ Embedded API
// Author      : Digital Envoy
// Version     : 6.4.1.3
// Date        : 25-Jun-2019
// Copyright   : Copyright 2000-2019, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_FILE_ACCESSOR_H_
#define NA_DB_FILE_ACCESSOR_H_

#include "NaDbBaseAccessor.h"

namespace netacuity {

class NaDbFileAccessor : public NaDbBaseAccessor {

#ifdef _WIN32
#define FileRef HANDLE
#else
#define FileRef int
#endif

public:
	NaDbFileAccessor( int featureCode, std::string directory );
	virtual ~NaDbFileAccessor();

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
	FileRef responseFileDescriptor;
	FileRef rangeIpv4FileDescriptor;
	FileRef rangeIpv6FileDescriptor;
	FileRef extendedRangeIpv4FileDescriptor;
	FileRef extendedRangeIpv6FileDescriptor;

	static FileRef openFile( std::string fileName, size_t fileSize );

	void loadResponses();
	void loadRangesIpv4();
	void loadRangesIpv6();
	void loadExtendedRangesIpv4();
	void loadExtendedRangesIpv6();
};

}
#endif /* NA_DB_FILE_ACCESSOR_H_ */
