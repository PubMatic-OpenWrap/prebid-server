//============================================================================
// Program     : C++ NetAcuity Embedded API
// Author      : Digital Envoy
// Version     : 7.0.0.1
// Date        : 2023-DEC-08
// Copyright   : Copyright 2000-2023, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_PARSER_H_
#define NA_DB_PARSER_H_

#include "NaDbAccessor.h"

namespace netacuity {

typedef struct AccessorInfo_struct {
	AccessorInfo_struct() : accessor(), fieldNameList() {}
	NaDbAccessor* accessor;
	StringList fieldNameList;
} AccessorInfo;

typedef std::map<int, AccessorInfo> AccessorInfoMap;

class NaDbParser {

public:
	NaDbParser(  );
	virtual ~NaDbParser();

	/// Load the specified accessor.
	void loadAccessor( NaDbAccessor &accessor );

	/// Return the accessor loaded for the specified featureCode.
	const NaDbAccessor* getAccessor( int featureCode ) const;

	/// Retrieve the mapped response for the specified IPv4 number.
	ResponseMap queryMappedResponseIpv4( u_int32_t ipNum ) const;

	/// Retrieve the mapped response for the specified IPv6 network number.
	/// The network number is the half of the IPv6 address which contains the most significant-bits.
	ResponseMap queryMappedResponseIpv6( u_int64_t numNetwork ) const;

	/// Retrieve the mapped response for the specified IPv6 full network and interface numbers.
	/// The network number is the half of the IPv6 address which contains the most significant bits.
	/// The interface number is the half of the IPv6 address which contains the least significant bits.
	ResponseMap queryMappedResponseIpv6( u_int64_t numNetwork, u_int64_t numInterface ) const;

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
