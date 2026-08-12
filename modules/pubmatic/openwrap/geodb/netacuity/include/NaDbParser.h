//============================================================================
// Program     : NetAcuity C++ Embedded API
// Author      : Digital Envoy
// Version     : 6.4.1.3
// Date        : 25-Jun-2019
// Copyright   : Copyright 2000-2019, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_PARSER_H_
#define NA_DB_PARSER_H_

#include "NaDbBaseAccessor.h"

namespace netacuity {

typedef struct AccessorInfo_struct {
	AccessorInfo_struct() : accessor(), fieldNameList() {}
	NaDbBaseAccessor* accessor;
	StringList fieldNameList;
} AccessorInfo;

typedef std::map<int, AccessorInfo> AccessorInfoMap;

class NaDbParser {

public:
	NaDbParser(  );
	virtual ~NaDbParser();

	/// Load the specified accessor.
	void loadAccessor( NaDbBaseAccessor &accessor );

	/// Return the accessor loaded for the specified featureCode.
	const NaDbBaseAccessor* getAccessor( int featureCode ) const;

	/// Retrieve the mapped response for the specified IPv4 number.
	ResponseMap queryMappedResponseIpv4( unsigned long ipNum ) const;

	/// Retrieve the mapped response for the specified IPv6 network number.
	/// The network number is the half of the IPv6 address which contains the most significant-bits.
	ResponseMap queryMappedResponseIpv6( unsigned long long numNetwork ) const;

	/// Retrieve the mapped response for the specified IPv6 full network and interface numbers.
	/// The network number is the half of the IPv6 address which contains the most significant bits.
	/// The interface number is the half of the IPv6 address which contains the least significant bits.
	ResponseMap queryMappedResponseIpv6( unsigned long long numNetwork, unsigned long long numInterface ) const;

	/// Retrieve the mapped response for the specified IPv4 address.
	ResponseMap queryMappedResponseIpv4( in_addr ipv4 ) const;

	/// Retrieve the mapped response for the specified IPv6 address.
	ResponseMap queryMappedResponseIpv6( in6_addr ipv6 ) const;

	/// Retrieve the mapped response for the specified IP "dotted" presentation address, IPv4 or IPv6.
	ResponseMap queryMappedResponse( const char* ipAddress ) const;

	/// Retrieve the mapped response for the specified IP "dotted" presentation address, IPv4 or IPv6.
	ResponseMap queryMappedResponse( std::string ipAddress ) const;


private:
	static const int MAX_FEATURE_CODE_COUNT = 100;

	AccessorInfoMap accessorInfoMap;

	ResponseMap getDefaultMappedResponse() const;
	static void addMappedResponse( ResponseMap &destMap, const StringList &fieldList, const std::string &rawResponse );

};

}
#endif /* NA_DB_PARSER_H_ */
