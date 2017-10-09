// DO NOT EDIT!
// Code generated by ffjson <https://github.com/pquerna/ffjson>
// source: api_receipt.go
// DO NOT EDIT!

package api

import (
	"bytes"
	"errors"
	"fmt"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

// MarshalJSON marshal bytes to json - template
func (j *ApiAcknowledgeDto) MarshalJSON() ([]byte, error) {
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
func (j *ApiAcknowledgeDto) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{"Status":`)
	fflib.WriteJsonString(buf, string(j.Status))
	buf.WriteString(`,"UnixTime":`)
	fflib.FormatBits2(buf, uint64(j.UnixTime), 10, j.UnixTime < 0)
	buf.WriteByte('}')
	return nil
}

const (
	ffjtApiAcknowledgeDtobase = iota
	ffjtApiAcknowledgeDtonosuchkey

	ffjtApiAcknowledgeDtoStatus

	ffjtApiAcknowledgeDtoUnixTime
)

var ffjKeyApiAcknowledgeDtoStatus = []byte("Status")

var ffjKeyApiAcknowledgeDtoUnixTime = []byte("UnixTime")

// UnmarshalJSON umarshall json - template of ffjson
func (j *ApiAcknowledgeDto) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *ApiAcknowledgeDto) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtApiAcknowledgeDtobase
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
				currentKey = ffjtApiAcknowledgeDtonosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'S':

					if bytes.Equal(ffjKeyApiAcknowledgeDtoStatus, kn) {
						currentKey = ffjtApiAcknowledgeDtoStatus
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'U':

					if bytes.Equal(ffjKeyApiAcknowledgeDtoUnixTime, kn) {
						currentKey = ffjtApiAcknowledgeDtoUnixTime
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiAcknowledgeDtoUnixTime, kn) {
					currentKey = ffjtApiAcknowledgeDtoUnixTime
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyApiAcknowledgeDtoStatus, kn) {
					currentKey = ffjtApiAcknowledgeDtoStatus
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtApiAcknowledgeDtonosuchkey
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

				case ffjtApiAcknowledgeDtoStatus:
					goto handle_Status

				case ffjtApiAcknowledgeDtoUnixTime:
					goto handle_UnixTime

				case ffjtApiAcknowledgeDtonosuchkey:
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

handle_Status:

	/* handler: j.Status type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Status = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_UnixTime:

	/* handler: j.UnixTime type=int64 kind=int64 quoted=false*/

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

			j.UnixTime = int64(tval)

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

// MarshalJSON marshal bytes to json - template
func (j *ApiReceiptDto) MarshalJSON() ([]byte, error) {
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
func (j *ApiReceiptDto) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{ "Id":`)
	fflib.FormatBits2(buf, uint64(j.ID), 10, j.ID < 0)
	buf.WriteString(`,"Code":`)
	fflib.WriteJsonString(buf, string(j.Code))
	buf.WriteString(`,"Transfer":`)

	{

		err = j.Transfer.MarshalJSONBuf(buf)
		if err != nil {
			return err
		}

	}
	buf.WriteString(`,"SentVia":`)
	fflib.WriteJsonString(buf, string(j.SentVia))
	buf.WriteByte(',')
	if len(j.SentTo) != 0 {
		buf.WriteString(`"SentTo":`)
		fflib.WriteJsonString(buf, string(j.SentTo))
		buf.WriteByte(',')
	}
	buf.Rewind(1)
	buf.WriteByte('}')
	return nil
}

const (
	ffjtApiReceiptDtobase = iota
	ffjtApiReceiptDtonosuchkey

	ffjtApiReceiptDtoID

	ffjtApiReceiptDtoCode

	ffjtApiReceiptDtoTransfer

	ffjtApiReceiptDtoSentVia

	ffjtApiReceiptDtoSentTo
)

var ffjKeyApiReceiptDtoID = []byte("Id")

var ffjKeyApiReceiptDtoCode = []byte("Code")

var ffjKeyApiReceiptDtoTransfer = []byte("Transfer")

var ffjKeyApiReceiptDtoSentVia = []byte("SentVia")

var ffjKeyApiReceiptDtoSentTo = []byte("SentTo")

// UnmarshalJSON umarshall json - template of ffjson
func (j *ApiReceiptDto) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *ApiReceiptDto) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtApiReceiptDtobase
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
				currentKey = ffjtApiReceiptDtonosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'C':

					if bytes.Equal(ffjKeyApiReceiptDtoCode, kn) {
						currentKey = ffjtApiReceiptDtoCode
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'I':

					if bytes.Equal(ffjKeyApiReceiptDtoID, kn) {
						currentKey = ffjtApiReceiptDtoID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'S':

					if bytes.Equal(ffjKeyApiReceiptDtoSentVia, kn) {
						currentKey = ffjtApiReceiptDtoSentVia
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyApiReceiptDtoSentTo, kn) {
						currentKey = ffjtApiReceiptDtoSentTo
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'T':

					if bytes.Equal(ffjKeyApiReceiptDtoTransfer, kn) {
						currentKey = ffjtApiReceiptDtoTransfer
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffjKeyApiReceiptDtoSentTo, kn) {
					currentKey = ffjtApiReceiptDtoSentTo
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyApiReceiptDtoSentVia, kn) {
					currentKey = ffjtApiReceiptDtoSentVia
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyApiReceiptDtoTransfer, kn) {
					currentKey = ffjtApiReceiptDtoTransfer
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptDtoCode, kn) {
					currentKey = ffjtApiReceiptDtoCode
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptDtoID, kn) {
					currentKey = ffjtApiReceiptDtoID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtApiReceiptDtonosuchkey
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

				case ffjtApiReceiptDtoID:
					goto handle_ID

				case ffjtApiReceiptDtoCode:
					goto handle_Code

				case ffjtApiReceiptDtoTransfer:
					goto handle_Transfer

				case ffjtApiReceiptDtoSentVia:
					goto handle_SentVia

				case ffjtApiReceiptDtoSentTo:
					goto handle_SentTo

				case ffjtApiReceiptDtonosuchkey:
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

handle_ID:

	/* handler: j.ID type=int64 kind=int64 quoted=false*/

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

			j.ID = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Code:

	/* handler: j.Code type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Code = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Transfer:

	/* handler: j.Transfer type=api.ApiReceiptTransferDto kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		err = j.Transfer.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_SentVia:

	/* handler: j.SentVia type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.SentVia = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_SentTo:

	/* handler: j.SentTo type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.SentTo = string(string(outBuf))

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

// MarshalJSON marshal bytes to json - template
func (j *ApiReceiptTransferDto) MarshalJSON() ([]byte, error) {
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
func (j *ApiReceiptTransferDto) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{ "Id":`)
	fflib.FormatBits2(buf, uint64(j.ID), 10, j.ID < 0)
	buf.WriteString(`,"Amount":`)

	{

		err = j.Amount.MarshalJSONBuf(buf)
		if err != nil {
			return err
		}

	}
	buf.WriteString(`,"From":`)

	{

		err = j.From.MarshalJSONBuf(buf)
		if err != nil {
			return err
		}

	}
	buf.WriteString(`,"To":`)

	{

		err = j.To.MarshalJSONBuf(buf)
		if err != nil {
			return err
		}

	}
	if j.IsOutstanding {
		buf.WriteString(`,"IsOutstanding":true`)
	} else {
		buf.WriteString(`,"IsOutstanding":false`)
	}
	buf.WriteString(`,"Creator":`)

	{

		err = j.Creator.MarshalJSONBuf(buf)
		if err != nil {
			return err
		}

	}
	buf.WriteByte(',')
	if len(j.CreatorComment) != 0 {
		buf.WriteString(`"CreatorComment":`)
		fflib.WriteJsonString(buf, string(j.CreatorComment))
		buf.WriteByte(',')
	}
	if j.Acknowledge != nil {
		if true {
			buf.WriteString(`"Acknowledge":`)

			{

				err = j.Acknowledge.MarshalJSONBuf(buf)
				if err != nil {
					return err
				}

			}
			buf.WriteByte(',')
		}
	}
	buf.Rewind(1)
	buf.WriteByte('}')
	return nil
}

const (
	ffjtApiReceiptTransferDtobase = iota
	ffjtApiReceiptTransferDtonosuchkey

	ffjtApiReceiptTransferDtoID

	ffjtApiReceiptTransferDtoAmount

	ffjtApiReceiptTransferDtoFrom

	ffjtApiReceiptTransferDtoTo

	ffjtApiReceiptTransferDtoIsOutstanding

	ffjtApiReceiptTransferDtoCreator

	ffjtApiReceiptTransferDtoCreatorComment

	ffjtApiReceiptTransferDtoAcknowledge
)

var ffjKeyApiReceiptTransferDtoID = []byte("Id")

var ffjKeyApiReceiptTransferDtoAmount = []byte("Amount")

var ffjKeyApiReceiptTransferDtoFrom = []byte("From")

var ffjKeyApiReceiptTransferDtoTo = []byte("To")

var ffjKeyApiReceiptTransferDtoIsOutstanding = []byte("IsOutstanding")

var ffjKeyApiReceiptTransferDtoCreator = []byte("Creator")

var ffjKeyApiReceiptTransferDtoCreatorComment = []byte("CreatorComment")

var ffjKeyApiReceiptTransferDtoAcknowledge = []byte("Acknowledge")

// UnmarshalJSON umarshall json - template of ffjson
func (j *ApiReceiptTransferDto) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *ApiReceiptTransferDto) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtApiReceiptTransferDtobase
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
				currentKey = ffjtApiReceiptTransferDtonosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'A':

					if bytes.Equal(ffjKeyApiReceiptTransferDtoAmount, kn) {
						currentKey = ffjtApiReceiptTransferDtoAmount
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyApiReceiptTransferDtoAcknowledge, kn) {
						currentKey = ffjtApiReceiptTransferDtoAcknowledge
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'C':

					if bytes.Equal(ffjKeyApiReceiptTransferDtoCreator, kn) {
						currentKey = ffjtApiReceiptTransferDtoCreator
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyApiReceiptTransferDtoCreatorComment, kn) {
						currentKey = ffjtApiReceiptTransferDtoCreatorComment
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'F':

					if bytes.Equal(ffjKeyApiReceiptTransferDtoFrom, kn) {
						currentKey = ffjtApiReceiptTransferDtoFrom
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'I':

					if bytes.Equal(ffjKeyApiReceiptTransferDtoID, kn) {
						currentKey = ffjtApiReceiptTransferDtoID
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyApiReceiptTransferDtoIsOutstanding, kn) {
						currentKey = ffjtApiReceiptTransferDtoIsOutstanding
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'T':

					if bytes.Equal(ffjKeyApiReceiptTransferDtoTo, kn) {
						currentKey = ffjtApiReceiptTransferDtoTo
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffjKeyApiReceiptTransferDtoAcknowledge, kn) {
					currentKey = ffjtApiReceiptTransferDtoAcknowledge
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptTransferDtoCreatorComment, kn) {
					currentKey = ffjtApiReceiptTransferDtoCreatorComment
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptTransferDtoCreator, kn) {
					currentKey = ffjtApiReceiptTransferDtoCreator
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyApiReceiptTransferDtoIsOutstanding, kn) {
					currentKey = ffjtApiReceiptTransferDtoIsOutstanding
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptTransferDtoTo, kn) {
					currentKey = ffjtApiReceiptTransferDtoTo
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptTransferDtoFrom, kn) {
					currentKey = ffjtApiReceiptTransferDtoFrom
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptTransferDtoAmount, kn) {
					currentKey = ffjtApiReceiptTransferDtoAmount
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiReceiptTransferDtoID, kn) {
					currentKey = ffjtApiReceiptTransferDtoID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtApiReceiptTransferDtonosuchkey
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

				case ffjtApiReceiptTransferDtoID:
					goto handle_ID

				case ffjtApiReceiptTransferDtoAmount:
					goto handle_Amount

				case ffjtApiReceiptTransferDtoFrom:
					goto handle_From

				case ffjtApiReceiptTransferDtoTo:
					goto handle_To

				case ffjtApiReceiptTransferDtoIsOutstanding:
					goto handle_IsOutstanding

				case ffjtApiReceiptTransferDtoCreator:
					goto handle_Creator

				case ffjtApiReceiptTransferDtoCreatorComment:
					goto handle_CreatorComment

				case ffjtApiReceiptTransferDtoAcknowledge:
					goto handle_Acknowledge

				case ffjtApiReceiptTransferDtonosuchkey:
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

handle_ID:

	/* handler: j.ID type=int64 kind=int64 quoted=false*/

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

			j.ID = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Amount:

	/* handler: j.Amount type=models.Amount kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		err = j.Amount.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_From:

	/* handler: j.From type=api.ContactDto kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		err = j.From.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_To:

	/* handler: j.To type=api.ContactDto kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		err = j.To.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_IsOutstanding:

	/* handler: j.IsOutstanding type=bool kind=bool quoted=false*/

	{
		if tok != fflib.FFTok_bool && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for bool", tok))
		}
	}

	{
		if tok == fflib.FFTok_null {

		} else {
			tmpb := fs.Output.Bytes()

			if bytes.Compare([]byte{'t', 'r', 'u', 'e'}, tmpb) == 0 {

				j.IsOutstanding = true

			} else if bytes.Compare([]byte{'f', 'a', 'l', 's', 'e'}, tmpb) == 0 {

				j.IsOutstanding = false

			} else {
				err = errors.New("unexpected bytes for true/false value")
				return fs.WrapErr(err)
			}

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Creator:

	/* handler: j.Creator type=api.ApiUserDto kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		err = j.Creator.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_CreatorComment:

	/* handler: j.CreatorComment type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.CreatorComment = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Acknowledge:

	/* handler: j.Acknowledge type=api.ApiAcknowledgeDto kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			j.Acknowledge = nil

			state = fflib.FFParse_after_value
			goto mainparse
		}

		if j.Acknowledge == nil {
			j.Acknowledge = new(ApiAcknowledgeDto)
		}

		err = j.Acknowledge.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
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

// MarshalJSON marshal bytes to json - template
func (j *ApiUserDto) MarshalJSON() ([]byte, error) {
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
func (j *ApiUserDto) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{ "Id":`)
	fflib.FormatBits2(buf, uint64(j.ID), 10, j.ID < 0)
	buf.WriteByte(',')
	if len(j.Name) != 0 {
		buf.WriteString(`"Name":`)
		fflib.WriteJsonString(buf, string(j.Name))
		buf.WriteByte(',')
	}
	buf.Rewind(1)
	buf.WriteByte('}')
	return nil
}

const (
	ffjtApiUserDtobase = iota
	ffjtApiUserDtonosuchkey

	ffjtApiUserDtoID

	ffjtApiUserDtoName
)

var ffjKeyApiUserDtoID = []byte("Id")

var ffjKeyApiUserDtoName = []byte("Name")

// UnmarshalJSON umarshall json - template of ffjson
func (j *ApiUserDto) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *ApiUserDto) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtApiUserDtobase
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
				currentKey = ffjtApiUserDtonosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'I':

					if bytes.Equal(ffjKeyApiUserDtoID, kn) {
						currentKey = ffjtApiUserDtoID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'N':

					if bytes.Equal(ffjKeyApiUserDtoName, kn) {
						currentKey = ffjtApiUserDtoName
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiUserDtoName, kn) {
					currentKey = ffjtApiUserDtoName
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyApiUserDtoID, kn) {
					currentKey = ffjtApiUserDtoID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtApiUserDtonosuchkey
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

				case ffjtApiUserDtoID:
					goto handle_ID

				case ffjtApiUserDtoName:
					goto handle_Name

				case ffjtApiUserDtonosuchkey:
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

handle_ID:

	/* handler: j.ID type=int64 kind=int64 quoted=false*/

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

			j.ID = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Name:

	/* handler: j.Name type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Name = string(string(outBuf))

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
