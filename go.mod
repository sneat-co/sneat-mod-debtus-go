module github.com/sneat-co/sneat-mod-debtus-go

go 1.22.7

toolchain go1.22.11

//replace github.com/sneat-co/sneat-core-modules => ../sneat-core-modules

require (
	github.com/crediterra/go-interest v0.0.0-20180510115340-54da66993b85
	github.com/crediterra/money v0.3.0
	github.com/dal-go/dalgo v0.14.2
	github.com/dal-go/mocks4dalgo v0.1.28
	github.com/julienschmidt/httprouter v1.3.0
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/sanity-io/litter v1.5.6
	github.com/shiyanhui/hero v0.0.2
	github.com/sneat-co/debtstracker-translations v0.3.0
	github.com/sneat-co/sneat-core-modules v0.18.0
	github.com/sneat-co/sneat-go-core v0.42.0
	github.com/strongo/decimal v0.1.1
	github.com/strongo/delaying v0.1.0
	github.com/strongo/gamp v0.0.1
	github.com/strongo/gotwilio v0.0.0-20160123000810-f024bbefe80f
	github.com/strongo/i18n v0.6.1
	github.com/strongo/logus v0.2.1
	github.com/strongo/random v0.0.1
	github.com/strongo/slice v0.3.1
	github.com/strongo/strongoapp v0.25.5
	github.com/strongo/validation v0.0.7
	golang.org/x/net v0.34.0
)

require (
	github.com/alexsergivan/transliterator v1.0.1 // indirect
	github.com/bots-go-framework/bots-fw-store v0.8.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/gosimple/slug v1.15.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	golang.org/x/crypto v0.32.0 // indirect
)
