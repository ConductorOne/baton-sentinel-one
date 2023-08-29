// Copyright 2022 The Sqlite Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite3

const (
	SQLITE_STATIC    = uintptr(0)  // ((sqlite3_destructor_type)0)
	SQLITE_TRANSIENT = ^uintptr(0) // ((sqlite3_destructor_type)-1)
)
