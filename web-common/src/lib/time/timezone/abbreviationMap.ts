/***
 * There is no standard way of getting the abbreviation for a timezone. 
 * The abbreviation are not unique and can change over time.
 * Luxon requires passing the regional user locale to fetch the correct abbreviation 

 * Use luxon to get the timezone name and use this mapping to get the abbreviation

 * This list is derived from https://www.timeanddate.com/time/zones/
 */

export const timeZoneNameToAbbreviationMap = {
  "Alfa Time Zone": "A",
  "Australian Central Daylight Time": "ACDT",
  "Australian Central Standard Time": "ACST",
  "Acre Time": "ACT",
  "Australian Central Time": "ACT",
  "Australian Central Western Standard Time": "ACWST",
  "Arabia Daylight Time": "ADT",
  "Atlantic Daylight Time": "ADT",
  "Australian Eastern Daylight Time": "AEDT",
  "Australian Eastern Standard Time": "AEST",
  "Australian Eastern Time": "AET",
  "Afghanistan Time": "AFT",
  "Alaska Daylight Time": "AKDT",
  "Alaska Standard Time": "AKST",
  "Alma-Ata Time": "ALMT",
  "Amazon Summer Time": "AMST",
  "Armenia Summer Time": "AMST",
  "Amazon Time": "AMT",
  "Armenia Time": "AMT",
  "Anadyr Summer Time": "ANAST",
  "Anadyr Time": "ANAT",
  "Aqtobe Time": "AQTT",
  "Argentina Time": "ART",
  "Arabia Standard Time": "AST",
  "Atlantic Standard Time": "AST",
  "Atlantic Time": "AT",
  "Australian Western Daylight Time": "AWDT",
  "Australian Western Standard Time": "AWST",
  "Azores Summer Time": "AZOST",
  "Azores Time": "AZOT",
  "Azerbaijan Summer Time": "AZST",
  "Azerbaijan Time": "AZT",
  "Anywhere on Earth": "AoE",
  "Bravo Time Zone": "B",
  "Brunei Darussalam Time": "BNT",
  "Bolivia Time": "BOT",
  "Brasília Summer Time": "BRST",
  "Brasília Time": "BRT",
  "Bangladesh Standard Time": "BST",
  "Bougainville Standard Time": "BST",
  "British Summer Time": "BST",
  "Bhutan Time": "BTT",
  "Charlie Time Zone": "C",
  "Casey Time": "CAST",
  "Central Africa Time": "CAT",
  "Cocos Islands Time": "CCT",
  "Central Daylight Time": "CDT",
  "Cuba Daylight Time": "CDT",
  "Central European Summer Time": "CEST",
  "Central European Standard Time": "CET",
  "Central European Time": "CET",
  "Chatham Island Daylight Time": "CHADT",
  "Chatham Island Standard Time": "CHAST",
  "Choibalsan Summer Time": "CHOST",
  "Choibalsan Time": "CHOT",
  "Chuuk Time": "CHUT",
  "Cayman Islands Daylight Saving Time": "CIDST",
  "Cayman Islands Standard Time": "CIST",
  "Cook Island Time": "CKT",
  "Chile Summer Time": "CLST",
  "Chile Standard Time": "CLT",
  "Colombia Time": "COT",
  "Central Standard Time": "CST",
  "China Standard Time": "CST",
  "Cuba Standard Time": "CST",
  "Central Time": "CT",
  "Cape Verde Time": "CVT",
  "Christmas Island Time": "CXT",
  "Chamorro Standard Time": "ChST",
  "Delta Time Zone": "D",
  "Davis Time": "DAVT",
  "Dumont-d'Urville Time": "DDUT",
  "Echo Time Zone": "E",
  "Easter Island Summer Time": "EASST",
  "Easter Island Standard Time": "EAST",
  "Eastern Africa Time": "EAT",
  "Ecuador Time": "ECT",
  "Eastern Daylight Time": "EDT",
  "Eastern European Summer Time": "EEST",
  "Eastern European Time": "EET",
  "Eastern Greenland Summer Time": "EGST",
  "East Greenland Time": "EGT",
  "Eastern Standard Time": "EST",
  "Eastern Time": "ET",
  "Foxtrot Time Zone": "F",
  "Further-Eastern European Time": "FET",
  "Fiji Summer Time": "FJST",
  "Fiji Time": "FJT",
  "Falkland Islands Summer Time": "FKST",
  "Falkland Island Time": "FKT",
  "Fernando de Noronha Time": "FNT",
  "Golf Time Zone": "G",
  "Galapagos Time": "GALT",
  "Gambier Time": "GAMT",
  "Georgia Standard Time": "GET",
  "French Guiana Time": "GFT",
  "Gilbert Island Time": "GILT",
  "Greenwich Mean Time": "GMT",
  "Gulf Standard Time": "GST",
  "South Georgia Time": "GST",
  "Guyana Time": "GYT",
  "Hotel Time Zone": "H",
  "Hawaii-Aleutian Daylight Time": "HDT",
  "Hong Kong Time": "HKT",
  "Hovd Summer Time": "HOVST",
  "Hovd Time": "HOVT",
  "Hawaii Standard Time": "HST",
  "India Time Zone": "I",
  "Indochina Time": "ICT",
  "Israel Daylight Time": "IDT",
  "Indian Chagos Time": "IOT",
  "Iran Daylight Time": "IRDT",
  "Irkutsk Summer Time": "IRKST",
  "Irkutsk Time": "IRKT",
  "Iran Standard Time": "IRST",
  "India Standard Time": "IST",
  "Irish Standard Time": "IST",
  "Israel Standard Time": "IST",
  "Japan Standard Time": "JST",
  "Kilo Time Zone": "K",
  "Kyrgyzstan Time": "KGT",
  "Kosrae Time": "KOST",
  "Krasnoyarsk Summer Time": "KRAST",
  "Krasnoyarsk Time": "KRAT",
  "Korea Standard Time": "KST",
  "Kuybyshev Time": "KUYT",
  "Lima Time Zone": "L",
  "Lord Howe Daylight Time": "LHDT",
  "Lord Howe Standard Time": "LHST",
  "Line Islands Time": "LINT",
  "Mike Time Zone": "M",
  "Magadan Summer Time": "MAGST",
  "Magadan Time": "MAGT",
  "Marquesas Time": "MART",
  "Mawson Time": "MAWT",
  "Mountain Daylight Time": "MDT",
  "Marshall Islands Time": "MHT",
  "Myanmar Time": "MMT",
  "Moscow Daylight Time": "MSD",
  "Moscow Standard Time": "MSK",
  "Mountain Standard Time": "MST",
  "Mountain Time": "MT",
  "Mauritius Time": "MUT",
  "Maldives Time": "MVT",
  "Malaysia Time": "MYT",
  "November Time Zone": "N",
  "New Caledonia Time": "NCT",
  "Newfoundland Daylight Time": "NDT",
  "Norfolk Daylight Time": "NFDT",
  "Norfolk Time": "NFT",
  "Novosibirsk Summer Time": "NOVST",
  "Novosibirsk Time": "NOVT",
  "Nepal Time": "NPT",
  "Nauru Time": "NRT",
  "Newfoundland Standard Time": "NST",
  "Niue Time": "NUT",
  "New Zealand Daylight Time": "NZDT",
  "New Zealand Standard Time": "NZST",
  "Oscar Time Zone": "O",
  "Omsk Summer Time": "OMSST",
  "Omsk Standard Time": "OMST",
  "Oral Time": "ORAT",
  "Papa Time Zone": "P",
  "Pacific Daylight Time": "PDT",
  "Peru Time": "PET",
  "Kamchatka Summer Time": "PETST",
  "Kamchatka Time": "PETT",
  "Papua New Guinea Time": "PGT",
  "Phoenix Island Time": "PHOT",
  "Philippine Time": "PHT",
  "Pakistan Standard Time": "PKT",
  "Pierre & Miquelon Daylight Time": "PMDT",
  "Pierre & Miquelon Standard Time": "PMST",
  "Pohnpei Standard Time": "PONT",
  "Pacific Standard Time": "PST",
  "Pitcairn Standard Time": "PST",
  "Pacific Time": "PT",
  "Palau Time": "PWT",
  "Paraguay Summer Time": "PYST",
  "Paraguay Time": "PYT",
  "Pyongyang Time": "PYT",
  "Quebec Time Zone": "Q",
  "Qyzylorda Time": "QYZT",
  "Romeo Time Zone": "R",
  "Reunion Time": "RET",
  "Rothera Time": "ROTT",
  "Sierra Time Zone": "S",
  "Sakhalin Time": "SAKT",
  "Samara Time": "SAMT",
  "South Africa Standard Time": "SAST",
  "Solomon Islands Time": "SBT",
  "Seychelles Time": "SCT",
  "Singapore Time": "SGT",
  "Srednekolymsk Time": "SRET",
  "Suriname Time": "SRT",
  "Samoa Standard Time": "SST",
  "Syowa Time": "SYOT",
  "Tango Time Zone": "T",
  "Tahiti Time": "TAHT",
  "French Southern and Antarctic Time": "TFT",
  "Tajikistan Time": "TJT",
  "Tokelau Time": "TKT",
  "East Timor Time": "TLT",
  "Turkmenistan Time": "TMT",
  "Tonga Summer Time": "TOST",
  "Tonga Time": "TOT",
  "Turkey Time": "TRT",
  "Tuvalu Time": "TVT",
  "Uniform Time Zone": "U",
  "Ulaanbaatar Summer Time": "ULAST",
  "Ulaanbaatar Time": "ULAT",
  "Coordinated Universal Time": "UTC",
  "Uruguay Summer Time": "UYST",
  "Uruguay Time": "UYT",
  "Uzbekistan Time": "UZT",
  "Victor Time Zone": "V",
  "Venezuelan Standard Time": "VET",
  "Vladivostok Summer Time": "VLAST",
  "Vladivostok Time": "VLAT",
  "Vostok Time": "VOST",
  "Vanuatu Time": "VUT",
  "Whiskey Time Zone": "W",
  "Wake Time": "WAKT",
  "Western Argentine Summer Time": "WARST",
  "West Africa Summer Time": "WAST",
  "West Africa Time": "WAT",
  "Western European Summer Time": "WEST",
  "Western European Time": "WET",
  "Wallis and Futuna Time": "WFT",
  "Western Greenland Summer Time": "WGST",
  "West Greenland Time": "WGT",
  "Western Indonesian Time": "WIB",
  "Eastern Indonesian Time": "WIT",
  "Central Indonesian Time": "WITA",
  "West Samoa Time": "WST",
  "Western Sahara Summer Time": "WST",
  "Western Sahara Standard Time": "WT",
  "X-ray Time Zone": "X",
  "Yankee Time Zone": "Y",
  "Yakutsk Summer Time": "YAKST",
  "Yakutsk Time": "YAKT",
  "Yap Time": "YAPT",
  "Yekaterinburg Summer Time": "YEKST",
  "Yekaterinburg Time": "YEKT",
  "Zulu Time Zone": "Z",
};
