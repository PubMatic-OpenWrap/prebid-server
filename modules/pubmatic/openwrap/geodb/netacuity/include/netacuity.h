#ifndef NETACUITY_H_
#define NETACUITY_H_

/**
 * @file netacuity.h
 * @brief NetAcuity embedded API for C.
 */

#include <cstdlib>

/**
 * @brief A NetAcuity database handle.
 */
typedef const void *NaDbHandle;

/**
 * @brief Returns a pointer to a char buffer containing the most recent error message.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls on the same thread.
 *
 * @return a pointer to a char buffer
 */
extern "C" char *nadb_error();

/**
 * @brief Opens a NetAcuity database and returns a database handle.
 *
 * Every database handle, whether created directly by nadb_open() or cloned by nadb_clone(), must be freed by
 * nadb_free() when it is no longer needed.
 *
 * @param directory the path to the directory that contains the database files
 * @param feature_code the database feature code
 * @return a new database handle or NULL on error
 */
extern "C" NaDbHandle nadb_open(const char *directory, u_int32_t feature_code);

/**
 * @brief Clones a NetAcuity database handle for use in a multithreaded environment.
 *
 * Every database handle, whether created directly by nadb_open() or cloned by nadb_clone(), must be freed by
 * nadb_free() when it is no longer needed.
 *
 * @param handle the database handle to clone
 * @return a cloned database handle or NULL on error
 */
extern "C" NaDbHandle nadb_clone(NaDbHandle handle);

/**
 * @brief Frees the resources held by a NetAcuity database and renders the database handle unusable.
 *
 * @param handle the database handle
 */
extern "C" void nadb_free(NaDbHandle handle);

/**
 * @brief Returns the database format version.
 *
 * @param handle the database handle
 * @return a non-negative database format version or -1 on error
 */
extern "C" int32_t nadb_format_version(NaDbHandle handle);

/**
 * @brief Returns the database version number.
 *
 * @param handle the database handle
 * @return a non-negative database version or -1 on error
 */
extern "C" int32_t nadb_db_version(NaDbHandle handle);

/**
 * @brief Returns a pointer to a char buffer containing the database build date string.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_db_build_date(NaDbHandle handle);

/**
 * @brief Returns a pointer to a char buffer containing the default response.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_default_response(NaDbHandle handle);

/**
 * @brief Returns a pointer to a char buffer containing the comma-separated list of field names.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_field_names(NaDbHandle handle);

/**
 * @brief Finds a schema entry by name and returns a pointer to a char buffer containing its value.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @param param_name the parameter name
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_schema_value(NaDbHandle handle, const char *param_name);

/**
 * @brief Returns a pointer to a char buffer containing the NetAcuity response associated with an IP address.
 *
 * This function accepts both IPv4 and IPv6 addresses in standard dot notation (e.g. "1.2.3.4", "1234:abcd::").
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @param ip the IP address string
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_query(NaDbHandle handle, const char *ip);

/**
 * @brief Returns a pointer to a char buffer containing the NetAcuity response associated with an IPv4 address.
 *
 * This function accepts an IPv4 address as a 32-bit unsigned integer.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @param ip the IPv4 address
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_query_ipv4(NaDbHandle handle, u_int32_t ip);

/**
 * @brief Returns a pointer to a char buffer containing the NetAcuity response associated with an IPv6 address.
 *
 * This function accepts an IPv6 address as a 64-bit unsigned integer network address and a 64-bit unsigned integer
 * interface address.
 *
 * The returned buffer is managed internally; do not attempt to free it. If the buffer contents are to be retained, they
 * must be copied out of the buffer before any subsequent API calls using the same database handle.
 *
 * @param handle the database handle
 * @param network the IPv6 network address
 * @param interface the IPv6 interface address
 * @return <ul><li>A pointer to a char buffer OR
 *        </li><li>NULL if an error occurs
 *        </li></ul>
 */
extern "C" char *nadb_query_ipv6(NaDbHandle handle, u_int64_t network, u_int64_t interface);

#endif // NETACUITY_H_
