package envsnapshot

import "encoding/json"

func project(m Manifest) ([]byte, error) { return json.Marshal(m) }
