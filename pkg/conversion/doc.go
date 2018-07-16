// Package conversion provides go object versioning and encoding/decoding
// mechanisms.
//
// Specifically, conversion provides a way for you to define multiple versions
// o the same object. You may write functions which implement conversion logic,
// but for the fields which did not change, copying is automated. This makes it
// easy to medify the structures you use in memory without affecting the format
// you store on disk or respond to in your external API calls.
//
// The send offering of this package is automated encoding/decodeing. The version
// and type of the object is recorded in the output, so it can be recreated upon
// reading. Currently, conversion writes JSON output, and interprets both JSON
// and YAML input.
//
// In the future, we plan to more explicitly separate the above two mechanisms, and
// add more serialization options, such as gob.
package conversion
