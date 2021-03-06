// Copyright 2014, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package lzma

import "io"
import "bitbucket.org/rawr/golib/errs"
import "bitbucket.org/rawr/golib/ioutil"
import "bitbucket.org/rawr/goxz/lib"

type reader struct {
	*lib.Stream
	rd     io.Reader
	closed bool
}

func NewReader(rd io.Reader) (io.ReadCloser, error) {
	stream, err := lib.NewStreamLZMAAloneDecoder(maxMemory)
	if err != nil {
		return nil, err
	}
	return &reader{stream, rd, false}, nil
}

func (r *reader) Read(data []byte) (cnt int, err error) {
	if r.closed {
		return 0, io.ErrClosedPipe
	}

	rdCnt, _, err := r.Process(lib.RUN, ioutil.NewWriter(data), r.rd)
	err = errs.Ignore(err, io.ErrShortWrite)
	err = errs.Convert(err, io.ErrUnexpectedEOF, io.EOF)
	err = errs.Convert(err, io.EOF, streamEnd)
	if err == io.EOF { // Ensure all input consumed
		if _, inBuf := r.GetBuffers(); inBuf.Length() != 0 {
			return int(rdCnt), dataError
		}
		if _, err := r.rd.Read([]byte{0}); err != io.EOF {
			return int(rdCnt), dataError
		}
	}
	return int(rdCnt), err
}

func (r *reader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true

	defer r.End()
	return nil
}
