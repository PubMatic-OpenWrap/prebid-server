//============================================================================
// Program     : C++ NetAcuity Embedded API
// Author      : Digital Envoy
// Version     : 7.0.0.1
// Date        : 2023-DEC-08
// Copyright   : Copyright 2000-2023, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_UTIL_H_
#define NA_DB_UTIL_H_

#ifdef _WIN32
#include <winsock2.h>
#include <windows.h>
#include <Inaddr.h>
#include <In6addr.h>
#else
#include <arpa/inet.h>
#endif

#include "NaDbDef.h"

namespace netacuity {


	u_int32_t getIpv4Num( in_addr ipv4 );

	u_int64_t getIpv6NetworkNum( in6_addr ipv6 );

	u_int64_t getIpv6InterfaceNum( in6_addr ipv6 );

	std::string toString( StringList list );

	std::string toString( ResponseMap map );

	std::string getFieldValue( const ResponseMap &responseMap, const std::string &fieldName );

}
#endif /* NA_DB_UTIL_H_ */
