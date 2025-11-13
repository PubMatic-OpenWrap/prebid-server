//============================================================================
// Program     : C++ NetAcuity Embedded API
// Author      : Digital Envoy
// Version     : 7.0.0.1
// Date        : 2023-DEC-08
// Copyright   : Copyright 2000-2023, Digital Envoy, Inc.  All rights reserved.
//============================================================================

#ifndef NA_DB_ACCESSOR_H_
#define NA_DB_ACCESSOR_H_

#include <stdexcept>
#include "netacuity.h"
#include "NaDbDef.h"
#include "NaDbUtil.h"

namespace netacuity {

    class NaDbAccessor {

    public:
        NaDbAccessor(int featureCode, std::string directory);

        virtual ~NaDbAccessor();

        //--------------- Public MetaData Getters ----------------
        int getFeatureCode() const;

        std::string getDirectory() const;

        int getDbVersion() const;

        std::string getBuildDateString() const;

        StringList getFieldNameList() const;

        StringList getDefaultResponseList() const;

        std::string getDefaultRawResponse() const;


        //--------------- Public Raw Query Functions ----------------

        /// Retrieve the raw response for the specified IPv4 number.
        std::string queryRawResponseIpv4(u_int32_t ipNum);

        /// Retrieve the raw response for the specified IPv6 network number.
        /// The network number is the half of the IPv6 address which contains the most significant-bits.
        std::string queryRawResponseIpv6(u_int64_t numNetwork);

        /// Retrieve the raw response for the specified IPv6 full network and interface numbers.
        /// The network number is the half of the IPv6 address which contains the most significant bits.
        /// The interface number is the half of the IPv6 address which contains the least significant bits.
        std::string queryRawResponseIpv6(u_int64_t numNetwork, u_int64_t numInterface);

        /// Retrieve the raw response for the specified IPv4 address.
        std::string queryRawResponseIpv4(in_addr ipv4);

        /// Retrieve the raw response for the specified IPv6 address.
        std::string queryRawResponseIpv6(in6_addr ipv6);

        /// Retrieve the raw response for the specified IP "dotted" presentation address, IPv4 or IPv6.
        std::string queryRawResponse(const char *ipAddress);

        /// Prints some interesting meta-data to the specified output stream.  Useful for testing or debugging.
        void printMetaData(std::ostream &out) const;

    protected:
        int featureCode;
        std::string directory;
        NaDbHandle handle;

        int dbVersion;
        std::string buildDateString;
        std::string defaultResponse;
        StringList fieldNameList;
        StringList defaultResponseList;

        void setupMetadata();

        std::runtime_error getNaDbError();
    };

}
#endif /* NA_DB_ACCESSOR_H_ */
