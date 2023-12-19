module github.com/sneat-co/sneat-mod-debtus-go

go 1.21

toolchain go1.21.4

//replace github.com/dal-go/dalgo => ../dal-go/dalgo
//replace github.com/bots-go-framework/bots-fw-store => ../../bots-go-framework/bots-fw-store

//replace github.com/bots-go-framework/bots-fw => ../../bots-go-framework/bots-fw

//replace github.com/bots-go-framework/bots-fw-telegram => ../../bots-go-framework/bots-fw-telegram

require (
	github.com/bots-go-framework/bots-api-telegram v0.4.2
	github.com/bots-go-framework/bots-fw v0.23.3
	github.com/bots-go-framework/bots-fw-store v0.4.0
	github.com/bots-go-framework/bots-fw-telegram v0.7.2
	github.com/bots-go-framework/bots-fw-telegram-models v0.0.16
	github.com/bots-go-framework/bots-go-core v0.0.2
	github.com/bots-go-framework/bots-host-gae v0.5.3
	github.com/captaincodeman/datastore-mapper v0.0.0-20170320145307-cb380a4c4d13
	github.com/crediterra/go-interest v0.0.0-20180510115340-54da66993b85
	github.com/crediterra/money v0.2.1
	github.com/dal-go/dalgo v0.12.0
	github.com/dal-go/mocks4dalgo v0.1.17
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/gorilla/sessions v1.2.2
	github.com/julienschmidt/httprouter v1.3.0
	github.com/matryer/is v1.4.1
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/sanity-io/litter v1.5.5
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
	github.com/shiyanhui/hero v0.0.2
	github.com/sneat-co/debtstracker-translations v0.0.17
	github.com/sneat-co/sneat-go-core v0.20.0
	github.com/strongo/app-host-gae v0.1.18
	github.com/strongo/decimal v0.0.1
	github.com/strongo/delaying v0.0.1
	github.com/strongo/facebook v1.8.1
	github.com/strongo/gamp v0.0.1
	github.com/strongo/gotwilio v0.0.0-20160123000810-f024bbefe80f
	github.com/strongo/i18n v0.0.4
	github.com/strongo/log v0.3.0
	github.com/strongo/random v0.0.1
	github.com/strongo/slice v0.1.4
	github.com/strongo/slices v0.0.0-20231201223919-29a6c669158a
	github.com/strongo/strongoapp v0.10.0
	github.com/strongo/validation v0.0.6
	github.com/yaa110/go-persian-calendar v1.1.5
	golang.org/x/crypto v0.17.0
	golang.org/x/net v0.19.0
	google.golang.org/appengine/v2 v2.0.5
)

require (
	cloud.google.com/go v0.110.10 // indirect
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.5 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/captaincodeman/datastore-locker v0.0.0-20170308203336-4eddc467484e // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/oauth2 v0.14.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.153.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
