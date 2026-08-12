//============================================================================
// Program     : NetAcuity C++ Embedded API
// Author      : Digital Envoy
// Version     : 6.4.1.3
// Date        : 25-Jun-2019
// Copyright   : Copyright 2000-2019, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_DEF_H_
#define NA_DB_DEF_H_

#include <string>
#include <list>
#include <map>
#include <stdlib.h>
#include <stdint.h>

namespace netacuity {

	typedef std::list<std::string> StringList;
	typedef std::map<std::string, std::string> ResponseMap;

	static const unsigned      long MAX_IPV4_NUM = 0xFFffFFffUL;
	static const unsigned long long MAX_IPV6_NETWORK_NUM = 0xFFFFffffFFFFffffULL;
	static const unsigned long long MAX_IPV6_INTERFACE_NUM = MAX_IPV6_NETWORK_NUM;

	typedef enum RangeRecordIpv6Type_enum {
		RANGE_RECORD_IPV6_NONE = 0,
		RANGE_RECORD_IPV6_NETWORK_ONLY = 20,
		RANGE_RECORD_IPV6_NETWORK_AND_INTERFACE = 36
	}
	RangeRecordIpv6Type;


	typedef struct RangeRecordIpv4_struct {
		RangeRecordIpv4_struct() : startNum(0), endNum(0), offset(0) {}
		//all these byte-orders are little-endian
		uint32_t startNum;
		uint32_t endNum;
		uint32_t offset;
	}
	__attribute__((__packed__))
	RangeRecordIpv4;


	typedef struct RangeRecordIpv6_struct {
		RangeRecordIpv6_struct() : startNetwork(0), startInterface(0), endNetwork(0), endInterface(0), offset(0) {}
		//all these byte-orders are little-endian
		uint64_t startNetwork;
		uint64_t startInterface;
		uint64_t endNetwork;
		uint64_t endInterface;
		uint32_t offset;
	}
	__attribute__((__packed__))
	RangeRecordIpv6;


	typedef struct RangeRecordIpv6NetworkOnly_struct {
		RangeRecordIpv6NetworkOnly_struct() : startNetwork(0), endNetwork(0), offset(0) {}
		//all these byte-orders are little-endian
		uint64_t startNetwork;
		uint64_t endNetwork;
		uint32_t offset;
	}
	__attribute__((__packed__))
	RangeRecordIpv6NetworkOnly;


	typedef struct ExtendedRangeRecordIpv4_struct {
		ExtendedRangeRecordIpv4_struct() : startNum(0), endNum(0), offset(0), offsetExtended(0) {}
		//all these byte-orders are little-endian
		uint32_t startNum;
		uint32_t endNum;
		uint32_t offset;
		uint8_t  offsetExtended;  //most-significant byte of offset occurs last in little-endian order
	}
	__attribute__((__packed__))
	ExtendedRangeRecordIpv4;


	typedef struct ExtendedRangeRecordIpv6_struct {
		ExtendedRangeRecordIpv6_struct() : startNetwork(0), startInterface(0), endNetwork(0), endInterface(0), offset(0), offsetExtended(0) {}
		//all these byte-orders are little-endian
		uint64_t startNetwork;
		uint64_t startInterface;
		uint64_t endNetwork;
		uint64_t endInterface;
		uint32_t offset;
		uint8_t  offsetExtended;  //most-significant byte of offset occurs last in little-endian order
	}
	__attribute__((__packed__))
	ExtendedRangeRecordIpv6;


	typedef struct ExtendedRangeRecordIpv6NetworkOnly_struct {
		ExtendedRangeRecordIpv6NetworkOnly_struct() : startNetwork(0), endNetwork(0), offset(0), offsetExtended(0) {}
		//all these byte-orders are little-endian
		uint64_t startNetwork;
		uint64_t endNetwork;
		uint32_t offset;
		uint8_t  offsetExtended;  //most-significant byte of offset occurs last in little-endian order
	}
	__attribute__((__packed__))
	ExtendedRangeRecordIpv6NetworkOnly;

}
#endif /* NA_DB_DEF_H_ */
