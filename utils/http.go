package utils

import (
	"fmt"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/mitchellh/mapstructure"

	"github.com/golang/protobuf/proto"
)

var (
	unmarshalOptions = &jsonpb.Unmarshaler{}
	marshalOptions   = &jsonpb.Marshaler{EnumsAsInts: true, EmitDefaults: true, OrigName: true}
)

func ReadHTTPReq(req *http.Request, m proto.Message) error {
	switch req.Method {
	case "POST":
		if err := unmarshalOptions.Unmarshal(req.Body, m); err != nil {
			fmt.Printf("e is %v", err)
			return err
		}
	case "GET":
		form := req.URL.Query()
		values := make(map[string]string, len(form))
		for key := range form {
			values[key] = form.Get(key)
		}
		if err := decodeMapStructure(values, &m); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func WriteHTTPJSONRes(w http.ResponseWriter, pb proto.Message) error {
	w.Header().Set("Content-Type", "application/json")
	if err := marshalOptions.Marshal(w, pb); err != nil {
		return err
	}
	return nil
}

func decodeMapStructure(valuesMap map[string]string, i interface{}) error {
	decodeConfig := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           i,
		TagName:          "json",
	}
	decoder, err := mapstructure.NewDecoder(decodeConfig)
	if err != nil {
		return err
	}
	err = decoder.Decode(valuesMap)
	if err != nil {
		return err
	}
	return nil
}
