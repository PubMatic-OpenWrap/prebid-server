//============================================================================
// Program     : NetAcuity C++ Embedded API
// Author      : Digital Envoy
// Version     : 6.4.1.3
// Date        : 25-Jun-2019
// Copyright   : Copyright 2000-2019, Digital Envoy, Inc.  All rights reserved.
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


	unsigned long getIpv4Num( in_addr ipv4 );

	unsigned long long getIpv6NetworkNum( in6_addr ipv6 );

	unsigned long long getIpv6InterfaceNum( in6_addr ipv6 );

	std::string toString( StringList list );

	std::string toString( ResponseMap map );

	std::string getString( const char* message, long long value );

	std::string getFieldValue( const ResponseMap &responseMap, const std::string &fieldName );

}
#endif /* NA_DB_UTIL_H_ */
