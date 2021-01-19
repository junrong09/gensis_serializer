package serializer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
)

/*
Serialized Format:
<method>\n<body_length>\n<strings_size>\n<numbers_size>\n<binaries_size>\n
<body>

Legend:
<body> = <value_1>\r\n[<value_x>\r\n]...
*/
type Gensis struct {
	Method   string
	length   int
	Strings  []string
	Numbers  []int64
	Binaries [][]byte
}

const (
	// Server Response Methods
	METHOD_ERROR    = "error"
	METHOD_RESPONSE = "response"

	// Client Request Methods
	METHOD_AUTH           = "auth"    // 2 x string; 1 x string
	METHOD_UPDATE_PROFILE = "update"  // 1 x string or/and 1 x binary; EMPTY
	METHOD_GET_PROFILE    = "profile" // EMPTY;
)

func (gen *Gensis) computeLength() {
	length := len(gen.Strings) + len(gen.Numbers) + len(gen.Binaries)
	length *= 2
	for _, s := range gen.Strings {
		length += len(s)
	}
	length += 8 * len(gen.Numbers)
	for _, b := range gen.Binaries {
		length += len(b)
	}
	gen.length = length
}

func (gen *Gensis) Encode() *bytes.Buffer {
	gen.computeLength()

	header := fmt.Sprintf("%s\n%d\n%d\n%d\n%d\n", gen.Method, gen.length, len(gen.Strings), len(gen.Numbers), len(gen.Binaries))
	load := new(bytes.Buffer)
	load.Write([]byte(header))
	for _, s := range gen.Strings {
		load.WriteString(s)
		load.Write([]byte{'\r', '\n'})
	}
	for _, n := range gen.Numbers {
		err := binary.Write(load, binary.BigEndian, n)
		load.Write([]byte{'\r', '\n'})
		if err != nil {
			log.Println(err)
		}
	}
	for _, b := range gen.Binaries {
		load.Write(b)
		load.Write([]byte{'\r', '\n'})
	}
	return load
}

func Decoder(reader io.Reader) (*Gensis, error) {
	bReader := bufio.NewReader(reader)
	gen := new(Gensis)

	if b, err := bReader.ReadBytes('\n'); err != nil {
		return nil, err
	} else {
		gen.Method = string(b[:len(b)-1])
	}

	lengths := make([]int, 4)
	for i, _ := range lengths {
		if b, err := bReader.ReadBytes('\n'); err != nil {
			return nil, err
		} else {
			if lengths[i], err = strconv.Atoi(string(b[:len(b)-1])); err != nil {
				return nil, err
			}
		}
	}
	gen.length = lengths[0]

	valuesBytes := make([]byte, gen.length)

	if _, err := io.ReadFull(bReader, valuesBytes); err != nil {
		return nil, err
	}
	if err := decodeBody(valuesBytes, lengths, gen); err != nil {
		return nil, err
	}
	return gen, nil
}

func decodeBody(valuesBytes []byte, lengths []int, gen *Gensis) error {
	values := bytes.Split(valuesBytes, []byte{'\r', '\n'})
	if len(values) > 0 {
		values = values[:len(values)-1]
	}
	if valuesSize := lengths[1] + lengths[2] + lengths[3]; len(values) != valuesSize {
		return errors.New(fmt.Sprintf("Mismatch size of values. Required %d, found %d", valuesSize, len(values)))
	}
	strings, numbers, binaries := make([]string, 0), make([]int64, 0), make([][]byte, 0)
	valuesPos := 0
	for i := 0; i < lengths[1]; i++ {
		strings = append(strings, string(values[valuesPos]))
		valuesPos++
	}
	for i := 0; i < lengths[2]; i++ {
		var n int64
		err := binary.Read(bytes.NewBuffer(values[valuesPos]), binary.BigEndian, &n)
		valuesPos++
		if err != nil {
			return err
		}
		numbers = append(numbers, n)
	}
	for i := 0; i < lengths[3]; i++ {
		binaries = append(binaries, values[valuesPos])
		valuesPos++
	}
	gen.Strings, gen.Numbers, gen.Binaries = strings, numbers, binaries
	return nil
}
