#ifndef NA_DB_WRAPPER_H
#define NA_DB_WRAPPER_H

#define THREE_COUNTRY_LEN	3
#define TWO_COUNTRY_LEN	2
#define ONE_CHAR_LEN		1
#define CONN_SPEED_LEN		100
#define POSTAL_CODE_LEN	50
#define GENERAL_STRING_LEN	255
#define IP_STRING_LEN		49
#define AREA_CODE_LEN		255
#define MCC_MNC_LEN		9
#ifdef __cplusplus
	extern "C" {
#endif
	/* Feature - 4 - Edge response structure */
	typedef struct na_edge_data {
		char  edge_country[THREE_COUNTRY_LEN +1];
		char  edge_region[GENERAL_STRING_LEN +1];
		char  edge_city[GENERAL_STRING_LEN +1];
		char  edge_conn_speed[CONN_SPEED_LEN +1];
		int   edge_metro_code;
		float edge_latitude;
		float edge_longitude;
		char  edge_postal_code[POSTAL_CODE_LEN +1];
		int   edge_country_code;
		int   edge_region_code;
		int   edge_city_code;
		int   edge_continent_code;
		char  edge_two_letter_country[TWO_COUNTRY_LEN +1];
		int   edge_internal_code;
		char  edge_area_codes[AREA_CODE_LEN +1];
		int   edge_country_conf;
		int   edge_region_conf;
		int   edge_city_conf;
		int   edge_postal_conf;
		int   edge_gmt_offset;
		char  edge_in_dst[ONE_CHAR_LEN +1];
	}na_edge_data_t;

	/* Feature - 24 - Mobile Carrier response structure*/
	typedef struct na_mobile_carrier {
		char mobile_carrier[GENERAL_STRING_LEN +1];
		char mcc[MCC_MNC_LEN +1];
		char mnc[MCC_MNC_LEN +1];
	}na_mobile_carrier_data_t;

	typedef struct {
		void *parser;
	}NetAcuityClient;

	/* Return NaDbParser object pointer*/
	extern NetAcuityClient* init_netacuity_client(const char* na_db_data_dir, int feature_code, int use_memory_mapped_file);
	/* Free NaDbParser object*/
	extern void free_netacuity_client(NetAcuityClient* na_client, int feature_code);
	/* Initialize na_edge_data_t object*/
	extern void init_na_edge_data(na_edge_data_t* edge_data);
	/* Initialize na_mobile_carrier_t object*/
	extern void init_na_mobile_carrier_data(na_mobile_carrier_data_t* edge_data);
	/* Fetch Geo details for given ip_addr and save these details into na_edge_data_t object*/
	extern int get_geo_detail_from_edge_db(NetAcuityClient* na_client, const char* ip_addr, na_edge_data_t* edge_data);
	/* Fetch Mobile Carrier details for given ip_addr and save these details into na_mobile_carrier_t object*/
	extern int get_mobile_carrier_detail(NetAcuityClient* na_client, const char* ip_addr, na_mobile_carrier_data_t* mc_data);
#ifdef __cplusplus
	}
#endif
#endif
