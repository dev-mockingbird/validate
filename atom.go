package validate

var (
	atoms map[string]func(any) bool
)

func init() {
	atoms = map[string]func(any) bool{
		"phone":               IsNumeric,
		"username":            IsUsername,
		"password":            IsPassword,
		"strongPassword":      IsStrongPassword,
		"countryCodeAlpha2":   IsCountryCodeAlpha2,
		"countryCodeAlpha3":   IsCountryCodeAlpha3,
		"countryCodeNumeric":  IsCountryCodeAlphaNumeric,
		"countryCode":         IsCountryCode,
		"currency":            IsCurrencyCode,
		"currencyNumeric":     IsCurrencyCodeNumeric,
		"alpha":               IsAlpha,
		"alphaNumeric":        IsAlphaNumeric,
		"alphaUnicode":        IsAlphaUnicode,
		"alphaUnicodeNumeric": IsAlphaUnicodeNumeric,
		"numeric":             IsNumeric,
		"number":              IsNumber,
		"hexadecimal":         IsHexadecimal,
		"hexColor":            IsHexColor,
		"rgb":                 IsRgb,
		"rgba":                IsRgba,
		"hsl":                 IsHsl,
		"hsla":                IsHsla,
		"e164":                IsE164,
		"email":               IsEmail,
		"base64":              IsBase64,
		"base64URL":           IsBase64URL,
		"base64RawURL":        IsBase64RawURL,
		"iSBN10":              IsISBN10,
		"iSBN13":              IsISBN13,
		"uUID3":               IsUUID3,
		"uUID4":               IsUUID4,
		"uUID5":               IsUUID5,
		"uUID":                IsUUID,
		"uUID3RFC4122":        IsUUID3RFC4122,
		"uUID4RFC4122":        IsUUID4RFC4122,
		"uUID5RFC4122":        IsUUID5RFC4122,
		"uUIDRFC4122":         IsUUIDRFC4122,
		"uLID":                IsULID,
		"md4":                 IsMd4,
		"md5":                 IsMd5,
		"sha256":              IsSha256,
		"sha384":              IsSha384,
		"sha512":              IsSha512,
		"ripemd128":           IsRipemd128,
		"ripemd160":           IsRipemd160,
		"tiger128":            IsTiger128,
		"tiger160":            IsTiger160,
		"tiger192":            IsTiger192,
		"aSCII":               IsASCII,
		"printableASCII":      IsPrintableASCII,
		"multibyte":           IsMultibyte,
		"dataURI":             IsDataURI,
		"latitude":            IsLatitude,
		"longitude":           IsLongitude,
		"sSN":                 IsSSN,
		"hostnameRFC952":      IsHostnameRFC952,
		"hostnameRFC1123":     IsHostnameRFC1123,
		"fqdn":                IsFqdn,
		"btcAddress":          IsBtcAddress,
		"btcUpperAddress":     IsBtcUpperAddress,
		"btcLowerAddress":     IsBtcLowerAddress,
		"ethAddress":          IsEthAddress,
		"uRLEncoded":          IsURLEncoded,
		"hTMLEncoded":         IsHTMLEncoded,
		"hTML":                IsHTML,
		"jWT":                 IsJWT,
		"splitParams":         IsSplitParams,
		"bic":                 IsBic,
		"semver":              IsSemver,
		"dns":                 IsDns,
		"cve":                 IsCve,
		"mongodb":             IsMongodb,
		"cron":                IsCron,
	}
}

func Register(name string, callback func(any) bool) {
	atoms[name] = callback
}

func IsPassword(val any) bool {
	if !IsASCII(val) {
		return false
	}
	password := val.(string)
	if len(password) < 8 {
		return false
	}
	hasNum, hasUpper, hasLower, _ := extractPassword(password)

	return hasNum && (hasUpper || hasLower)
}

func IsStrongPassword(val any) bool {
	if !IsASCII(val) {
		return false
	}
	password := val.(string)
	if len(password) < 8 {
		return false
	}
	hasNum, hasUpper, hasLower, hasSpecial := extractPassword(password)

	return hasNum && hasUpper && hasLower && hasSpecial
}

func extractPassword(password string) (hasNum, hasUpper, hasLower, hasSpecial bool) {
	for i := 0; i < len(password); i++ {
		if 'a' <= password[i] && password[i] <= 'z' {
			hasLower = true
		} else if 'A' <= password[i] && password[i] <= 'Z' {
			hasUpper = true
		} else if '0' <= password[i] && password[i] <= '9' {
			hasNum = true
		} else {
			hasSpecial = true
		}
	}
	return
}
