package validate

var iso4217 = map[string]bool{
	"AFN": true, "EUR": true, "ALL": true, "DZD": true, "USD": true,
	"AOA": true, "XCD": true, "ARS": true, "AMD": true, "AWG": true,
	"AUD": true, "AZN": true, "BSD": true, "BHD": true, "BDT": true,
	"BBD": true, "BYN": true, "BZD": true, "XOF": true, "BMD": true,
	"INR": true, "BTN": true, "BOB": true, "BOV": true, "BAM": true,
	"BWP": true, "NOK": true, "BRL": true, "BND": true, "BGN": true,
	"BIF": true, "CVE": true, "KHR": true, "XAF": true, "CAD": true,
	"KYD": true, "CLP": true, "CLF": true, "CNY": true, "COP": true,
	"COU": true, "KMF": true, "CDF": true, "NZD": true, "CRC": true,
	"HRK": true, "CUP": true, "CUC": true, "ANG": true, "CZK": true,
	"DKK": true, "DJF": true, "DOP": true, "EGP": true, "SVC": true,
	"ERN": true, "SZL": true, "ETB": true, "FKP": true, "FJD": true,
	"XPF": true, "GMD": true, "GEL": true, "GHS": true, "GIP": true,
	"GTQ": true, "GBP": true, "GNF": true, "GYD": true, "HTG": true,
	"HNL": true, "HKD": true, "HUF": true, "ISK": true, "IDR": true,
	"XDR": true, "IRR": true, "IQD": true, "ILS": true, "JMD": true,
	"JPY": true, "JOD": true, "KZT": true, "KES": true, "KPW": true,
	"KRW": true, "KWD": true, "KGS": true, "LAK": true, "LBP": true,
	"LSL": true, "ZAR": true, "LRD": true, "LYD": true, "CHF": true,
	"MOP": true, "MKD": true, "MGA": true, "MWK": true, "MYR": true,
	"MVR": true, "MRU": true, "MUR": true, "XUA": true, "MXN": true,
	"MXV": true, "MDL": true, "MNT": true, "MAD": true, "MZN": true,
	"MMK": true, "NAD": true, "NPR": true, "NIO": true, "NGN": true,
	"OMR": true, "PKR": true, "PAB": true, "PGK": true, "PYG": true,
	"PEN": true, "PHP": true, "PLN": true, "QAR": true, "RON": true,
	"RUB": true, "RWF": true, "SHP": true, "WST": true, "STN": true,
	"SAR": true, "RSD": true, "SCR": true, "SLL": true, "SGD": true,
	"XSU": true, "SBD": true, "SOS": true, "SSP": true, "LKR": true,
	"SDG": true, "SRD": true, "SEK": true, "CHE": true, "CHW": true,
	"SYP": true, "TWD": true, "TJS": true, "TZS": true, "THB": true,
	"TOP": true, "TTD": true, "TND": true, "TRY": true, "TMT": true,
	"UGX": true, "UAH": true, "AED": true, "USN": true, "UYU": true,
	"UYI": true, "UYW": true, "UZS": true, "VUV": true, "VES": true,
	"VND": true, "YER": true, "ZMW": true, "ZWL": true, "XBA": true,
	"XBB": true, "XBC": true, "XBD": true, "XTS": true, "XXX": true,
	"XAU": true, "XPD": true, "XPT": true, "XAG": true,
}

var iso4217_numeric = map[int]bool{
	8: true, 12: true, 32: true, 36: true, 44: true,
	48: true, 50: true, 51: true, 52: true, 60: true,
	64: true, 68: true, 72: true, 84: true, 90: true,
	96: true, 104: true, 108: true, 116: true, 124: true,
	132: true, 136: true, 144: true, 152: true, 156: true,
	170: true, 174: true, 188: true, 191: true, 192: true,
	203: true, 208: true, 214: true, 222: true, 230: true,
	232: true, 238: true, 242: true, 262: true, 270: true,
	292: true, 320: true, 324: true, 328: true, 332: true,
	340: true, 344: true, 348: true, 352: true, 356: true,
	360: true, 364: true, 368: true, 376: true, 388: true,
	392: true, 398: true, 400: true, 404: true, 408: true,
	410: true, 414: true, 417: true, 418: true, 422: true,
	426: true, 430: true, 434: true, 446: true, 454: true,
	458: true, 462: true, 480: true, 484: true, 496: true,
	498: true, 504: true, 512: true, 516: true, 524: true,
	532: true, 533: true, 548: true, 554: true, 558: true,
	566: true, 578: true, 586: true, 590: true, 598: true,
	600: true, 604: true, 608: true, 634: true, 643: true,
	646: true, 654: true, 682: true, 690: true, 694: true,
	702: true, 704: true, 706: true, 710: true, 728: true,
	748: true, 752: true, 756: true, 760: true, 764: true,
	776: true, 780: true, 784: true, 788: true, 800: true,
	807: true, 818: true, 826: true, 834: true, 840: true,
	858: true, 860: true, 882: true, 886: true, 901: true,
	927: true, 928: true, 929: true, 930: true, 931: true,
	932: true, 933: true, 934: true, 936: true, 938: true,
	940: true, 941: true, 943: true, 944: true, 946: true,
	947: true, 948: true, 949: true, 950: true, 951: true,
	952: true, 953: true, 955: true, 956: true, 957: true,
	958: true, 959: true, 960: true, 961: true, 962: true,
	963: true, 964: true, 965: true, 967: true, 968: true,
	969: true, 970: true, 971: true, 972: true, 973: true,
	975: true, 976: true, 977: true, 978: true, 979: true,
	980: true, 981: true, 984: true, 985: true, 986: true,
	990: true, 994: true, 997: true, 999: true,
}

func IsCurrencyCode(val any) bool {
	code, ok := val.(string)
	if !ok {
		return false
	}
	return iso4217[code]
}

func IsCurrencyCodeNumeric(val any) bool {
	code, ok := val.(int)
	if !ok {
		return false
	}
	return iso4217_numeric[code]
}
