package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func respondJSONError(w http.ResponseWriter, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
}

func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	// If the payload is a Protobuf message, use protojson
	if msg, ok := data.(proto.Message); ok {
		marshaler := protojson.MarshalOptions{
			UseProtoNames:   true,  // Keeps snake_case field names in JSON
			EmitUnpopulated: false, // Omit empty/zero fields
		}

		jsonBytes, err := marshaler.Marshal(msg)

		if err != nil {
			respondJSONError(w, "failed to encode proto response", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonBytes)

		if err != nil {
			respondJSONError(w, "failed to write proto response", http.StatusInternalServerError)
		}

		return
	}

	// Fallback for standard Go structs / maps
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func decodeProtoJSON(r io.Reader, msg proto.Message) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true, // Ignore extra fields sent by client
	}
	return unmarshaler.Unmarshal(body, msg)
}
