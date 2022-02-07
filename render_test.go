package graphiteapi

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	repeated               = 2
	flt32                  = 5
	protobufMaxVarintBytes = 10 // maximum length of a varint
)

var pbVarints []byte

func VarintEncode(x uint64) []byte {
	var buf [protobufMaxVarintBytes]byte
	var n int
	for n = 0; x > 127; n++ {
		buf[n] = 0x80 | uint8(x&0x7F)
		x >>= 7
	}
	buf[n] = uint8(x)
	n++
	return buf[0:n]
}

func Fixed64Encode(x uint64) []byte {
	return []byte{
		uint8(x),
		uint8(x >> 8),
		uint8(x >> 16),
		uint8(x >> 24),
		uint8(x >> 32),
		uint8(x >> 40),
		uint8(x >> 48),
		uint8(x >> 56),
	}
}

func Fixed32Encode(x uint32) []byte {
	return []byte{
		uint8(x),
		uint8(x >> 8),
		uint8(x >> 16),
		uint8(x >> 24),
	}
}

func VarintWrite(w io.Writer, x uint64) {
	// for ResponseWriter. ignore write result
	if x < 128 {
		w.Write(pbVarints[x : x+1])
	} else if x < 16384 {
		w.Write(pbVarints[x*2-128 : x*2-126])
	} else {
		w.Write(VarintEncode(x))
	}
}

func ProtobufWriteSingle(w io.Writer, value float32) {
	w.Write(Fixed32Encode(math.Float32bits(value)))
}

func ProtobufWriteDouble(w io.Writer, value float64) {
	w.Write(Fixed64Encode(math.Float64bits(value)))
}

func initV2PB() {
	// precalculate varints
	buf := bytes.NewBuffer(nil)

	for i := uint64(0); i < 16384; i++ {
		buf.Write(VarintEncode(i))
	}

	pbVarints = buf.Bytes()
}

type V2PB struct {
	b1 *bytes.Buffer
	b2 *bytes.Buffer
}

func (v *V2PB) initBuffer() {
	v.b1 = new(bytes.Buffer)
	v.b2 = new(bytes.Buffer)
}

func (v *V2PB) writeBody(writer *bufio.Writer, name string, from, until, step uint32, points []float64) {
	v.b1.Reset()
	v.b2.Reset()

	// name
	VarintWrite(v.b1, (1<<3)+repeated) // tag
	VarintWrite(v.b1, uint64(len(name)))
	v.b1.WriteString(name)

	// start
	VarintWrite(v.b1, 2<<3)
	VarintWrite(v.b1, uint64(from))

	// stop
	VarintWrite(v.b1, 3<<3)
	VarintWrite(v.b1, uint64(until))

	// step
	VarintWrite(v.b1, 4<<3)
	VarintWrite(v.b1, uint64(step))

	// start write to output
	// Write values
	VarintWrite(v.b1, (5<<3)+repeated)
	VarintWrite(v.b1, uint64(8*len(points)))

	// Write isAbsent
	VarintWrite(v.b2, (6<<3)+repeated)
	VarintWrite(v.b2, uint64(len(points)))

	for _, value := range points {
		if math.IsNaN(value) {
			ProtobufWriteDouble(v.b1, 0)
			v.b2.WriteByte(1)
		} else {
			ProtobufWriteDouble(v.b1, value)
			v.b2.WriteByte(0)
		}
	}

	// repeated FetchResponse metrics = 1;
	// write tag and len
	VarintWrite(writer, (1<<3)+repeated)
	VarintWrite(writer, uint64(v.b1.Len())+uint64(v.b2.Len()))

	writer.Write(v.b1.Bytes())
	writer.Write(v.b2.Bytes())
}

func renderTest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse form error", http.StatusBadRequest)
		return
	}
	targets := r.Form.Get("target")
	from := r.Form.Get("from")
	until := r.Form.Get("until")
	format := r.Form.Get("format")

	if format == "protobuf" {
		pb := V2PB{}
		pb.initBuffer()
		writer := bufio.NewWriterSize(w, 1024*1024)
		defer writer.Flush()

		if from == "-5m" || until == "now" {
			if targets == "TEST.*" {
				pb.writeBody(writer, "TEST.1", 1643964180, 1643964240, 60, []float64{10.0, 5.0})
				pb.writeBody(writer, "TEST.2", 1643964180, 1643964240, 60, []float64{1.0, 2.0})
			}
		}
	} else {
		http.Error(w, "invalid format", http.StatusBadRequest)
	}
}

func TestRenderQuery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/render/", renderTest)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	initV2PB()

	base := "http://" + ts.Listener.Addr().String()

	tests := []struct {
		name    string
		targets []string
		from    string
		until   string
		want    []Series
	}{
		{
			name:    "TEST.*",
			targets: []string{"TEST.*"},
			from:    "-1m",
			until:   "now",
			want: []Series{
				{
					Target:    "TEST.1",
					StartTime: 1643964180,
					StopTime:  1643964240,
					StepTime:  60,
					Values:    []float64{10.0, 5.0},
				},
				{
					Target:    "TEST.2",
					StartTime: 1643964180,
					StopTime:  1643964240,
					StepTime:  60,
					Values:    []float64{1.0, 2.0},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			render := NewRenderQuery(base, tt.from, tt.until, tt.targets)
			if gotSeries, err := render.Request(ctx); err != nil {
				t.Errorf("NewRenderQuery().Request(ctx) got err = %v", err)
			} else if len(tt.want) != 0 && len(tt.want) != len(gotSeries) {
				if !reflect.DeepEqual(gotSeries, tt.want) {
					t.Errorf("NewRenderQuery().Request(ctx) = %v, want %v", gotSeries, tt.want)
				}
			}
		})
	}
}
