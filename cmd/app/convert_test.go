package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestConvertService_Convert(t *testing.T) {
	type fields struct {
		Getter func(url string) (resp *http.Response, err error)
	}
	type args struct {
		from       string
		to         string
		fromAmount float64
	}

	type test struct {
		name   string
		fields fields
		args   args
		want   ConversionResponse
		// TODO assert on the actual error message
		wantErr bool
	}

	// TODO there are more error conditions
	tests := []test{
		func() test {
			const name = "given upstream error, expect error"

			return test{
				name: name,
				fields: fields{
					Getter: func(url string) (resp *http.Response, err error) {
						return nil, errors.New("upstream error")
					},
				},
				args: args{
					from:       "abc",
					to:         "def",
					fromAmount: 1,
				},
				wantErr: true,
			}
		}(),
		func() test {
			const name = "given non OK status code response, expect error"

			return test{
				name: name,
				fields: fields{
					Getter: func(url string) (resp *http.Response, err error) {
						return &http.Response{
							StatusCode: http.StatusTeapot,
						}, nil
					},
				},
				args: args{
					from:       "abc",
					to:         "def",
					fromAmount: 1,
				},
				wantErr: true,
			}
		}(),
		func() test {
			const name = "given invalid json response, expect error"

			return test{
				name: name,
				fields: fields{
					Getter: func(url string) (resp *http.Response, err error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("unquoted")),
						}, nil
					},
				},
				args: args{
					from:       "abc",
					to:         "def",
					fromAmount: 1,
				},
				wantErr: true,
			}
		}(),
		func() test {
			const name = "given json response, but unknown value, expect error"

			return test{
				name: name,
				fields: fields{
					Getter: func(url string) (resp *http.Response, err error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`"quoted"`)),
						}, nil
					},
				},
				args: args{
					from:       "abc",
					to:         "def",
					fromAmount: 1,
				},
				wantErr: true,
			}
		}(),
		func() test {
			const name = "given valid json response, expect response"

			expectedDate := time.Now()
			expectedRate := 123.4

			want := ConversionResponse{
				Date:       expectedDate.String(),
				Rate:       expectedRate,
				From:       "abc",
				FromAmount: 2,
				To:         "def",
				ToAmount:   expectedRate * 2,
			}

			return test{
				name: name,
				fields: fields{
					Getter: func(url string) (resp *http.Response, err error) {
						body := map[string]interface{}{
							"date":  expectedDate.String(),
							want.To: expectedRate,
						}

						bodyBytes, err := json.Marshal(body)
						if err != nil {
							panic(err)
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
						}, nil
					},
				},
				args: args{
					from:       want.From,
					to:         want.To,
					fromAmount: want.FromAmount,
				},
				want: want,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ConvertService{
				getter: tt.fields.Getter,
			}
			got, err := s.Convert(tt.args.from, tt.args.to, tt.args.fromAmount)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, got, tt.want)
		})
	}
}
