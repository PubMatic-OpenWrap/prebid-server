//============================================================================
// Program     : NetAcuity C++ Embedded API
// Author      : Digital Envoy
// Version     : 6.4.1.3
// Date        : 25-Jun-2019
// Copyright   : Copyright 2000-2019, Digital Envoy, Inc.  All rights reserved.
//============================================================================


#ifndef NA_DB_BASE_ACCESSOR_H_
#define NA_DB_BASE_ACCESSOR_H_

#include <stdexcept>
#include "NaDbDef.h"
#include "NaDbUtil.h"

namespace netacuity {

class NaDbBaseAccessor {

public:
	NaDbBaseAccessor( int featureCode, std::string directory );
	virtual ~NaDbBaseAccessor();

	//--------------- Public MetaData Getters ----------------
	int getFeatureCode() const;
	std::string getDirectory() const;
	std::string getResponseFilePath() const;
	std::string getRangeIpv4FilePath() const;
	std::string getSchemaFilePath() const;
	std::string getRangeIpv6FilePath() const;
	std::string getExtendedRangeIpv4FilePath() const;
	std::string getExtendedRangeIpv6FilePath() const;

	RangeRecordIpv6Type getRangeRecordIpv6Type() const;
	bool supportsIpv6() const;
	unsigned long getRangeCountIpv4() const;
	unsigned long getRangeCountIpv6() const;
	unsigned long getExtendedRangeCountIpv4() const;
	unsigned long getExtendedRangeCountIpv6() const;
	int getDbVersion() const;
	std::string getBuildDateString() const;
	StringList getFieldNameList() const;
	StringList getDefaultResponseList() const;
	std::string getDefaultRawResponse() const;


	//--------------- Public Raw Query Functions ----------------

	/// Retrieve the raw response for the specified IPv4 number.
	std::string queryRawResponseIpv4( unsigned long ipNum ) const;

	/// Retrieve the raw response for the specified IPv6 network number.
	/// The network number is the half of the IPv6 address which contains the most significant-bits.
	std::string queryRawResponseIpv6( unsigned long long numNetwork ) const;

	/// Retrieve the raw response for the specified IPv6 full network and interface numbers.
	/// The network number is the half of the IPv6 address which contains the most significant bits.
	/// The interface number is the half of the IPv6 address which contains the least significant bits.
	std::string queryRawResponseIpv6( unsigned long long numNetwork, unsigned long long numInterface ) const;

	/// Retrieve the raw response for the specified IPv4 address.
	std::string queryRawResponseIpv4( in_addr ipv4 ) const;

	/// Retrieve the raw response for the specified IPv6 address.
	std::string queryRawResponseIpv6( in6_addr ipv6 ) const;

	/// Retrieve the raw response for the specified IP "dotted" presentation address, IPv4 or IPv6.
	std::string queryRawResponse( const char* ipAddress ) const;


	/// Prints some interesting meta-data to the specified output stream.  Useful for testing or debugging.
	void printMetaData( std::ostream &out );

protected:
	int featureCode;
	std::string directory;

	std::string responseFilename;
	std::string rangeIpv4Filename;
	std::string schemaFilename;
	std::string rangeIpv6Filename;
	std::string extendedRangeIpv4Filename;
	std::string extendedRangeIpv6Filename;

	std::string responseFilepath;
	std::string rangeIpv4Filepath;
	std::string schemaFilepath;
	std::string rangeIpv6Filepath;
	std::string extendedRangeIpv4Filepath;
	std::string extendedRangeIpv6Filepath;

	RangeRecordIpv6Type rangeRecordIpv6Type;

	size_t responseFileSize;
	size_t rangeIpv4FileSize;
	size_t rangeIpv6FileSize;
	size_t extendedRangeIpv4FileSize;
	size_t extendedRangeIpv6FileSize;

	unsigned long recordCountIpv4;
	unsigned long recordCountIpv6;
	unsigned long extendedRecordCountIpv4;
	unsigned long extendedRecordCountIpv6;

	unsigned long ipNumStandardCutoffIpv4;  // the inclusive cutoff ipNum for accessing the standard IPv4 range-file, beyond which the extended IPv4 range-file should be accessed
	unsigned long long numNetworkStandardCutoffIpv6;   // the inclusive cutoff  networkNum  for accessing the standard IPv6 range-file, beyond which the extended IPv6 range-file should be accessed
	unsigned long long numInterfaceStandardCutoffIpv6; // the inclusive cutoff interfaceNum for accessing the standard IPv6 range-file, beyond which the extended IPv6 range-file should be accessed

	int dbVersion;
	std::string buildDateString;
	std::string defaultResponse;
	StringList fieldNameList;
	StringList defaultResponseList;


	//--------------- Protected Query Helper Functions ----------------

	virtual std::string getRawResponse( unsigned long offset) const = 0;
	virtual RangeRecordIpv4 getRangeRecordIpv4( unsigned int index ) const = 0;
	virtual RangeRecordIpv6 getRangeRecordIpv6( unsigned int index ) const = 0;
	virtual RangeRecordIpv6NetworkOnly getRangeRecordIpv6NetworkOnly( unsigned int index ) const = 0;
	virtual ExtendedRangeRecordIpv4 getExtendedRangeRecordIpv4( unsigned int index ) const = 0;
	virtual ExtendedRangeRecordIpv6 getExtendedRangeRecordIpv6( unsigned int index ) const = 0;
	virtual ExtendedRangeRecordIpv6NetworkOnly getExtendedRangeRecordIpv6NetworkOnly( unsigned int index ) const = 0;

	unsigned long getOffset( unsigned long ip ) const;
	unsigned long getOffset( unsigned long long numNetwork, unsigned long long numInterface ) const;
	unsigned long getOffset( unsigned long long numNetwork ) const;
	unsigned long getStandardOffset( unsigned long ip ) const;
	unsigned long getStandardOffset( unsigned long long numNetwork, unsigned long long numInterface ) const;
	unsigned long getStandardOffset( unsigned long long numNetwork ) const;
	unsigned long getExtendedOffset( unsigned long ip ) const;
	unsigned long getExtendedOffset( unsigned long long numNetwork, unsigned long long numInterface ) const;
	unsigned long getExtendedOffset( unsigned long long numNetwork ) const;

	void parseMetaData( std::string defaultResponse, std::string metaData );
	void prepareStandardCutoffs();


private:
	static RangeRecordIpv6Type processSchemaFile( std::string schemaFilepath );
	void setFilePaths();
	void setFileSizes();
};

}
#endif /* NA_DB_BASE_ACCESSOR_H_ */
