package models

import (
	"fmt"
	"sort"
	"strings"
)

type Currency string

var currencies = []string{ // Must be sorted in ascending order!
	"AED",    // د.إ - United Arab Emirates dirham
	"AFN",    // ؋ - Afghan afghani
	"ALL",    // L - Albanian lek
	"AMD",    //  - Armenian dram
	"ANG",    // ƒ - Netherlands Antillean guilder
	"AOA",    // Kz - Angolan kwanza
	"ARS",    // $ - Argentine peso
	"AUD",    // $ - Australian dollar
	"AWG",    // ƒ - Aruban florin
	"AZN",    //  - Azerbaijani manat
	"BAM",    // KM or КМ[I] - Bosnia and Herzegovina convertible mark
	"BBD",    // $ - Barbadian dollar
	"BDT",    // ৳ - Bangladeshi taka
	"BGN",    // лв - Bulgarian lev
	"BHD",    // .د.ب - Bahraini dinar
	"BIF",    // Fr - Burundian franc
	"BMD",    // $ - Bermudian dollar
	"BND",    // $ - Brunei dollar
	"BOB",    // Bs. - Bolivian boliviano
	"BRL",    // R$ - Brazilian real
	"BSD",    // $ - Bahamian dollar
	"BTN",    // Nu. - Bhutanese ngultrum
	"BWP",    // P - Botswana pula
	"BYN",    // Br - New Belarusian ruble
	"BYR",    // Br - Old Belarusian ruble[H]
	"BZD",    // $ - Belize dollar
	"CAD",    // $ - Canadian dollar
	"CDF",    // Fr - Congolese franc
	"CHF",    // Fr - Swiss franc
	"CLP",    // $ - Chilean peso
	"CNY",    // ¥ or 元 - Chinese yuan
	"COP",    // $ - Colombian peso
	"CRC",    // ₡ - Costa Rican colón
	"CUC",    // $ - Cuban convertible peso
	"CUP",    // $ - Cuban peso
	"CVE",    // Esc or $ - Cape Verdean escudo
	"CZK",    // Kč - Czech koruna
	"DJF",    // Fr - Djiboutian franc
	"DKK",    // kr - Danish krone
	"DOP",    // $ - Dominican peso
	"DZD",    // د.ج - Algerian dinar
	"EGP",    // £ or ج.م - Egyptian pound
	"ERN",    // Nfk - Eritrean nakfa
	"ETB",    // Br - Ethiopian birr
	"EUR",    // € - Euro
	"FJD",    // $ - Fijian dollar
	"FKP",    // £ - Falkland Islands pound
	"GBP",    // £ - British pound[F]
	"GBP",    // £ - British pound
	"GEL",    // ₾ - Georgian lari
	"GGP[G]", // £ - Guernsey pound
	"GHS",    // ₵ - Ghanaian cedi
	"GIP",    // £ - Gibraltar pound
	"GMD",    // D - Gambian dalasi
	"GNF",    // Fr - Guinean franc
	"GTQ",    // Q - Guatemalan quetzal
	"GYD",    // $ - Guyanese dollar
	"HKD",    // $ - Hong Kong dollar
	"HNL",    // L - Honduran lempira
	"HRK",    // kn - Croatian kuna
	"HTG",    // G - Haitian gourde
	"HUF",    // Ft - Hungarian forint
	"IDR",    // Rp - Indonesian rupiah
	"ILS",    // ₪ - Israeli new shekel
	"IMP[G]", // £ - Manx pound
	"INR",    // ₹ - Indian rupee
	"IQD",    // ع.د - Iraqi dinar
	"IRR",    // ﷼ - Iranian rial
	"ISK",    // kr - Icelandic króna
	"JEP[G]", // £ - Jersey pound
	"JMD",    // $ - Jamaican dollar
	"JOD",    // د.ا - Jordanian dinar
	"JPY",    // ¥ - Japanese yen
	"KES",    // Sh - Kenyan shilling
	"KGS",    // с - Kyrgyzstani som
	"KHR",    // ៛ - Cambodian riel
	"KMF",    // Fr - Comorian franc
	"KPW",    // ₩ - North Korean won
	"KRW",    // ₩ - South Korean won
	"KWD",    // د.ك - Kuwaiti dinar
	"KYD",    // $ - Cayman Islands dollar
	"KZT",    //  - Kazakhstani tenge
	"LAK",    // ₭ - Lao kip
	"LBP",    // ل.ل - Lebanese pound
	"LKR",    // Rs or රු - Sri Lankan rupee
	"LRD",    // $ - Liberian dollar
	"LSL",    // L - Lesotho loti
	"LYD",    // ل.د - Libyan dinar
	"MAD",    // د.م. - Moroccan dirham
	"MAD",    // د. م. - Moroccan dirham
	"MDL",    // L - Moldovan leu
	"MGA",    // Ar - Malagasy ariary
	"MKD",    // ден - Macedonian denar
	"MMK",    // Ks - Burmese kyat
	"MNT",    // ₮ - Mongolian tögrög
	"MOP",    // P - Macanese pataca
	"MRO",    // UM - Mauritanian ouguiya
	"MUR",    // ₨ - Mauritian rupee
	"MVR",    // .ރ - Maldivian rufiyaa
	"MWK",    // MK - Malawian kwacha
	"MXN",    // $ - Mexican peso
	"MYR",    // RM - Malaysian ringgit
	"MZN",    // MT - Mozambican metical
	"NAD",    // $ - Namibian dollar
	"NGN",    // ₦ - Nigerian naira
	"NIO",    // C$ - Nicaraguan córdoba
	"NOK",    // kr - Norwegian krone
	"NPR",    // ₨ - Nepalese rupee
	"NZD",    // $ - New Zealand dollar
	"OMR",    // ر.ع. - Omani rial
	"PAB",    // B/. - Panamanian balboa
	"PEN",    // S/. - Peruvian sol
	"PGK",    // K - Papua New Guinean kina
	"PHP",    // ₱ - Philippine peso
	"PKR",    // ₨ - Pakistani rupee
	"PLN",    // zł - Polish złoty
	"PRB[G]", // р. - Transnistrian ruble
	"PYG",    // ₲ - Paraguayan guaraní
	"QAR",    // ر.ق - Qatari riyal
	"RON",    // lei - Romanian leu
	"RSD",    // дин. or din. - Serbian dinar
	"RUB",    //  - Russian ruble
	"RWF",    // Fr - Rwandan franc
	"SAR",    // ر.س - Saudi riyal
	"SBD",    // $ - Solomon Islands dollar
	"SCR",    // ₨ - Seychellois rupee
	"SDG",    // ج.س. - Sudanese pound
	"SEK",    // kr - Swedish krona
	"SGD",    // $ - Singapore dollar
	"SHP",    // £ - Saint Helena pound
	"SLL",    // Le - Sierra Leonean leone
	"SOS",    // Sh - Somali shilling
	"SRD",    // $ - Surinamese dollar
	"SSP",    // £ - South Sudanese pound
	"STD",    // Db - São Tomé and Príncipe dobra
	"SYP",    // £ or ل.س - Syrian pound
	"SZL",    // L - Swazi lilangeni
	"THB",    // ฿ - Thai baht
	"TJS",    // ЅМ - Tajikistani somoni
	"TMT",    // m - Turkmenistan manat
	"TND",    // د.ت - Tunisian dinar
	"TOP",    // T$ - Tongan paʻanga[P]
	"TRY",    //  - Turkish lira
	"TTD",    // $ - Trinidad and Tobago dollar
	"TVD[G]", // $ - Tuvaluan dollar
	"TWD",    // $ - New Taiwan dollar
	"TZS",    // Sh - Tanzanian shilling
	"UAH",    // ₴ - Ukrainian hryvnia
	"UGX",    // Sh - Ugandan shilling
	"USD",    // $ - United States dollar
	"UYU",    // $ - Uruguayan peso
	"UZS",    //  - Uzbekistani soʻm
	"VEF",    // Bs - Venezuelan bolívar
	"VND",    // ₫ - Vietnamese đồng
	"VUV",    // Vt - Vanuatu vatu
	"WST",    // T - Samoan tālā
	"XAF",    // Fr - Central African CFA franc
	"XCD",    // $ - East Caribbean dollar
	"XOF",    // Fr - West African CFA franc
	"XPF",    // Fr - CFP franc
	"YER",    // ﷼ - Yemeni rial
	"ZAR",    // R - South African rand
	"ZAR",    // Rs - South African rand
	"ZMW",    // ZK - Zambian kwacha
}

