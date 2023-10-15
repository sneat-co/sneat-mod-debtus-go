module github.com/sneat-co/sneat-mod-debtus-go

go 1.20

require (
	cloud.google.com/go/firestore v1.13.0
	github.com/aws/aws-sdk-go v1.45.25
	github.com/bots-go-framework/bots-api-telegram v0.4.1
	github.com/bots-go-framework/bots-fw v0.21.5
	github.com/bots-go-framework/bots-fw-store v0.1.0
	github.com/bots-go-framework/bots-fw-telegram v0.6.21
	github.com/bots-go-framework/bots-fw-telegram-models v0.0.13
	github.com/bots-go-framework/bots-go-core v0.0.1
	github.com/bots-go-framework/bots-host-gae v0.4.16
	github.com/captaincodeman/datastore-mapper v0.0.0-20170320145307-cb380a4c4d13
	github.com/crediterra/go-interest v0.0.0-20180510115340-54da66993b85
	github.com/crediterra/money v0.2.1
	github.com/dal-go/dalgo v0.12.0
	github.com/dal-go/dalgo2firestore v0.1.32
	github.com/dal-go/mocks4dalgo v0.1.17
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/gorilla/sessions v1.2.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/matryer/is v1.4.1
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/sanity-io/litter v1.5.5
	github.com/sendgrid/sendgrid-go v3.13.0+incompatible
	github.com/shiyanhui/hero v0.0.2
	github.com/sneat-co/debtstracker-translations v0.0.14
	github.com/strongo/app v0.5.7
	github.com/strongo/app-host-gae v0.1.16
	github.com/strongo/decimal v0.0.1
	github.com/strongo/delaying v0.0.1
	github.com/strongo/facebook v1.8.1
	github.com/strongo/gamp v0.0.1
	github.com/strongo/gotwilio v0.0.0-20160123000810-f024bbefe80f
	github.com/strongo/i18n v0.0.4
	github.com/strongo/log v0.3.0
	github.com/strongo/random v0.0.1
	github.com/strongo/slice v0.1.4
	github.com/strongo/slices v0.0.0-20180713073818-553769fcb80b
	github.com/strongo/validation v0.0.5
	github.com/yaa110/go-persian-calendar v1.1.5
	golang.org/x/crypto v0.14.0
	golang.org/x/net v0.17.0
	google.golang.org/appengine/v2 v2.0.5
)

//replace github.com/dal-go/dalgo => ../dal-go/dalgo
//replace github.com/bots-go-framework/bots-fw => ../bots-go-framework/bots-fw
//replace github.com/bots-go-framework/bots-fw-telegram => ../bots-go-framework/bots-fw-telegram

require (
	cloud.google.com/go v0.110.8 // indirect
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/datastore v1.15.0 // indirect
	cloud.google.com/go/iam v1.1.2 // indirect
	cloud.google.com/go/longrunning v0.5.1 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/captaincodeman/datastore-locker v0.0.0-20170308203336-4eddc467484e // indirect
	github.com/dal-go/dalgo2datastore v0.0.29 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.1 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/oauth2 v0.13.0 // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.147.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231009173412-8bfb1ae86b6c // indirect
	google.golang.org/grpc v1.58.3 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
