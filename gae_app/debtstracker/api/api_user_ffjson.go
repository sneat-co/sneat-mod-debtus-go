// DO NOT EDIT!
// Code generated by ffjson <https://github.com/pquerna/ffjson>
// source: api_user.go
// DO NOT EDIT!

package api

import (
	"bytes"
	"fmt"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

// MarshalJSON marshal bytes to json - template
func (j *UserMeDto) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if j == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := j.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalJSONBuf marshal buff to json - template
func (j *UserMeDto) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{ "UserID":`)
	fflib.FormatBits2(buf, uint64(j.UserID), 10, j.UserID < 0)
	buf.WriteByte(',')
	if len(j.FullName) != 0 {
		buf.WriteString(`"FullName":`)
		fflib.WriteJsonString(buf, string(j.FullName))
		buf.WriteByte(',')
	}
	if len(j.GoogleUserID) != 0 {
		buf.WriteString(`"GoogleUserID":`)
		fflib.WriteJsonString(buf, string(j.GoogleUserID))
		buf.WriteByte(',')
	}
	if len(j.FbUserID) != 0 {
		buf.WriteString(`"FbUserID":`)
		fflib.WriteJsonString(buf, string(j.FbUserID))
		buf.WriteByte(',')
	}
	if j.VkUserID != 0 {
		buf.WriteString(`"VkUserID":`)
		fflib.FormatBits2(buf, uint64(j.VkUserID), 10, j.VkUserID < 0)
		buf.WriteByte(',')
	}
	if len(j.ViberUserID) != 0 {
		buf.WriteString(`"ViberUserID":`)
		fflib.WriteJsonString(buf, string(j.ViberUserID))
		buf.WriteByte(',')
	}
	buf.Rewind(1)
	buf.WriteByte('}')
	return nil
}

const (
	ffjtUserMeDtobase = iota
	ffjtUserMeDtonosuchkey

	ffjtUserMeDtoUserID

	ffjtUserMeDtoFullName

	ffjtUserMeDtoGoogleUserID

	ffjtUserMeDtoFbUserID

	ffjtUserMeDtoVkUserID

	ffjtUserMeDtoViberUserID
)

var ffjKeyUserMeDtoUserID = []byte("UserID")

var ffjKeyUserMeDtoFullName = []byte("FullName")

var ffjKeyUserMeDtoGoogleUserID = []byte("GoogleUserID")

var ffjKeyUserMeDtoFbUserID = []byte("FbUserID")

var ffjKeyUserMeDtoVkUserID = []byte("VkUserID")

var ffjKeyUserMeDtoViberUserID = []byte("ViberUserID")

// UnmarshalJSON umarshall json - template of ffjson
func (j *UserMeDto) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *UserMeDto) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtUserMeDtobase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffjtUserMeDtonosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'F':

					if bytes.Equal(ffjKeyUserMeDtoFullName, kn) {
						currentKey = ffjtUserMeDtoFullName
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyUserMeDtoFbUserID, kn) {
						currentKey = ffjtUserMeDtoFbUserID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'G':

					if bytes.Equal(ffjKeyUserMeDtoGoogleUserID, kn) {
						currentKey = ffjtUserMeDtoGoogleUserID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'U':

					if bytes.Equal(ffjKeyUserMeDtoUserID, kn) {
						currentKey = ffjtUserMeDtoUserID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'V':

					if bytes.Equal(ffjKeyUserMeDtoVkUserID, kn) {
						currentKey = ffjtUserMeDtoVkUserID
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyUserMeDtoViberUserID, kn) {
						currentKey = ffjtUserMeDtoViberUserID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffjKeyUserMeDtoViberUserID, kn) {
					currentKey = ffjtUserMeDtoViberUserID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyUserMeDtoVkUserID, kn) {
					currentKey = ffjtUserMeDtoVkUserID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyUserMeDtoFbUserID, kn) {
					currentKey = ffjtUserMeDtoFbUserID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyUserMeDtoGoogleUserID, kn) {
					currentKey = ffjtUserMeDtoGoogleUserID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyUserMeDtoFullName, kn) {
					currentKey = ffjtUserMeDtoFullName
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyUserMeDtoUserID, kn) {
					currentKey = ffjtUserMeDtoUserID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtUserMeDtonosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffjtUserMeDtoUserID:
					goto handle_UserID

				case ffjtUserMeDtoFullName:
					goto handle_FullName

				case ffjtUserMeDtoGoogleUserID:
					goto handle_GoogleUserID

				case ffjtUserMeDtoFbUserID:
					goto handle_FbUserID

				case ffjtUserMeDtoVkUserID:
					goto handle_VkUserID

				case ffjtUserMeDtoViberUserID:
					goto handle_ViberUserID

				case ffjtUserMeDtonosuchkey:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_UserID:

	/* handler: j.UserID type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			j.UserID = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_FullName:

	/* handler: j.FullName type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.FullName = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_GoogleUserID:

	/* handler: j.GoogleUserID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.GoogleUserID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_FbUserID:

	/* handler: j.FbUserID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.FbUserID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_VkUserID:

	/* handler: j.VkUserID type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			j.VkUserID = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ViberUserID:

	/* handler: j.ViberUserID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.ViberUserID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:

	return nil
}
