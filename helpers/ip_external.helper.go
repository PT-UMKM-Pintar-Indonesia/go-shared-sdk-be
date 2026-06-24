package sdk_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type ipExternal struct {
	key    string
	client *HttpClient
	opt    HttpClientOptions
}

func NewIPExternal(key string, opt ...HttpClientOptions) sdk_inf.IPExternal {
	config := HttpClientOptions{
		MaxRetry:            3,
		RetryWaitMin:        1 * time.Second,
		RetryWaitMax:        3 * time.Second,
		MaxResponseBodySize: 1 * 1024 * 1024,
		Logger:              slog.Default(),
	}

	if len(opt) > 0 {
		config = opt[0]
	}

	return &ipExternal{
		key:    key,
		client: NewHttpClient(config),
		opt:    config,
	}
}

func (h *ipExternal) doRequest(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return h.client.Do(ctx, RequestOptions{
		Method: "GET",
		URL:    url,
	})
}

/*
Url: https://api.ipify.org

	Limit: Unlimited
	Apikey: No
	Example: 180.252.87.22
*/
func (h *ipExternal) One() string {
	body, err := h.doRequest("https://api.ipify.org")
	if err != nil {
		return sdk_cons.EMPTY
	}
	return string(body)
}

/*
Url: https://ipapi.co/api

	Limit: 1000 Per Day
	Apikey: No
	Example: {
		"ip": "180.252.87.22",
		"network": "180.252.80.0/21",
		"version": "IPv4",
		"city": "Depok",
		"region": "West Java",
		"region_code": "JB",
		"country": "ID",
		"country_name": "Indonesia",
		"country_code": "ID",
		"country_code_iso3": "IDN",
		"country_capital": "Jakarta",
		"country_tld": ".id",
		"continent_code": "AS",
		"in_eu": false,
		"postal": "16426",
		"latitude": -6.3792,
		"longitude": 106.8201,
		"timezone": "Asia/Jakarta",
		"utc_offset": "+0700",
		"country_calling_code": "+62",
		"currency": "IDR",
		"currency_name": "Rupiah",
		"languages": "id,en,nl,jv",
		"country_area": 1919440.0,
		"country_population": 267663435,
		"asn": "AS7713",
		"org": "PT Telekomunikasi Indonesia"
	}
*/
func (h *ipExternal) Two(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://ipapi.co/%s/json", ip))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://api.ipgeolocation.io/v3/ipgeo

	Limit: Unknown
	Apikey: Yes
	Example: {
		"ip": "180.252.87.22",
		"location": {
			"continent_code": "AS",
			"continent_name": "Asia",
			"country_code2": "ID",
			"country_code3": "IDN",
			"country_name": "Indonesia",
			"country_name_official": "Republic of Indonesia",
			"country_capital": "Jakarta",
			"state_prov": "West Java",
			"state_code": "ID-JB",
			"district": "Depok City",
			"city": "Sawangan",
			"zipcode": "16435",
			"latitude": "-6.40248",
			"longitude": "106.79424",
			"is_eu": false,
			"country_flag": "https://ipgeolocation.io/static/flags/id_64.png",
			"geoname_id": "1986811",
			"country_emoji": "🇮🇩"
		},
		"country_metadata": {
			"calling_code": "+62",
			"tld": ".id",
			"languages": [
				"id",
				"en",
				"nl",
				"jv"
			]
		},
		"currency": {
			"code": "IDR",
			"name": "Indonesian Rupiah",
			"symbol": "Rp"
		},
		"asn": {
			"as_number": "AS7713",
			"organization": "Telekomunikasi Indonesia PT",
			"country": "ID"
		},
		"time_zone": {
			"name": "Asia/Jakarta",
			"offset": 7,
			"offset_with_dst": 7,
			"current_time": "2026-03-14 15:29:49.535+0700",
			"current_time_unix": 1773476989.535,
			"current_tz_abbreviation": "WIB",
			"current_tz_full_name": "Western Indonesia Time",
			"standard_tz_abbreviation": "WIB",
			"standard_tz_full_name": "Western Indonesia Time",
			"is_dst": false,
			"dst_savings": 0,
			"dst_exists": false,
			"dst_tz_abbreviation": "",
			"dst_tz_full_name": "",
			"dst_start": {},
			"dst_end": {}
		}
	}
*/
func (h *ipExternal) Three(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://api.ipgeolocation.io/v3/ipgeo?apiKey=%s&ip=%s", h.key, ip))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://api.ipapi.is

	Limit: 1000 Per Day
	Apikey: Yes
	Example: {
		"ip": "180.252.87.22",
		"rir": "APNIC",
		"is_bogon": false,
		"is_mobile": false,
		"is_satellite": false,
		"is_crawler": false,
		"is_datacenter": false,
		"is_tor": false,
		"is_proxy": false,
		"is_vpn": false,
		"is_abuser": false,
		"company": {
			"name": "PT TELKOM INDONESIA",
			"abuser_score": "0.0049 (Low)",
			"domain": "telkom.co.id",
			"type": "isp",
			"network": "180.252.64.0 - 180.252.127.255",
			"whois": "https://api.ipapi.is/?whois=180.252.64.0"
		},
		"abuse": {
			"name": "ABUSE IDTELKOMID",
			"address": "PT. TELKOM INDONESIA, Indibiz Experience Center 3rd Floor, Kebon Sirih No 36, Jakarta",
			"email": "abuse@telkom.co.id",
			"phone": "+000000000"
		},
		"asn": {
			"asn": 7713,
			"abuser_score": "0.0012 (Low)",
			"route": "180.252.80.0/20",
			"descr": "TELKOMNET-AS-AP PT Telekomunikasi Indonesia, ID",
			"country": "id",
			"active": true,
			"org": "Telekomunikasi Indonesia (PT)",
			"domain": "telkom.co.id",
			"abuse": "abuse@telkom.co.id",
			"type": "isp",
			"updated": "2025-02-26",
			"rir": "APNIC",
			"whois": "https://api.ipapi.is/?whois=AS7713"
		},
		"location": {
			"is_eu_member": false,
			"calling_code": "62",
			"currency_code": "IDR",
			"continent": "AS",
			"country": "Indonesia",
			"country_code": "ID",
			"state": "West Java",
			"city": "Bogor",
			"latitude": -6.59444,
			"longitude": 106.789,
			"zip": "",
			"timezone": "Asia/Jakarta",
			"local_time": "2026-03-14T15:22:37+07:00",
			"local_time_unix": 1773476557,
			"is_dst": false
		},
		"elapsed_ms": 0.82
	}
*/
func (h *ipExternal) Four(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://api.ipapi.is?q=%s&key=%s", ip, h.key))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://ip-api.com

	Limit: Unlimited
	Apikey: No
	Example: {
		"status": "success",
		"country": "Indonesia",
		"countryCode": "ID",
		"region": "JB",
		"regionName": "West Java",
		"city": "Depok",
		"zip": "16431",
		"lat": -6.3966,
		"lon": 106.8164,
		"timezone": "Asia/Jakarta",
		"isp": "PT. TELKOM INDONESIA",
		"org": "",
		"as": "AS7713 PT Telekomunikasi Indonesia",
		"query": "180.252.87.22"
	}
*/
func (h *ipExternal) Five(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://ip-api.com/json/%s", ip))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://api.ipdata.co

	Limit: 1500 Per Day
	Apikey: Yes
	Example: {
		"ip": "180.252.87.22",
		"is_eu": false,
		"city": "Depok",
		"region": "Jawa Barat",
		"region_code": "JB",
		"region_type": "province",
		"country_name": "Indonesia",
		"country_code": "ID",
		"continent_name": "Asia",
		"continent_code": "AS",
		"latitude": -6.396599769592285,
		"longitude": 106.81639862060547,
		"postal": "16431",
		"calling_code": "62",
		"flag": "https://ipdata.co/flags/id.png",
		"emoji_flag": "🇮🇩",
		"emoji_unicode": "U+1F1EE U+1F1E9",
		"asn": {
			"asn": "AS7713",
			"name": "PT Telekomunikasi Indonesia",
			"domain": "telkom.co.id",
			"route": "180.252.80.0/20",
			"type": "isp"
		},
		"languages": [
			{
				"name": "Indonesian",
				"native": "Bahasa Indonesia",
				"code": "id"
			}
		],
		"currency": {
			"name": "Indonesian Rupiah",
			"code": "IDR",
			"symbol": "Rp",
			"native": "Rp",
			"plural": "Indonesian rupiahs"
		},
		"time_zone": {
			"name": "Asia/Jakarta",
			"abbr": "WIB",
			"offset": "+0700",
			"is_dst": false,
			"current_time": "2026-03-14T08:19:38+07:00"
		},
		"threat": {
			"is_tor": false,
			"is_icloud_relay": false,
			"is_proxy": false,
			"is_datacenter": false,
			"is_anonymous": false,
			"is_known_attacker": false,
			"is_known_abuser": false,
			"is_threat": false,
			"is_bogon": false,
			"blocklists": []
		},
		"count": "1"
	}
*/
func (h *ipExternal) Six(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://api.ipdata.co/%s?api-key=%s", ip, h.key))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://api.hackertarget.com

	Limit: Unlimited
	Apikey: No
	Example: {
		"city": "Depok",
		"country": "Indonesia",
		"ip": "180.252.87.22",
		"latitude": -6.3792,
		"longitude": 106.8201,
		"state": "West Java"
	}
*/
func (h *ipExternal) Seven(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://api.hackertarget.com/geoip/?output=json&q=%s", ip))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://api.country.is

	Limit: Unlimited
	Apikey: No
	Example: {
		"ip": "180.252.87.22",
		"country": "ID",
		"city": "Depok",
		"continent": "AS",
		"subdivision": "JB",
		"postal": "16426",
		"location": {
			"latitude": -6.3792,
			"longitude": 106.8201,
			"accuracy_radius": 20,
			"time_zone": "Asia/Jakarta"
		},
		"asn": {
			"number": 7713,
			"organization": "PT Telekomunikasi Indonesia"
		}
	}
*/
func (h *ipExternal) Eight(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://api.country.is/%s?fields=city,continent,subdivision,postal,location,asn", ip))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https://ip-intelligence.abstractapi.com/v1

	Limit: 1000 Per Day
	Apikey: Yes
	Example:{
		"ip_address": "180.252.87.22",
		"security": {
			"is_vpn": false,
			"is_proxy": false,
			"is_tor": false,
			"is_hosting": false,
			"is_relay": false,
			"is_mobile": false,
			"is_abuse": false
		},
		"asn": {
			"asn": 7713,
			"name": "PT Telekomunikasi Indonesia",
			"domain": null,
			"type": "isp"
		},
		"company": {
			"name": "PT Telekomunikasi Indonesia",
			"domain": null,
			"type": "isp"
		},
		"domains": {
			"domains": []
		},
		"location": {
			"city": "Depok",
			"city_geoname_id": 1645524,
			"region": "West Java",
			"region_iso_code": "JB",
			"region_geoname_id": 1642672,
			"postal_code": "16426",
			"country": "Indonesia",
			"country_code": "ID",
			"country_geoname_id": 1643084,
			"is_country_eu": false,
			"continent": "Asia",
			"continent_code": "AS",
			"continent_geoname_id": 6255147,
			"longitude": 106.8201,
			"latitude": -6.3792
		},
		"timezone": {
			"name": "Asia/Jakarta",
			"abbreviation": "WIB",
			"utc_offset": 7,
			"local_time": "15:09:11",
			"is_dst": false
		},
		"flag": {
			"emoji": "🇮🇩",
			"unicode": "U+1F1EE U+1F1E9",
			"png": "https://static.abstractapi.com/country-flags/ID_flag.png",
			"svg": "https://static.abstractapi.com/country-flags/ID_flag.svg"
		},
		"currency": {
			"name": "Indonesian Rupiah",
			"code": "IDR",
			"symbol": ""
		}
	}
*/
func (h *ipExternal) Nine(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://ip-intelligence.abstractapi.com/v1/?api_key=%s&ip_address=%s", h.key, ip))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}

/*
Url: https:///api.geoapify.com/v1/ipinfo

		Limit: 3000 Per Day
		Apikey: Yes
		Example:{
		"city": {
			"name": "Bogor",
			"names": {
				"en": "Bogor",
				"de": "Bogor",
				"fa": "بوگور",
				"fr": "Bogor",
				"ja": "ボゴール",
				"ko": "보고르",
				"pt-BR": "Bogor",
				"ru": "Богор",
				"zh-CN": "茂物"
			}
		},
		"country": {
			"name": "Indonesia",
			"iso_code": "ID",
			"names": {
				"en": "Indonesia",
				"de": "Indonesien",
				"es": "Indonesia",
				"fa": "اندونزی",
				"fr": "Indonésie",
				"ja": "インドネシア",
				"ko": "인도네시아",
				"pt-BR": "Indonésia",
				"ru": "Индонезия",
				"zh-CN": "印尼"
			},
			"geoname_id": 1643084,
			"name_native": "Indonesia",
			"phone_code": "62",
			"capital": "Jakarta",
			"currency": "IDR",
			"flag": "🇮🇩",
			"languages": [
				{
					"iso_code": "id",
					"name": "Indonesian",
					"name_native": "Bahasa Indonesia"
				}
			]
		},
		"state": {
			"name": "West Java"
		},
		"location": {
			"latitude": -6.59444,
			"longitude": 106.789
		},
		"continent": {
			"code": "AS",
			"name": "Asia",
			"names": {
				"de": "Asien",
				"en": "Asia",
				"es": "Asia",
				"fa": " آسیا",
				"fr": "Asie",
				"ja": "アジア大陸",
				"ko": "아시아",
				"pt-BR": "Ásia",
				"ru": "Азия",
				"zh-CN": "亚洲"
			},
			"geoname_id": 6255147
		},
		"subdivisions": [
			{
				"names": {
					"en": "West Java",
					"de": "Jawa Barat",
					"es": "Java Occidental",
					"fa": "جاوه غربی",
					"fr": "Java occidental",
					"ja": "西ジャワ州",
					"ko": "자와바랏 주",
					"pt-BR": "Java Ocidental",
					"ru": "Западная Ява",
					"zh-CN": "西爪哇省"
				}
			},
			{
				"names": {
					"en": "Kota Bogor"
				}
			}
		],
		"ip": "180.252.87.22"
	}
*/
func (h *ipExternal) Teen(ip string) map[string]any {
	body, err := h.doRequest(fmt.Sprintf("https://api.geoapify.com/v1/ipinfo?ip=%s&apiKey=%s", ip, h.key))
	if err != nil {
		return nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	return result
}
