// Package runtime includes helper functions for working with API objects
// That follow the kubernetes API objecct conventions, which are:
//
// 0. Your API objects have a common metadata struct member, JSONBase.
// 1. Your code refers to an internal or API objects.
// 2. In a separate package, you have an external set of API objects.
// 3. The external set is considered to be versioned, and no breaking
//    changes are ever made to it (fields may be added but not changed
//    or removed).
// 4. As your api evelves, you'll make an additional versioned package
//    with every major change.
// 5. Versioned packages have conversion functions which convert to
//    and from the internal version.
// 6. You'll continue to support older versions according to your
//    deprecation policy, and you can easily provide a program/library
//    to update old versions into new versions because of 5.
// 7. All of yoru serializations and deserializations are handled in a
//    centralized place.
//
// Package runtime provides a conversion helper to make 5 easy, and the
// Encode/Decode/DecodeInto trio to accomplish 7. You can also register
// additional "codecs" which use a version of your choice. It's
// recommanded that you register your types with runtime in your
// package's init function.
//
// As a bouns, a few common types useful from all api objects and versions
// are provided in types.go.
package runtime