func (c Currency) IsMoney() bool {
	return sort.SearchStrings(currencies, string(c)) >= 0
}

var CURRENCY_USD = Currency("USD")
var CURRENCY_EUR = Currency("EUR")
var CURRENCY_GBP = Currency("GPB")
var CURRENCY_JPY = Currency("JPY")

var CURRENCY_RUB = Currency("RUB")
var CURRENCY_UAH = Currency("UAH")
var CURRENCY_BYN = Currency("BYN")
var CURRENCY_UZS = Currency("UZS")
var CURRENCY_TJS = Currency("TJS")
var CURRENCY_KZT = Currency("KZT")

var CURRENCY_IRR = Currency("IRR")

const (
	EUR_SIGN = "€"
	USD_SIGN = "$"
	GPB_SIGN = "£"
	JPY_SIGN = "¥"
	RUR_SIGN = "₽"
	IRR_SIGN = "﷼"
	UAH_SIGN = "₴"
	UZS_SIGN = "сўм"
	BYN_SIGN = "Br"
	TJS_SIGN = "смн."
	KZT_SIGN = "₸"
)

func HasCurrencyPrefix(s string) bool {
	for _, currencySign := range currencySigns {
		if strings.HasPrefix(s, currencySign) {
			return true
		}
	}
	return false
}

func CleanupCurrency(s string) Currency {
	for currency := range currencySigns {
		if currency.SignAndCode() == s || string(currency) == s {
			return currency
		}
	}
	return Currency(s)
}

var currencySigns = map[Currency]string{
	CURRENCY_USD: USD_SIGN,
	CURRENCY_EUR: EUR_SIGN,
	CURRENCY_GBP: GPB_SIGN,
	CURRENCY_IRR: IRR_SIGN,
	CURRENCY_JPY: JPY_SIGN,

	CURRENCY_RUB: RUR_SIGN,
	CURRENCY_UAH: UAH_SIGN,
	CURRENCY_BYN: BYN_SIGN,
	CURRENCY_UZS: UZS_SIGN,
	CURRENCY_TJS: TJS_SIGN,
	CURRENCY_KZT: KZT_SIGN,
}

func (c Currency) Sign() string {
	if sign, ok := currencySigns[c]; ok {
		return sign
	}
	return string(c)
}

func (c Currency) SignAndCode() string {
	if sign, ok := currencySigns[c]; ok {
		return fmt.Sprintf("%v %v", sign, c)
	}
	return string(c)
}
