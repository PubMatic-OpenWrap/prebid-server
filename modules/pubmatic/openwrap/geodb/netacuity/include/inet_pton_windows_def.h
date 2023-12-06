#ifndef INET_PTON_WINDOWS_DEF_H_
#define INET_PTON_WINDOWS_DEF_H_

#ifdef _WIN32
int inet_pton(int af, const char *src, void *dst);
#endif /* _WIN32 */

#endif /* INET_PTON_WINDOWS_DEF_H_*/
