module github.com/sneat-co/debtstracker-go

go 1.20

require (
	cloud.google.com/go/firestore v1.10.0
	github.com/aws/aws-sdk-go v1.44.282
	github.com/bots-go-framework/bots-api-telegram v0.4.1
	github.com/bots-go-framework/bots-fw v0.17.2
	github.com/bots-go-framework/bots-fw-store v0.0.7
	github.com/bots-go-framework/bots-fw-telegram v0.6.13
	github.com/bots-go-framework/bots-fw-telegram-models v0.0.12
	github.com/bots-go-framework/bots-go-core v0.0.1
	github.com/bots-go-framework/bots-host-gae v0.4.4
	github.com/bots-go-framework/dalgo4botsfw v0.0.14
	github.com/captaincodeman/datastore-mapper v0.0.0-20170320145307-cb380a4c4d13
	github.com/crediterra/go-interest v0.0.0-20180510115340-54da66993b85
	github.com/crediterra/money v0.0.1
	github.com/dal-go/dalgo v0.2.31
	github.com/dal-go/mocks4dalgo v0.1.11
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/gorilla/sessions v1.2.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/matryer/is v1.4.1
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/sanity-io/litter v1.5.5
	github.com/sendgrid/sendgrid-go v3.12.0+incompatible
	github.com/shiyanhui/hero v0.0.2
	github.com/sneat-co/debtstracker-translations v0.0.8
	github.com/strongo/app v0.5.7
	github.com/strongo/app-host-gae v0.1.14
	github.com/strongo/decimal v0.0.0-20180523215323-a1521d8f65fa
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
	github.com/yaa110/go-persian-calendar v1.1.4
	golang.org/x/crypto v0.10.0
	golang.org/x/net v0.11.0
	google.golang.org/appengine v1.6.7
)

//replace github.com/dal-go/dalgo => ../dal-go/dalgo
//replace github.com/bots-go-framework/bots-fw => ../bots-go-framework/bots-fw
//replace github.com/bots-go-framework/bots-fw-telegram => ../bots-go-framework/bots-fw-telegram

require (
	cloud.google.com/go v0.110.2 // indirect
	cloud.google.com/go/compute v1.19.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/datastore v1.11.0 // indirect
	cloud.google.com/go/iam v1.0.0 // indirect
	cloud.google.com/go/longrunning v0.4.2 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/captaincodeman/datastore-locker v0.0.0-20170308203336-4eddc467484e // indirect
	github.com/dal-go/dalgo2datastore v0.0.12 // indirect
	github.com/dal-go/dalgo2firestore v0.1.9 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.4 // indirect
	github.com/googleapis/gax-go/v2 v2.10.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.127.0 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.55.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
